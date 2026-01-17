package fail2ban

// Константы для работы с внешними командами и парсинга
const (
	CmdSudo               = "sudo"
	CmdFail2Ban           = "fail2ban-client"
	ArgStatus             = "status"
	ArgSet                = "set"
	ArgUnbanIP            = "unbanip"
	ParamJailName         = "name"
	ReJailList            = `Jail list:\s*(.*)`
	ReCurrentlyFailed     = `Currently failed:\s*(\d+)`
	ReTotalFailed         = `Total failed:\s*(\d+)`
	ReCurrentlyBanned     = `Currently banned:\s*(\d+)`
	ReTotalBanned         = `Total banned:\s*(\d+)`
	ReBannedIPList        = `Banned IP list:\s*([\s\S]*)`
	ErrOutputDoesNotExist = "Does not exist"
	ErrOutputNotFound     = "not found"
	ErrOutputIsNotBanned  = "is not banned"
	ErrOutputJailNotFound = "Jail not found"
)

type Fail2BanStatusDTO struct {
	JailCount int      `json:"jail_count" example:"3"`
	JailList  []string `json:"jail_list" example:"['sshd', 'nginx-forbidden']"`
}

type JailDetailsDTO struct {
	JailName        string   `json:"jail_name" example:"sshd"`
	CurrentlyFailed int      `json:"currently_failed" example:"2"`
	TotalFailed     int      `json:"total_failed" example:"284"`
	CurrentlyBanned int      `json:"currently_banned" example:"259"`
	TotalBanned     int      `json:"total_banned" example:"259"`
	BannedIPList    []string `json:"banned_ip_list" example:"['1.2.3.4', '5.6.7.8']"`
}

type BanActionRequest struct {
	IP   string `json:"ip" binding:"required,ip" example:"1.2.3.4"`
	Jail string `json:"jail" binding:"required" example:"sshd"`
}

type BanActionResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"IP unbanned successfully"`
}
