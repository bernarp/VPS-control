package pm2

type Action string

const (
	ActionStart   Action = "start"
	ActionStop    Action = "stop"
	ActionRestart Action = "restart"
)

type ProcessBasicDTO struct {
	Name   string `json:"name" example:"discordBot-DEV"`
	PID    int    `json:"pid" example:"697065"`
	Active bool   `json:"active" example:"true"`
}

type ProcessWithCwdDTO struct {
	Name   string `json:"name" example:"discordBot-DEV"`
	PID    int    `json:"pid" example:"697065"`
	Cwd    string `json:"cwd" example:"/opt/apps/wthBotStatistics"`
	Active bool   `json:"active" example:"true"`
}

type ProcessFullDTO struct {
	Name      string  `json:"name" example:"discordBot-DEV"`
	PID       int     `json:"pid" example:"697065"`
	Cwd       string  `json:"cwd" example:"/opt/apps/wthBotStatistics"`
	Mem       float64 `json:"mem" example:"10.9"`
	CPU       float64 `json:"cpu" example:"1.6"`
	StartedAt string  `json:"started_at" example:"2026-01-13T10:25:43+06:00"`
	Active    bool    `json:"active" example:"true"`
}

type ProcessBasicGrouped map[string][]ProcessBasicDTO
type ProcessWithCwdGrouped map[string][]ProcessWithCwdDTO
type ProcessFullGrouped map[string][]ProcessFullDTO

type ProcessActionResponse struct {
	Success bool   `json:"success" example:"true"`
	Action  Action `json:"action" example:"restart"`
	Target  string `json:"target" example:"discordBot-DEV"`
	Message string `json:"message" example:"process action executed"`
}
