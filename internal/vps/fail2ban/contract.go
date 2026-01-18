package fail2ban

type fail2banControl interface {
	GetGlobalStatus() (*Fail2BanStatusDTO, error)
	GetJailDetails(jailName string) (*JailDetailsDTO, error)
	UnbanIP(jail, ip string) error
	parseIntField(input, pattern string) int
}
