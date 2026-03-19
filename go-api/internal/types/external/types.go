package types_external

type GenerateTokenRequest struct {
	User       string `json:"user"`
	Password   string `json:"password"`
	TelegramID int64  `json:"telegram_id,omitempty"`
}
type PaymentRequest struct {
	ID          int64
	UserID      int64
	TelegrammID int64
	Amount      int64
	Status      string
	Provider    string
}
type ResumeRequest struct {
	UserID   int64  `json:"user_id,omitempty"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Telegram bool   `json:"telegram,omitempty"`
}
