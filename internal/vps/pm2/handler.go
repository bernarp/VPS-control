package pm2

import (
	"VPS-control/internal/apierror"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type handler struct {
	listSvc    ProcessLister
	controlSvc ProcessController
	logger     *zap.Logger
}

func NewHandler(
	ls ProcessLister,
	cs ProcessController,
	l *zap.Logger,
) Handler {
	return &handler{
		listSvc:    ls,
		controlSvc: cs,
		logger:     l,
	}
}

// GetProcessesBasic godoc
// @Summary      Get basic PM2 processes
// @Description  Returns processes grouped by PPID. Optional filter by ppid.
// @Tags         pm2
// @Security     CookieAuth
// @Param        ppid query string false "Filter by Parent PID"
// @Produce      json
// @Success      200  {object}  ProcessBasicGrouped
// @Router       /vps/pm2/processes/basic [get]
func (h *handler) GetProcessesBasic(c *gin.Context) {
	ppid := c.Query("ppid")
	data, err := h.listSvc.GetProcessesBasic()
	if err != nil {
		apierror.Abort(c, apierror.Errors.INTERNAL_ERROR.Wrap(err))
		return
	}

	if ppid != "" {
		if val, ok := data[ppid]; ok {
			c.JSON(http.StatusOK, ProcessBasicGrouped{ppid: val})
			return
		}
		c.JSON(http.StatusOK, ProcessBasicGrouped{})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetProcessesWithCwd godoc
// @Summary      Get PM2 processes with cwd
// @Description  Returns processes with working directory. Optional filter by ppid.
// @Tags         pm2
// @Security     CookieAuth
// @Param        ppid query string false "Filter by Parent PID"
// @Produce      json
// @Success      200  {object}  ProcessWithCwdGrouped
// @Router       /vps/pm2/processes/cwd [get]
func (h *handler) GetProcessesWithCwd(c *gin.Context) {
	ppid := c.Query("ppid")
	data, err := h.listSvc.GetProcessesWithCwd()
	if err != nil {
		apierror.Abort(c, apierror.Errors.INTERNAL_ERROR.Wrap(err))
		return
	}

	if ppid != "" {
		if val, ok := data[ppid]; ok {
			c.JSON(http.StatusOK, ProcessWithCwdGrouped{ppid: val})
			return
		}
		c.JSON(http.StatusOK, ProcessWithCwdGrouped{})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetProcessesFull godoc
// @Summary      Get full PM2 processes
// @Description  Returns processes with metrics. Optional filter by ppid.
// @Tags         pm2
// @Security     CookieAuth
// @Param        ppid query string false "Filter by Parent PID"
// @Produce      json
// @Success      200  {object}  ProcessFullGrouped
// @Router       /vps/pm2/processes/full [get]
func (h *handler) GetProcessesFull(c *gin.Context) {
	ppid := c.Query("ppid")
	data, err := h.listSvc.GetProcessesFull()
	if err != nil {
		apierror.Abort(c, apierror.Errors.INTERNAL_ERROR.Wrap(err))
		return
	}

	if ppid != "" {
		if val, ok := data[ppid]; ok {
			c.JSON(http.StatusOK, ProcessFullGrouped{ppid: val})
			return
		}
		c.JSON(http.StatusOK, ProcessFullGrouped{})
		return
	}

	c.JSON(http.StatusOK, data)
}

// Restart godoc
// @Summary      Restart PM2 process
// @Tags         pm2
// @Security     CookieAuth
// @Param        name  query  string  true  "Process Name or PID"
// @Produce      json
// @Success      200  {object}  ProcessActionResponse
// @Router       /vps/pm2/restart [post]
func (h *handler) Restart(c *gin.Context) {
	target := c.Query("name")
	if target == "" {
		apierror.Abort(c, apierror.Errors.INVALID_REQUEST.WithMeta("query parameter 'name' is required"))
		return
	}

	resolved, err := h.controlSvc.Restart(target)
	if err != nil {
		apierror.Abort(c, err)
		return
	}

	c.JSON(
		http.StatusOK, ProcessActionResponse{
			Success: true,
			Action:  ActionRestart,
			Target:  resolved,
			Message: "process restarted successfully",
		},
	)
}

// Start godoc
// @Summary      Start PM2 process
// @Tags         pm2
// @Security     CookieAuth
// @Param        name  query  string  true  "Process Name or PID"
// @Produce      json
// @Success      200  {object}  ProcessActionResponse
// @Router       /vps/pm2/start [post]
func (h *handler) Start(c *gin.Context) {
	target := c.Query("name")
	if target == "" {
		apierror.Abort(c, apierror.Errors.INVALID_REQUEST.WithMeta("query parameter 'name' is required"))
		return
	}

	resolved, err := h.controlSvc.Start(target)
	if err != nil {
		apierror.Abort(c, err)
		return
	}

	c.JSON(
		http.StatusOK, ProcessActionResponse{
			Success: true,
			Action:  ActionStart,
			Target:  resolved,
			Message: "process started successfully",
		},
	)
}

// Stop godoc
// @Summary      Stop PM2 process
// @Tags         pm2
// @Security     CookieAuth
// @Param        name  query  string  true  "Process Name or PID"
// @Produce      json
// @Success      200  {object}  ProcessActionResponse
// @Router       /vps/pm2/stop [post]
func (h *handler) Stop(c *gin.Context) {
	target := c.Query("name")
	if target == "" {
		apierror.Abort(c, apierror.Errors.INVALID_REQUEST.WithMeta("query parameter 'name' is required"))
		return
	}

	resolved, err := h.controlSvc.Stop(target)
	if err != nil {
		apierror.Abort(c, err)
		return
	}

	c.JSON(
		http.StatusOK, ProcessActionResponse{
			Success: true,
			Action:  ActionStop,
			Target:  resolved,
			Message: "process stopped successfully",
		},
	)
}
