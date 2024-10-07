package dto

type NotificationDTO struct {
	UUID     string `json:"uuid"`
	Type     string `json:"type"`
	TypeName string `json:"type_name"`

	Count map[string]interface{} `json:"count"`

	Star bool `json:"star"`

	Score float64 `json:"score"`
}

type NotificationTaskDTO struct {
	UUID string `json:"uuid"`
	Name string `json:"type_name"`
	Type string `json:"type"`

	Count map[string]interface{} `json:"count"`

	Comments  []CommentDTO  `json:"comments"`
	Reminders []ReminderDTO `json:"reminders"`
	Uploads   []FileDTOs    `json:"uploads"`

	Score  float64 `json:"score"`
	Opened bool    `json:"opened"`
	Group  int     `json:"group"`

	Star bool `json:"star"`
}

type NotificationReminderDTO struct {
	UUID string `json:"uuid"`
	Name string `json:"type_name"`
	Type string `json:"type"`

	Score  float64 `json:"score"`
	Opened bool    `json:"opened"`
	Group  int     `json:"group"`
}
