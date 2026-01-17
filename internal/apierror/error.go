package apierror

import "fmt"

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Meta    any    `json:"meta,omitempty"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) WithMeta(meta any) *AppError {
	newErr := *e
	newErr.Meta = meta
	return &newErr
}

func (e *AppError) Wrap(err error) *AppError {
	newErr := *e
	newErr.Err = err
	return &newErr
}
