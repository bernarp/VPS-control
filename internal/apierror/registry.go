package apierror

import (
	"fmt"
	"reflect"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type errorRegistry struct {
	INTERNAL_ERROR           *AppError
	INVALID_REQUEST          *AppError
	DATABASE_ERROR           *AppError
	INVALID_CREDENTIALS      *AppError
	TOKEN_EXPIRED            *AppError
	PERMISSION_DENIED        *AppError
	RATE_LIMIT_EXCEEDED      *AppError
	PM2_PROCESS_NOT_FOUND    *AppError
	ACTION_NOT_ALLOWED       *AppError
	PROCESS_ALREADY_RUNNING  *AppError
	PROCESS_ALREADY_STOPPED  *AppError
	MALICIOUS_INPUT_DETECTED *AppError
	FAIL2BAN_JAIL_NOT_FOUND  *AppError
	FAIL2BAN_IP_NOT_BANNED   *AppError
	FAIL2BAN_EXECUTION_ERROR *AppError
}

var Errors = &errorRegistry{
	INTERNAL_ERROR:           &AppError{Code: "INTERNAL_ERROR", Status: 500},
	INVALID_REQUEST:          &AppError{Code: "INVALID_REQUEST", Status: 400},
	DATABASE_ERROR:           &AppError{Code: "DATABASE_ERROR", Status: 500},
	INVALID_CREDENTIALS:      &AppError{Code: "INVALID_CREDENTIALS", Status: 401},
	TOKEN_EXPIRED:            &AppError{Code: "TOKEN_EXPIRED", Status: 401},
	PERMISSION_DENIED:        &AppError{Code: "PERMISSION_DENIED", Status: 403},
	RATE_LIMIT_EXCEEDED:      &AppError{Code: "RATE_LIMIT_EXCEEDED", Status: 429},
	PM2_PROCESS_NOT_FOUND:    &AppError{Code: "PM2_PROCESS_NOT_FOUND", Status: 404},
	ACTION_NOT_ALLOWED:       &AppError{Code: "ACTION_NOT_ALLOWED", Status: 403},
	PROCESS_ALREADY_RUNNING:  &AppError{Code: "PROCESS_ALREADY_RUNNING", Status: 409},
	PROCESS_ALREADY_STOPPED:  &AppError{Code: "PROCESS_ALREADY_STOPPED", Status: 409},
	MALICIOUS_INPUT_DETECTED: &AppError{Code: "MALICIOUS_INPUT_DETECTED", Status: 400},
	FAIL2BAN_JAIL_NOT_FOUND:  &AppError{Code: "FAIL2BAN_JAIL_NOT_FOUND", Status: 404},
	FAIL2BAN_IP_NOT_BANNED:   &AppError{Code: "FAIL2BAN_IP_NOT_BANNED", Status: 404},
	FAIL2BAN_EXECUTION_ERROR: &AppError{Code: "FAIL2BAN_EXECUTION_ERROR", Status: 500},
}

var log *zap.Logger

type yamlWrapper struct {
	Errors map[string]struct {
		Status  int    `yaml:"status"`
		Message string `yaml:"message"`
	} `yaml:"errors"`
}

func Init(
	data []byte,
	logger *zap.Logger,
) error {
	log = logger

	var config yamlWrapper
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("could not parse error config: %w", err)
	}

	val := reflect.ValueOf(Errors).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		if entry, ok := config.Errors[fieldName]; ok {
			field := val.Field(i)
			appErr := field.Interface().(*AppError)
			appErr.Message = entry.Message
			appErr.Status = entry.Status
		} else {
			log.Warn("Error code defined in Go but missing in YAML", zap.String("code", fieldName))
		}
	}
	return nil
}
