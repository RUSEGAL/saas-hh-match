package states

type UserState string

const (
	StateNone              UserState = ""
	StateMainMenu          UserState = "main_menu"
	StateCreatingResume    UserState = "creating_resume"
	StateResumeTitle       UserState = "resume_title"
	StateResumeContent     UserState = "resume_content"
	StateEditingResume     UserState = "editing_resume"
	StateEditResumeTitle   UserState = "edit_resume_title"
	StateEditResumeContent UserState = "edit_resume_content"
	StateVacancySearch     UserState = "vacancy_search"
	StateSelectResume      UserState = "select_resume"
	StateSearchQuery       UserState = "search_query"
	StateSearchFilters     UserState = "search_filters"
	StateScheduleSetup     UserState = "schedule_setup"
	StateScheduleTime      UserState = "schedule_time"
	StateScheduleDays      UserState = "schedule_days"
	StateScheduleQuery     UserState = "schedule_query"
	StatePaymentSelect     UserState = "payment_select"
	StateNoSubscription    UserState = "no_subscription"
	StateExpired           UserState = "expired"
)

type UserStateData struct {
	State            UserState
	ResumeTitle      string
	ResumeContent    string
	EditingResumeID  int64
	SelectedResumeID int64
	SearchQuery      string
	ScheduleTime     string
	ScheduleDays     []string
	ScheduleResumeID int64
	PaymentDuration  int
	EmploymentFilter []string
	WorkFormatFilter []string
}

type StateManager struct {
	states map[int64]*UserStateData
}

func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[int64]*UserStateData),
	}
}

func (m *StateManager) GetState(userID int64) *UserStateData {
	if state, ok := m.states[userID]; ok {
		return state
	}
	m.states[userID] = &UserStateData{State: StateNone}
	return m.states[userID]
}

func (m *StateManager) SetState(userID int64, state UserState) {
	if _, ok := m.states[userID]; !ok {
		m.states[userID] = &UserStateData{}
	}
	m.states[userID].State = state
}

func (m *StateManager) ClearState(userID int64) {
	m.states[userID] = &UserStateData{State: StateNone}
}

func (m *StateManager) SetStateData(userID int64, data *UserStateData) {
	m.states[userID] = data
}

func (m *StateManager) IsInState(userID int64, state UserState) bool {
	if s, ok := m.states[userID]; ok {
		return s.State == state
	}
	return false
}
