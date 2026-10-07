package models

// AiChatMessage riwayat singkat percakapan.
type AiChatMessage struct {
	Role    string `json:"role"` // user | assistant
	Content string `json:"content"`
}

// AiChatAction saran aksi UI yang bisa diterapkan user.
type AiChatAction struct {
	Type        string `json:"type"` // set_location | set_shift | set_status | set_date | navigate | refresh_insight | open_chart
	Value       string `json:"value"`
	Label       string `json:"label,omitempty"`
	NeedsConfirm bool  `json:"needsConfirm"`
}

// AiChatRequest body POST /api/ai/chat.
type AiChatRequest struct {
	Question    string                 `json:"question"`
	Page        string                 `json:"page"`
	Date        string                 `json:"date"`
	Shift       string                 `json:"shift"`
	Location    string                 `json:"location"`
	Status      string                 `json:"status"`
	History     []AiChatMessage        `json:"history"`
	Machines    []AiInsightMachineInput `json:"machines"`
	Insight     *AiInsightAnalyzeResponse `json:"insight,omitempty"`
}

// AiChatResponse jawaban Copilot.
type AiChatResponse struct {
	Status   string         `json:"status"`
	Reply    string         `json:"reply"`
	Actions  []AiChatAction `json:"actions"`
	Model    string         `json:"model,omitempty"`
	Warning  string         `json:"warning,omitempty"`
}
