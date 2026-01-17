package middleware

import (
	"bytes"
	"io"
	"regexp"
	"strings"

	"DiscordBotControl/internal/apierror"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type InputSanitizer struct {
	logger  *zap.Logger
	sqlRe   *regexp.Regexp
	shellRe *regexp.Regexp
	pathRe  *regexp.Regexp
}

func NewInputSanitizer(logger *zap.Logger) *InputSanitizer {
	return &InputSanitizer{
		logger: logger.Named("input-sanitizer"),

		sqlRe: regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE|DROP|UNION|ALTER|TRUNCATE|EXEC|EXECUTE|DECLARE|CAST)\b.*\b(FROM|INTO|TABLE|WHERE|SET|VALUES)\b|--|\b(OR|AND)\b\s+\d+\s*=\s*\d+|'\s*(OR|AND)\s*'`),

		shellRe: regexp.MustCompile(`(?i)(^|[;&|])\s*(sudo|rm\s+-rf|wget|curl|nc|netcat|bash|sh\s+-c|eval|exec)\b|\$\(|` + "`" + `|>\s*/|<\s*/|\|\s*(bash|sh)|;\s*(rm|cat|chmod)`),

		pathRe: regexp.MustCompile(`\.\./|\.\.\\|%2e%2e%2f|%2e%2e/|\.%2e/|%2e\./`),
	}
}

func (s *InputSanitizer) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, param := range c.Params {
			if s.isMalicious(param.Value) {
				s.logAttempt(c, "path_param", param.Key, param.Value)
				apierror.Abort(c, apierror.Errors.MALICIOUS_INPUT_DETECTED)
				return
			}
		}

		for key, values := range c.Request.URL.Query() {
			for _, val := range values {
				if s.isMalicious(val) {
					s.logAttempt(c, "query_param", key, val)
					apierror.Abort(c, apierror.Errors.MALICIOUS_INPUT_DETECTED)
					return
				}
			}
		}

		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			body, err := io.ReadAll(c.Request.Body)
			if err == nil && len(body) > 0 {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

				bodyStr := string(body)
				if s.isMalicious(bodyStr) {
					s.logAttempt(c, "body", "", bodyStr)
					apierror.Abort(c, apierror.Errors.MALICIOUS_INPUT_DETECTED)
					return
				}
			}
		}

		for _, header := range []string{"X-Forwarded-For", "Referer", "User-Agent"} {
			if val := c.GetHeader(header); val != "" && s.isMalicious(val) {
				s.logAttempt(c, "header", header, val)
				apierror.Abort(c, apierror.Errors.MALICIOUS_INPUT_DETECTED)
				return
			}
		}

		c.Next()
	}
}

func (s *InputSanitizer) isMalicious(input string) bool {
	if input == "" || len(input) > 10000 {
		return len(input) > 10000
	}

	normalized := strings.ToLower(input)

	return s.pathRe.MatchString(normalized) ||
		s.sqlRe.MatchString(input) ||
		s.shellRe.MatchString(input)
}

func (s *InputSanitizer) logAttempt(
	c *gin.Context,
	source, key, value string,
) {
	if len(value) > 200 {
		value = value[:200] + "..."
	}

	s.logger.Warn(
		"Malicious input detected",
		zap.String("ip", c.ClientIP()),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("source", source),
		zap.String("key", key),
		zap.String("value", value),
	)
}
