package fail2ban

import (
	"DiscordBotControl/internal/apierror"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	controlSvc *ControlService
	logger     *zap.Logger
}

func NewHandler(
	cs *ControlService,
	l *zap.Logger,
) *Handler {
	return &Handler{
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
func (h *Handler) GetStatus(c *gin.Context) {
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
// @Param        name  path  string  true  "Jail Name"
// @Produce      json
// @Success      200  {object}  JailDetailsDTO
// @Router       /vps/fail2ban/status/{name} [get]
func (h *Handler) GetJailDetails(c *gin.Context) {
	jailName := c.Param(ParamJailName)
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
func (h *Handler) Unban(c *gin.Context) {
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
