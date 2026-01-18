package pm2

import (
	"VPS-control/internal/apierror"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	listSvc    ProcessLister
	controlSvc ProcessController
	logger     *zap.Logger
}

func NewHandler(
	ls ProcessLister,
	cs ProcessController,
	l *zap.Logger,
) *Handler {
	return &Handler{
		listSvc:    ls,
		controlSvc: cs,
		logger:     l,
	}
}

// GetProcessesBasic godoc
// @Summary      Get basic PM2 processes grouped by PPID
// @Description  Returns processes grouped by parent PID with name, PID and active status
// @Tags         pm2
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  ProcessBasicGrouped
// @Failure      401  {object}  apierror.AppError
// @Failure      500  {object}  apierror.AppError
// @Router       /vps/pm2/processes/basic [get]
func (h *Handler) GetProcessesBasic(c *gin.Context) {
	data, err := h.listSvc.GetProcessesBasic()
	if err != nil {
		apierror.Abort(c, apierror.Errors.INTERNAL_ERROR.Wrap(err))
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetProcessesWithCwd godoc
// @Summary      Get PM2 processes with cwd grouped by PPID
// @Description  Returns processes grouped by parent PID with name, PID, working directory and active status
// @Tags         pm2
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  ProcessWithCwdGrouped
// @Failure      401  {object}  apierror.AppError
// @Failure      500  {object}  apierror.AppError
// @Router       /vps/pm2/processes/cwd [get]
func (h *Handler) GetProcessesWithCwd(c *gin.Context) {
	data, err := h.listSvc.GetProcessesWithCwd()
	if err != nil {
		apierror.Abort(c, apierror.Errors.INTERNAL_ERROR.Wrap(err))
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetProcessesFull godoc
// @Summary      Get full PM2 processes grouped by PPID
// @Description  Returns processes grouped by parent PID with full metrics (CPU, Memory, started_at, active)
// @Tags         pm2
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  ProcessFullGrouped
// @Failure      401  {object}  apierror.AppError
// @Failure      500  {object}  apierror.AppError
// @Router       /vps/pm2/processes/full [get]
func (h *Handler) GetProcessesFull(c *gin.Context) {
	data, err := h.listSvc.GetProcessesFull()
	if err != nil {
		apierror.Abort(c, apierror.Errors.INTERNAL_ERROR.Wrap(err))
		return
	}
	c.JSON(http.StatusOK, data)
}

// Restart godoc
// @Summary      Restart PM2 process
// @Description  Restarts a PM2 process by name or PID
// @Tags         pm2
// @Security     CookieAuth
// @Param        name  path  string  true  "Process Name or PID"
// @Produce      json
// @Success      200  {object}  ProcessActionResponse
// @Failure      401  {object}  apierror.AppError
// @Failure      404  {object}  apierror.AppError
// @Failure      500  {object}  apierror.AppError
// @Router       /vps/pm2/{name}/restart [post]
func (h *Handler) Restart(c *gin.Context) {
	target := c.Param("name")
	if target == "" {
		apierror.Abort(c, apierror.Errors.INVALID_REQUEST)
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
// @Description  Starts a PM2 process by name or PID
// @Tags         pm2
// @Security     CookieAuth
// @Param        name  path  string  true  "Process Name or PID"
// @Produce      json
// @Success      200  {object}  ProcessActionResponse
// @Failure      401  {object}  apierror.AppError
// @Failure      404  {object}  apierror.AppError
// @Failure      500  {object}  apierror.AppError
// @Router       /vps/pm2/{name}/start [post]
func (h *Handler) Start(c *gin.Context) {
	target := c.Param("name")
	if target == "" {
		apierror.Abort(c, apierror.Errors.INVALID_REQUEST)
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
// @Description  Stops a PM2 process by name or PID
// @Tags         pm2
// @Security     CookieAuth
// @Param        name  path  string  true  "Process Name or PID"
// @Produce      json
// @Success      200  {object}  ProcessActionResponse
// @Failure      401  {object}  apierror.AppError
// @Failure      404  {object}  apierror.AppError
// @Failure      500  {object}  apierror.AppError
// @Router       /vps/pm2/{name}/stop [post]
func (h *Handler) Stop(c *gin.Context) {
	target := c.Param("name")
	if target == "" {
		apierror.Abort(c, apierror.Errors.INVALID_REQUEST)
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
