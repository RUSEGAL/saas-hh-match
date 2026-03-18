package keyboards

import (
	"fmt"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-bot/internal/api"
)

func MainMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 Мои резюме", "menu_resumes"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 Поиск вакансий", "menu_vacancies"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳 Оплата", "menu_payment"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 Статистика", "menu_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Настройки", "menu_settings"),
		),
	)
}

func ResumesMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Создать резюме", "resume_create"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Список резюме", "resume_list"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_main"),
		),
	)
}

func VacanciesMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚀 Начать поиск", "vacancy_start"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Расписание", "vacancy_schedule"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 История поиска", "vacancy_history"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_main"),
		),
	)
}

func PaymentMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Текущий статус", "payment_status"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💰 Оплатить", "payment_pay"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_main"),
		),
	)
}

func PaymentTariffs() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1 месяц — 299₽", "payment_1month"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("3 месяца — 799₽ (скидка 10%)", "payment_3month"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("12 месяцев — 2499₽ (скидка 30%)", "payment_12month"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "menu_payment"),
		),
	)
}

func PaymentConfirm() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💰 Я оплатил", "payment_confirmed"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Отмена", "menu_payment"),
		),
	)
}

func SettingsMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔔 Уведомления вкл/выкл", "settings_notifications"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "back_main"),
		),
	)
}

func BackToMain() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Главное меню", "back_main"),
		),
	)
}

func ResumeList(resumes []api.Resume) [][]tgbotapi.InlineKeyboardButton {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, r := range resumes {
		status := "⏳ Анализ..."
		if r.Analyzed {
			status = fmt.Sprintf("✅ AI: %.0f%%", r.AIScore*100)
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s (%s)", r.Title, status),
				fmt.Sprintf("resume_%d", r.ID),
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "menu_resumes"),
	))
	return rows
}

func ResumeActions(resumeID int64) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Редактировать", fmt.Sprintf("resume_edit_%d", resumeID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить", fmt.Sprintf("resume_delete_%d", resumeID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "resume_list"),
		),
	)
}

func SearchFilters() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Трудоустройство", "filter_employment"),
			tgbotapi.NewInlineKeyboardButtonData("Самозанятость", "filter_selfemployed"),
			tgbotapi.NewInlineKeyboardButtonData("ИП", "filter_ip"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Удалёнка", "filter_remote"),
			tgbotapi.NewInlineKeyboardButtonData("Офис", "filter_office"),
			tgbotapi.NewInlineKeyboardButtonData("Гибрид", "filter_hybrid"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пропустить", "filter_skip"),
		),
	)
}

func SearchResults(vacancies []api.Vacancy) [][]tgbotapi.InlineKeyboardButton {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, v := range vacancies {
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
		default:
			emoji = "🔹"
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s %s", emoji, v.Title), fmt.Sprintf("vacancy_%d", v.ID)),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📨 Откликнуться", fmt.Sprintf("response_%d", v.ID)),
			tgbotapi.NewInlineKeyboardButtonData("💾 Сохранить", fmt.Sprintf("save_%d", v.ID)),
		))
	}
	return rows
}

func ScheduleMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🕐 Установить время", "schedule_time"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Выбрать дни", "schedule_days"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 Поисковый запрос", "schedule_query"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отключить", "schedule_disable"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "menu_vacancies"),
		),
	)
}

func ScheduleTimeOptions() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("10:00", "time_10:00"),
			tgbotapi.NewInlineKeyboardButtonData("12:00", "time_12:00"),
			tgbotapi.NewInlineKeyboardButtonData("14:00", "time_14:00"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("18:00", "time_18:00"),
			tgbotapi.NewInlineKeyboardButtonData("20:00", "time_20:00"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "vacancy_schedule"),
		),
	)
}

func ScheduleDaysOptions() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Каждый день", "days_everyday"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пн-Ср-Пт", "days_monwedfri"),
			tgbotapi.NewInlineKeyboardButtonData("Вт-Чт-Сб", "days_tuethusat"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("По будням", "days_weekdays"),
			tgbotapi.NewInlineKeyboardButtonData("По выходным", "days_weekend"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "vacancy_schedule"),
		),
	)
}

func CancelSearch() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel_search"),
		),
	)
}

func ConfirmDelete() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да, удалить", "confirm_delete_yes"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет", "confirm_delete_no"),
		),
	)
}

func YesNoButtons(yesCallback, noCallback string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да", yesCallback),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет", noCallback),
		),
	)
}
