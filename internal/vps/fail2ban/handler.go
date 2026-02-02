package fail2ban

import (
	"VPS-control/internal/apierror"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type handler struct {
	controlSvc fail2banControl
	logger     *zap.Logger
}

func NewHandler(
	cs fail2banControl,
	l *zap.Logger,
) Handler {
	return &handler{
		controlSvc: cs,
		logger:     l,
	}
}

// GetStatus godoc
// @Summary      Get all jails status
// @Tags         fail2ban
// @Security     CookieAuth
// @Produce      json
// @Success      200  {object}  Fail2BanStatusDTO
// @Router       /vps/fail2ban/status [get]
func (h *handler) GetStatus(c *gin.Context) {
	data, err := h.controlSvc.GetGlobalStatus()
	if err != nil {
		apierror.Abort(c, apierror.Errors.INTERNAL_ERROR.Wrap(err))
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetJailDetails godoc
// @Summary      Get specific jail details
// @Tags         fail2ban
// @Security     CookieAuth
// @Param        name  query  string  true  "Jail Name"
// @Produce      json
// @Success      200  {object}  JailDetailsDTO
// @Failure      400  {object}  apierror.AppError
// @Failure      404  {object}  apierror.AppError
// @Router       /vps/fail2ban/jail [get]
func (h *handler) GetJailDetails(c *gin.Context) {
	jailName := c.Query(ParamJailName)
	if jailName == "" {
		apierror.Abort(c, apierror.Errors.INVALID_REQUEST.WithMeta("query parameter 'name' is required"))
		return
	}

	data, err := h.controlSvc.GetJailDetails(jailName)
	if err != nil {
		apierror.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

// Unban godoc
// @Summary      Unban an IP
// @Tags         fail2ban
// @Security     CookieAuth
// @Accept       json
// @Param        request  body  BanActionRequest  true  "Unban details"
// @Success      200      {object}  BanActionResponse
// @Router       /vps/fail2ban/unban [post]
func (h *handler) Unban(c *gin.Context) {
	var req BanActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Abort(c, apierror.Errors.INVALID_REQUEST)
		return
	}

	if err := h.controlSvc.UnbanIP(req.Jail, req.IP); err != nil {
		apierror.Abort(c, apierror.Errors.INTERNAL_ERROR.Wrap(err))
		return
	}

	c.JSON(
		http.StatusOK, BanActionResponse{
			Success: true,
			Message: "IP unbanned successfully",
		},
	)
}
