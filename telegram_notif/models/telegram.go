package models

type TelegramUpdate struct {
	UpdateID int64            `json:"update_id"`
	Message  *TelegramMessage `json:"message"`
}

type TelegramMessage struct {
	MessageID int64        `json:"message_id"`
	Text      string       `json:"text"`
	Chat      TelegramChat `json:"chat"`
	From      TelegramUser `json:"from"`
}

type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type TelegramUpdatesResponse struct {
	OK          bool             `json:"ok"`
	Result      []TelegramUpdate `json:"result"`
	Description string           `json:"description"`
}

type TelegramBasicResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}
