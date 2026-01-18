package pm2

import "github.com/gin-gonic/gin"

type Handler interface {
	GetProcessesBasic(c *gin.Context)
	GetProcessesWithCwd(c *gin.Context)
	GetProcessesFull(c *gin.Context)
	Restart(c *gin.Context)
	Start(c *gin.Context)
	Stop(c *gin.Context)
}

type ProcessLister interface {
	GetProcessesBasic() (ProcessBasicGrouped, error)
	GetProcessesWithCwd() (ProcessWithCwdGrouped, error)
	GetProcessesFull() (ProcessFullGrouped, error)
}

type ProcessController interface {
	Restart(target string) (string, error)
	Start(target string) (string, error)
	Stop(target string) (string, error)
}
