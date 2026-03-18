package bot

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-bot/internal/api"
	"telegram-bot/internal/bot/keyboards"
	"telegram-bot/internal/bot/states"
	"telegram-bot/internal/cache"
	"telegram-bot/internal/config"
	"telegram-bot/internal/database"
	"telegram-bot/internal/logger"
	"telegram-bot/internal/scheduler"
)

type Bot struct {
	API          *tgbotapi.BotAPI
	Config       *config.Config
	APIClient    *api.APIClient
	StateManager *states.StateManager
	DB           *database.DB
	Scheduler    *scheduler.Scheduler
	Cache        *cache.Cache
	Server       *http.Server
}

func NewBot(cfg *config.Config) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	db, err := database.NewDB(cfg.DatabaseURL)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to connect to database")
	}

	c := cache.NewCache(cfg)
	if err := c.Ping(context.Background()); err != nil {
		logger.Warn().Err(err).Msg("failed to connect to redis")
		c = nil
	}

	apiClient := api.NewAPIClient(cfg)
	sched := scheduler.NewScheduler(cfg, apiClient)
	if cfg.SchedulerEnabled {
		sched.Start()
	}

	return &Bot{
		API:          botAPI,
		Config:       cfg,
		APIClient:    apiClient,
		StateManager: states.NewStateManager(),
		DB:           db,
		Scheduler:    sched,
		Cache:        c,
	}, nil
}

func (b *Bot) Start() error {
	logger.Info().Str("username", b.API.Self.UserName).Msg("bot started")

	if b.Config.Mode == "webhook" {
		return b.startWebhook()
	}

	return b.startPolling()
}

func (b *Bot) startPolling() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.API.GetUpdatesChan(u)

	for update := range updates {
		b.handleUpdate(update)
	}

	return nil
}

func (b *Bot) startWebhook() error {
	wh, err := tgbotapi.NewWebhook(b.Config.WebhookURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook config: %w", err)
	}

	_, err = b.API.Request(wh)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	info, err := b.API.GetWebhookInfo()
	if err != nil {
		return fmt.Errorf("failed to get webhook info: %w", err)
	}

	logger.Info().Interface("info", info).Msg("webhook info")

	mux := http.NewServeMux()
	mux.HandleFunc("/"+b.Config.WebhookSecret, b.webhookHandler)
	mux.HandleFunc("/health", b.healthHandler)

	b.Server = &http.Server{
		Addr:    b.Config.WebhookListen,
		Handler: mux,
	}

	go func() {
		logger.Info().Str("addr", b.Config.WebhookListen).Msg("starting webhook server")
		if err := b.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("webhook server error")
		}
	}()

	info, _ = b.API.GetWebhookInfo()
	for info.LastErrorDate != 0 {
		time.Sleep(5 * time.Second)
		info, _ = b.API.GetWebhookInfo()
	}

	return nil
}

func (b *Bot) webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error().Err(err).Msg("error reading body")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var update tgbotapi.Update
	if err := json.Unmarshal(body, &update); err != nil {
		logger.Error().Err(err).Msg("error unmarshaling update")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	go b.handleUpdate(update)

	w.WriteHeader(http.StatusOK)
}

func (b *Bot) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		b.handleMessage(update.Message)
	} else if update.CallbackQuery != nil {
		b.handleCallback(update.CallbackQuery)
	}
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	userID := msg.From.ID
	text := msg.Text

	if b.Cache != nil {
		if err := b.Cache.SetRateLimit(userID, "messages", 20, time.Minute); err != nil {
			b.sendWithKeyboard(userID, "⏳ Слишком много сообщений. Подождите минуту.", keyboards.BackToMain())
			return
		}
	}

	state := b.getUserState(userID)

	if text == "/start" {
		b.handleStart(msg)
		return
	}

	if text == "/menu" {
		b.showMainMenu(msg)
		return
	}

	if text == "/help" {
		b.handleHelp(msg)
		return
	}

	if text == "/resumes" {
		b.handleResumesMenu(msg)
		return
	}

	if text == "/vacancies" {
		b.handleVacanciesMenu(msg)
		return
	}

	if text == "/schedule" {
		b.handleScheduleMenu(msg)
		return
	}

	if text == "/stats" {
		b.handleStats(msg)
		return
	}

	if text == "/payment" {
		b.handlePaymentMenu(msg)
		return
	}

	switch state.State {
	case states.StateResumeTitle:
		state.ResumeTitle = text
		state.State = states.StateResumeContent
		b.setUserState(userID, state)
		b.sendMessage(userID, "Введите текст резюме")

	case states.StateResumeContent:
		state.ResumeContent = text
		b.setUserState(userID, state)
		b.createResume(userID, state.ResumeTitle, state.ResumeContent)
		b.clearUserState(userID)

	case states.StateEditResumeTitle:
		state.ResumeTitle = text
		state.State = states.StateEditResumeContent
		b.setUserState(userID, state)
		b.sendMessage(userID, "Введите новый текст резюме")

	case states.StateEditResumeContent:
		state.ResumeContent = text
		b.setUserState(userID, state)
		b.updateResume(userID, state.EditingResumeID, state.ResumeTitle, state.ResumeContent)
		b.clearUserState(userID)

	case states.StateSearchQuery:
		state.SearchQuery = text
		b.setUserState(userID, state)
		b.startVacancySearch(userID, state.SelectedResumeID, state.SearchQuery, nil)
		b.clearUserState(userID)

	default:
		b.sendMessage(userID, "Используйте /menu для навигации")
	}
}

func (b *Bot) getUserState(userID int64) *states.UserStateData {
	if b.Cache != nil {
		state, err := b.Cache.GetUserState(userID)
		if err == nil && state != nil {
			return state
		}
	}
	return b.StateManager.GetState(userID)
}

func (b *Bot) setUserState(userID int64, state *states.UserStateData) {
	if b.Cache != nil {
		b.Cache.SetUserState(userID, state, 24*time.Hour)
	}
	b.StateManager.SetStateData(userID, state)
}

func (b *Bot) clearUserState(userID int64) {
	if b.Cache != nil {
		b.Cache.DeleteUserState(userID)
	}
	b.StateManager.ClearState(userID)
}

func (b *Bot) handleCallback(callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	data := callback.Data

	b.API.Request(tgbotapi.CallbackConfig{
		CallbackQueryID: callback.ID,
	})

	switch data {
	case "menu_resumes":
		b.handleResumesMenuByID(userID)
	case "menu_vacancies":
		b.handleVacanciesMenuByID(userID)
	case "menu_payment":
		b.handlePaymentMenuByID(userID)
	case "menu_stats":
		b.handleStatsByID(userID)
	case "menu_settings":
		b.handleSettingsMenuByID(userID)
	case "back_main":
		b.showMainMenuByID(userID)

	case "resume_create":
		b.startCreateResume(userID)
	case "resume_list":
		b.showResumeList(userID)

	case "vacancy_start":
		b.startVacancySearchFlow(userID)
	case "vacancy_schedule":
		b.handleScheduleMenuByID(userID)
	case "vacancy_history":
		b.showVacancyHistory(userID)

	case "payment_status":
		b.showPaymentStatus(userID)
	case "payment_pay":
		b.showPaymentTariffs(userID)
	case "payment_1month":
		b.processPayment(userID, 30)
	case "payment_3month":
		b.processPayment(userID, 90)
	case "payment_12month":
		b.processPayment(userID, 365)
	case "payment_confirmed":
		b.handlePaymentConfirmed(userID)

	case "settings_notifications":
		b.toggleNotifications(userID)

	case "filter_skip":
		state := b.getUserState(userID)
		var filters *api.VacancyFilters
		if len(state.EmploymentFilter) > 0 || len(state.WorkFormatFilter) > 0 {
			filters = &api.VacancyFilters{
				Employment: state.EmploymentFilter,
				Format:     state.WorkFormatFilter,
			}
		}
		b.startVacancySearch(userID, state.SelectedResumeID, state.SearchQuery, filters)
		b.clearUserState(userID)

	case "filter_employment":
		b.handleEmploymentFilter(userID)
	case "filter_selfemployed":
		b.handleSelfEmployedFilter(userID)
	case "filter_ip":
		b.handleIPFilter(userID)
	case "filter_remote":
		b.handleRemoteFilter(userID)
	case "filter_office":
		b.handleOfficeFilter(userID)
	case "filter_hybrid":
		b.handleHybridFilter(userID)

	case "schedule_time":
		b.showScheduleTimeOptions(userID)
	case "schedule_days":
		b.showScheduleDaysOptions(userID)
	case "schedule_query":
		b.askScheduleQuery(userID)
	case "schedule_disable":
		b.disableSchedule(userID)

	default:
		b.handleCallbackData(callback)
	}
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	userID := msg.From.ID
	username := msg.From.UserName

	if b.DB != nil {
		b.DB.CreateUser(userID, username)
	}

	subscribed, _, _ := b.APIClient.IsSubscribed(userID)
	if !subscribed {
		b.sendMessage(userID, "Добро пожаловать! Для использования бота нужна подписка.")
		b.showPaymentMenuByID(userID)
		return
	}

	b.sendMessage(userID, fmt.Sprintf("Добро пожаловать, %s!", msg.From.FirstName))
	b.showMainMenu(msg)
}

func (b *Bot) handleHelp(msg *tgbotapi.Message) {
	helpText := `
📖 Список команд:
/start - Запуск бота
/menu - Главное меню
/resumes - Управление резюме
/vacancies - Поиск вакансий
/schedule - Настройка расписания
/stats - Ваша статистика
/payment - Управление подпиской
/help - Помощь
`
	b.sendMessage(msg.Chat.ID, helpText)
}

func (b *Bot) sendMessage(chatID int64, text string) {
	b.API.Send(tgbotapi.NewMessage(chatID, text))
}

func (b *Bot) sendWithKeyboard(chatID int64, text string, keyboard interface{}) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.API.Send(msg)
}

func (b *Bot) showMainMenu(msg *tgbotapi.Message) {
	b.sendWithKeyboard(msg.Chat.ID, "📋 Меню", keyboards.MainMenu())
}

func (b *Bot) showMainMenuByID(userID int64) {
	b.sendWithKeyboard(userID, "📋 Меню", keyboards.MainMenu())
}

func (b *Bot) handleResumesMenu(msg *tgbotapi.Message) {
	b.sendWithKeyboard(msg.Chat.ID, "📄 Мои резюме", keyboards.ResumesMenu())
}

func (b *Bot) handleResumesMenuByID(userID int64) {
	b.sendWithKeyboard(userID, "📄 Мои резюме", keyboards.ResumesMenu())
}

func (b *Bot) handleVacanciesMenu(msg *tgbotapi.Message) {
	b.sendWithKeyboard(msg.Chat.ID, "🔍 Поиск вакансий", keyboards.VacanciesMenu())
}

func (b *Bot) handleVacanciesMenuByID(userID int64) {
	b.sendWithKeyboard(userID, "🔍 Поиск вакансий", keyboards.VacanciesMenu())
}

func (b *Bot) handlePaymentMenu(msg *tgbotapi.Message) {
	b.sendWithKeyboard(msg.Chat.ID, "💳 Оплата", keyboards.PaymentMenu())
}

func (b *Bot) handlePaymentMenuByID(userID int64) {
	b.sendWithKeyboard(userID, "💳 Оплата", keyboards.PaymentMenu())
}

func (b *Bot) showPaymentMenuByID(userID int64) {
	b.sendWithKeyboard(userID, "💳 Для использования бота необходима подписка", keyboards.PaymentMenu())
}

func (b *Bot) handleStats(msg *tgbotapi.Message) {
	b.sendUserStats(msg.Chat.ID, msg.From.ID)
}

func (b *Bot) handleStatsByID(userID int64) {
	b.sendUserStats(userID, userID)
}

func (b *Bot) sendUserStats(chatID int64, userID int64) {
	stats, err := b.APIClient.GetUserStats(userID)
	if err != nil {
		b.sendMessage(chatID, "Не удалось получить статистику")
		return
	}

	text := fmt.Sprintf(`📈 Ваша статистика:
Резюме: %d
Проведено поисков: %d
Найдено вакансий: %d
Просмотрено: %d
Откликов: %d
📊 За этот месяц:
├── Резюме: %d
├── Поисков: %d
└── Вакансий: %d`, stats.ResumesCount, stats.SearchesCount, stats.VacanciesFound, stats.VacanciesViewed, stats.ResponsesCount, stats.MonthResumes, stats.MonthSearches, stats.MonthVacancies)

	b.sendMessage(chatID, text)
}

func (b *Bot) handleSettingsMenuByID(userID int64) {
	b.sendWithKeyboard(userID, "⚙️ Настройки", keyboards.SettingsMenu())
}

func (b *Bot) startCreateResume(userID int64) {
	state := b.getUserState(userID)
	state.State = states.StateResumeTitle
	b.setUserState(userID, state)
	b.sendMessage(userID, "📝 Введите название резюме")
}

func (b *Bot) createResume(userID int64, title, content string) {
	resume, err := b.APIClient.CreateResume(userID, title, content)
	if err != nil {
		b.sendMessage(userID, "❌ Ошибка создания резюме: "+err.Error())
		return
	}

	if b.Cache != nil {
		b.Cache.InvalidateResumesCache(userID)
	}

	b.sendMessage(userID, fmt.Sprintf("✅ Резюме '%s' создано! Отправлено на анализ AI.", resume.Title))
	b.showResumeList(userID)
}

func (b *Bot) showResumeList(userID int64) {
	if b.Cache != nil {
		if data, err := b.Cache.GetCachedResumes(userID); err == nil && data != nil {
			b.sendWithKeyboard(userID, string(data), keyboards.ResumesMenu())
			return
		}
	}

	resumes, err := b.APIClient.GetResumes(userID)
	if err != nil {
		b.sendMessage(userID, "❌ Не удалось получить список резюме")
		return
	}

	if len(resumes) == 0 {
		b.sendWithKeyboard(userID, "У вас пока нет резюме. Создайте первое!", keyboards.ResumesMenu())
		return
	}

	text := "📄 Ваши резюме:\n"
	for i, r := range resumes {
		status := "⏳ Анализ..."
		if r.Analyzed {
			status = fmt.Sprintf("✅ AI: %.0f%%", r.AIScore*100)
		}
		text += fmt.Sprintf("%d. %s (%s)\n", i+1, r.Title, status)
	}

	b.sendWithKeyboard(userID, text, keyboards.ResumeList(resumes))
}

func (b *Bot) updateResume(userID int64, resumeID int64, title, content string) {
	resume, err := b.APIClient.UpdateResume(resumeID, title, content)
	if err != nil {
		b.sendMessage(userID, "❌ Ошибка обновления резюме: "+err.Error())
		return
	}

	if b.Cache != nil {
		b.Cache.InvalidateResumesCache(userID)
	}

	b.sendMessage(userID, fmt.Sprintf("✅ Резюме '%s' обновлено!", resume.Title))
}

func (b *Bot) startVacancySearchFlow(userID int64) {
	resumes, err := b.APIClient.GetResumes(userID)
	if err != nil {
		b.sendMessage(userID, "❌ Не удалось получить список резюме")
		return
	}

	if len(resumes) == 0 {
		b.sendWithKeyboard(userID, "Сначала создайте резюме!", keyboards.ResumesMenu())
		return
	}

	state := b.getUserState(userID)
	state.State = states.StateSelectResume
	b.setUserState(userID, state)

	b.sendMessage(userID, "Выберите резюме для сравнения:\n"+b.formatResumeList(resumes))
}

func (b *Bot) formatResumeList(resumes []api.Resume) string {
	var text string
	for i, r := range resumes {
		text += fmt.Sprintf("%d. %s\n", i+1, r.Title)
	}
	return text
}

func (b *Bot) startVacancySearch(userID int64, resumeID int64, query string, filters *api.VacancyFilters) {
	b.sendWithKeyboard(userID, "🔍 Ищу вакансии...", keyboards.CancelSearch())

	job, err := b.APIClient.MatchVacancies(userID, &api.MatchRequest{
		ResumeID: resumeID,
		Query:    query,
		Filters:  filters,
	})
	if err != nil {
		b.sendMessage(userID, "❌ Ошибка поиска: "+err.Error())
		return
	}

	go b.waitForResults(userID, job.ID)
}

func (b *Bot) waitForResults(userID int64, jobID int64) {
	for i := 0; i < 30; i++ {
		result, err := b.APIClient.GetMatchResults(jobID)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		if result != nil && len(result.Vacancies) > 0 {
			b.sendVacancyResults(userID, result.Vacancies)
			return
		}
		time.Sleep(2 * time.Second)
	}
	b.sendMessage(userID, "⏳ Поиск仍在进行中...")
}

func (b *Bot) sendVacancyResults(userID int64, vacancies []api.Vacancy) {
	text := fmt.Sprintf("🎯 Найдено %d подходящих вакансий:\n\n", len(vacancies))

	for i, v := range vacancies {
		if i >= 5 {
			break
		}
		emoji := ""
		switch i {
		case 0:
			emoji = "1️⃣"
		case 1:
			emoji = "2️⃣"
		case 2:
			emoji = "3️⃣"
		case 3:
			emoji = "4️⃣"
		case 4:
			emoji = "5️⃣"
		}
		text += fmt.Sprintf("%s %s — %s\n", emoji, v.Title, v.Company)
		text += fmt.Sprintf("   📊 Совпадение: %.0f%%\n", v.MatchScore*100)
		text += fmt.Sprintf("   💰 %s\n", v.Salary)
		text += fmt.Sprintf("   📍 %s\n", v.Location)
		text += fmt.Sprintf("   🔗 %s\n\n", v.URL)
	}

	b.sendWithKeyboard(userID, text, keyboards.SearchResults(vacancies))
}

func (b *Bot) showVacancyHistory(userID int64) {
	b.sendMessage(userID, "📊 История поиска пока пуста")
}

func (b *Bot) handleScheduleMenu(msg *tgbotapi.Message) {
	b.sendWithKeyboard(msg.Chat.ID, "📅 Настройка расписания поиска", keyboards.ScheduleMenu())
}

func (b *Bot) handleScheduleMenuByID(userID int64) {
	b.sendWithKeyboard(userID, "📅 Настройка расписания поиска", keyboards.ScheduleMenu())
}

func (b *Bot) showScheduleTimeOptions(userID int64) {
	b.sendWithKeyboard(userID, "🕐 Выберите время:", keyboards.ScheduleTimeOptions())
}

func (b *Bot) showScheduleDaysOptions(userID int64) {
	b.sendWithKeyboard(userID, "📅 Выберите дни:", keyboards.ScheduleDaysOptions())
}

func (b *Bot) askScheduleQuery(userID int64) {
	state := b.getUserState(userID)
	state.State = states.StateScheduleQuery
	b.setUserState(userID, state)
	b.sendMessage(userID, "🔍 Введите поисковый запрос для расписания:")
}

func (b *Bot) disableSchedule(userID int64) {
	if b.DB != nil {
		b.DB.DeleteSchedule(userID)
	}
	b.sendMessage(userID, "✅ Расписание отключено")
}

func (b *Bot) handleEmploymentFilter(userID int64) {
	state := b.getUserState(userID)
	state.EmploymentFilter = append(state.EmploymentFilter, "трудоустройство")
	b.setUserState(userID, state)
	b.sendWithKeyboard(userID, "✅ Добавлен фильтр: Трудоустройство", keyboards.SearchFilters())
}

func (b *Bot) handleSelfEmployedFilter(userID int64) {
	state := b.getUserState(userID)
	state.EmploymentFilter = append(state.EmploymentFilter, "самозанятость")
	b.setUserState(userID, state)
	b.sendWithKeyboard(userID, "✅ Добавлен фильтр: Самозанятость", keyboards.SearchFilters())
}

func (b *Bot) handleIPFilter(userID int64) {
	state := b.getUserState(userID)
	state.EmploymentFilter = append(state.EmploymentFilter, "ИП")
	b.setUserState(userID, state)
	b.sendWithKeyboard(userID, "✅ Добавлен фильтр: ИП", keyboards.SearchFilters())
}

func (b *Bot) handleRemoteFilter(userID int64) {
	state := b.getUserState(userID)
	state.WorkFormatFilter = append(state.WorkFormatFilter, "удалёнка")
	b.setUserState(userID, state)
	b.sendWithKeyboard(userID, "✅ Добавлен фильтр: Удалёнка", keyboards.SearchFilters())
}

func (b *Bot) handleOfficeFilter(userID int64) {
	state := b.getUserState(userID)
	state.WorkFormatFilter = append(state.WorkFormatFilter, "офис")
	b.setUserState(userID, state)
	b.sendWithKeyboard(userID, "✅ Добавлен фильтр: Офис", keyboards.SearchFilters())
}

func (b *Bot) handleHybridFilter(userID int64) {
	state := b.getUserState(userID)
	state.WorkFormatFilter = append(state.WorkFormatFilter, "гибрид")
	b.setUserState(userID, state)
	b.sendWithKeyboard(userID, "✅ Добавлен фильтр: Гибрид", keyboards.SearchFilters())
}

func (b *Bot) showPaymentStatus(userID int64) {
	subscribed, payment, err := b.APIClient.IsSubscribed(userID)
	if err != nil {
		b.sendMessage(userID, "❌ Ошибка проверки подписки")
		return
	}

	if subscribed && payment != nil {
		text := fmt.Sprintf(`💳 Подписка:
Статус: ✅ Активна
Истекает: %s
Дней осталось: %d`, payment.ExpiresAt.Format("02.01.2006"), 0)
		b.sendWithKeyboard(userID, text, keyboards.PaymentMenu())
	} else {
		b.sendWithKeyboard(userID, "💳 Подписка не активна", keyboards.PaymentTariffs())
	}
}

func (b *Bot) showPaymentTariffs(userID int64) {
	b.sendWithKeyboard(userID, "Выберите срок подписки:", keyboards.PaymentTariffs())
}

func (b *Bot) processPayment(userID int64, duration int) {
	state := b.getUserState(userID)
	state.PaymentDuration = duration
	b.setUserState(userID, state)

	var periodText string
	var price int
	switch duration {
	case 30:
		periodText = "1 месяц"
		price = 299
	case 90:
		periodText = "3 месяца"
		price = 799
	case 365:
		periodText = "12 месяцев"
		price = 2499
	default:
		periodText = fmt.Sprintf("%d дней", duration)
		price = duration * 10
	}

	text := fmt.Sprintf(`💳 Оплата подписки: %s

📋 Реквизиты для оплаты:
卡 Номер карты: 2200 1234 5678 9012
🏦 Получатель: ИП Иванов И.И.

💬 В комментарии к платежу укажите:
%s

💰 Сумма: %d ₽

После оплаты нажмите "💰 Я оплатил"`, periodText, b.getPaymentComment(userID), price)

	b.sendWithKeyboard(userID, text, keyboards.PaymentConfirm())
}

func (b *Bot) getPaymentComment(userID int64) string {
	return fmt.Sprintf("PAY-%d", userID)
}

func (b *Bot) handlePaymentConfirmed(userID int64) {
	state := b.getUserState(userID)

	text := fmt.Sprintf(`✅ Отлично! Вы выбрали подписку на %d дней.

📧 Отправьте чек об оплате на почту:
📧 support@example.com

⏳ После проверки чека мы активируем подписку.
Обычно это занимает до 24 часов.`, state.PaymentDuration)

	b.sendMessage(userID, text)
	b.clearUserState(userID)
}

func (b *Bot) toggleNotifications(userID int64) {
	b.sendMessage(userID, "🔔 Уведомления переключены")
}

func (b *Bot) Stop() {
	if b.Scheduler != nil {
		b.Scheduler.Stop()
	}
	if b.Cache != nil {
		b.Cache.Close()
	}
	if b.DB != nil {
		b.DB.Close()
	}
	if b.Server != nil {
		b.Server.Shutdown(context.Background())
	}
}

func (b *Bot) handleCallbackData(callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	data := callback.Data

	if strings.HasPrefix(data, "resume_") && !strings.Contains(data, "_edit") && !strings.Contains(data, "_delete") {
		id := strings.TrimPrefix(data, "resume_")
		b.sendMessage(userID, fmt.Sprintf("Выбрано резюме #%s", id))
		return
	}

	if strings.HasPrefix(data, "resume_edit_") {
		id := strings.TrimPrefix(data, "resume_edit_")
		var resumeID int64
		fmt.Sscanf(id, "%d", &resumeID)

		state := b.getUserState(userID)
		state.State = states.StateEditResumeTitle
		state.EditingResumeID = resumeID
		b.setUserState(userID, state)

		b.sendMessage(userID, "✏️ Введите новое название резюме")
		return
	}

	if strings.HasPrefix(data, "resume_delete_") {
		id := strings.TrimPrefix(data, "resume_delete_")
		var resumeID int64
		fmt.Sscanf(id, "%d", &resumeID)

		state := b.getUserState(userID)
		state.EditingResumeID = resumeID
		b.setUserState(userID, state)

		b.sendWithKeyboard(userID, fmt.Sprintf("Удалить резюме #%d?", resumeID), keyboards.ConfirmDelete())
		return
	}

	if strings.HasPrefix(data, "vacancy_") && !strings.HasPrefix(data, "vacancy_start") {
		id := strings.TrimPrefix(data, "vacancy_")
		var vacancyID int64
		fmt.Sscanf(id, "%d", &vacancyID)

		b.sendMessage(userID, fmt.Sprintf("📄 Детали вакансии #%d", vacancyID))
		return
	}

	if strings.HasPrefix(data, "response_") {
		id := strings.TrimPrefix(data, "response_")
		var vacancyID int64
		fmt.Sscanf(id, "%d", &vacancyID)

		err := b.APIClient.SaveVacancyResponse(userID, vacancyID)
		if err != nil {
			b.sendMessage(userID, "❌ Ошибка отправки отклика")
		} else {
			b.sendMessage(userID, "✅ Отклик отправлен!")
		}
		return
	}

	if strings.HasPrefix(data, "save_") {
		id := strings.TrimPrefix(data, "save_")
		var vacancyID int64
		fmt.Sscanf(id, "%d", &vacancyID)

		b.sendMessage(userID, fmt.Sprintf("💾 Вакансия #%d сохранена", vacancyID))
		return
	}

	if strings.HasPrefix(data, "time_") {
		timeStr := strings.TrimPrefix(data, "time_")
		state := b.getUserState(userID)
		state.ScheduleTime = timeStr
		b.setUserState(userID, state)

		if b.DB != nil {
			b.DB.SaveSchedule(userID, "", timeStr, state.ScheduleResumeID, state.SearchQuery)
		}

		b.sendMessage(userID, fmt.Sprintf("🕐 Время установлено: %s", timeStr))
		b.handleScheduleMenuByID(userID)
		return
	}

	if strings.HasPrefix(data, "days_") {
		days := strings.TrimPrefix(data, "days_")
		state := b.getUserState(userID)
		state.ScheduleDays = parseScheduleDays(days)
		b.setUserState(userID, state)

		if b.DB != nil {
			b.DB.SaveSchedule(userID, "", state.ScheduleTime, state.ScheduleResumeID, state.SearchQuery)
		}

		b.sendMessage(userID, fmt.Sprintf("📅 Дни установлены: %s", days))
		b.handleScheduleMenuByID(userID)
		return
	}

	if data == "confirm_delete_yes" {
		state := b.getUserState(userID)
		resumeID := state.EditingResumeID

		if resumeID > 0 {
			err := b.APIClient.DeleteResume(resumeID)
			if err != nil {
				b.sendMessage(userID, "❌ Ошибка удаления резюме")
			} else {
				b.sendMessage(userID, "✅ Резюме удалено")
				if b.Cache != nil {
					b.Cache.InvalidateResumesCache(userID)
				}
			}
		}
		state.EditingResumeID = 0
		b.clearUserState(userID)
		b.showResumeList(userID)
		return
	}

	if data == "confirm_delete_no" {
		b.showResumeList(userID)
		return
	}

	if data == "cancel_search" {
		b.clearUserState(userID)
		b.sendWithKeyboard(userID, "❌ Поиск отменён", keyboards.VacanciesMenu())
		return
	}

	if strings.HasPrefix(data, "select_resume_") {
		id := strings.TrimPrefix(data, "select_resume_")
		var resumeID int64
		fmt.Sscanf(id, "%d", &resumeID)

		state := b.getUserState(userID)
		state.SelectedResumeID = resumeID
		state.State = states.StateSearchQuery
		b.setUserState(userID, state)

		b.sendMessage(userID, "Введите поисковый запрос")
		return
	}
}

func parseScheduleDays(days string) []string {
	switch days {
	case "everyday":
		return []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}
	case "monwedfri":
		return []string{"mon", "wed", "fri"}
	case "tuethusat":
		return []string{"tue", "thu", "sat"}
	case "weekdays":
		return []string{"mon", "tue", "wed", "thu", "fri"}
	case "weekend":
		return []string{"sat", "sun"}
	default:
		return []string{}
	}
}

func verifyWebhookSignature(token, secret, data, signature string) bool {
	if secret == "" {
		return true
	}
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	expected := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
