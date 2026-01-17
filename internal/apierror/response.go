package apierror

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type errorResponse struct {
	Errors []*AppError `json:"errors"`
}

func Abort(
	c *gin.Context,
	err error,
) {
	var finalErr *AppError

	if err == nil {
		finalErr = Errors.INTERNAL_ERROR.Wrap(fmt.Errorf("nil error passed to Abort"))
	} else {
		if ae, ok := err.(*AppError); ok {
			finalErr = ae
		} else {
			finalErr = Errors.INTERNAL_ERROR.Wrap(err)
		}
	}
	if log != nil {
		logFields := []zap.Field{
			zap.String("code", finalErr.Code),
			zap.Int("http_status", finalErr.Status),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		}

		if finalErr.Err != nil {
			logFields = append(logFields, zap.Error(finalErr.Err))

			if finalErr.Status >= 500 {
				log.Error("Server Error Response", logFields...)
			} else {
				log.Warn("Client Error Response with Cause", logFields...)
			}
		} else {
			log.Info("Client Error Response", logFields...)
		}
	}
	c.AbortWithStatusJSON(
		finalErr.Status, errorResponse{
			Errors: []*AppError{finalErr},
		},
	)
}
