package fail2ban

import "github.com/gin-gonic/gin"

type Handler interface {
	GetStatus(c *gin.Context)
	GetJailDetails(c *gin.Context)
	Unban(c *gin.Context)
}

type fail2banControl interface {
	GetGlobalStatus() (*Fail2BanStatusDTO, error)
	GetJailDetails(jailName string) (*JailDetailsDTO, error)
	UnbanIP(jail, ip string) error
}
