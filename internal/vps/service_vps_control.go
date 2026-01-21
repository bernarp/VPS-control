package vps

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

type BaseVpsService struct {
	Timeout time.Duration
}

func NewBaseVpsService() *BaseVpsService {
	return &BaseVpsService{
		Timeout: DefaultTimeout,
	}
}

func NewBaseVpsServiceWithTimeout(timeout time.Duration) *BaseVpsService {
	return &BaseVpsService{
		Timeout: timeout,
	}
}

func (s *BaseVpsService) RunScript(
	script string,
	target any,
) error {
	return s.RunScriptWithContext(context.Background(), script, target)
}

func (s *BaseVpsService) ExecuteSimple(
	name string,
	args ...string,
) error {
	return s.ExecuteWithContext(context.Background(), name, args...)
}

func (s *BaseVpsService) RunScriptWithContext(
	ctx context.Context,
	script string,
	target any,
) error {
	ctx, cancel := s.contextWithTimeout(ctx)
	defer cancel()

	cmd := exec.CommandContext(ctx, shellName, shellFlag, script)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return s.wrapError(opScriptExecution, err, ctx.Err(), stderr.String())
	}

	if stdout.Len() == 0 {
		return nil
	}

	if err := json.Unmarshal(stdout.Bytes(), target); err != nil {
		return fmt.Errorf(errJsonUnmarshal, err, truncate(stdout.String(), 200))
	}

	return nil
}

func (s *BaseVpsService) ExecuteWithContext(
	ctx context.Context,
	name string,
	args ...string,
) error {
	ctx, cancel := s.contextWithTimeout(ctx)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return s.wrapError(fmt.Sprintf(opCommand, name), err, ctx.Err(), stderr.String())
	}

	return nil
}

func (s *BaseVpsService) contextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if s.Timeout > 0 {
		return context.WithTimeout(ctx, s.Timeout)
	}
	return context.WithCancel(ctx)
}

func (s *BaseVpsService) wrapError(
	operation string,
	err, ctxErr error,
	stderr string,
) error {
	if errors.Is(ctxErr, context.DeadlineExceeded) {
		return fmt.Errorf(errTimeout, operation, s.Timeout)
	}
	if errors.Is(ctxErr, context.Canceled) {
		return fmt.Errorf(errCancelled, operation)
	}
	if stderr != "" {
		return fmt.Errorf(errWithStderr, operation, err, stderr)
	}
	return fmt.Errorf(errGeneric, operation, err)
}

func truncate(
	s string,
	maxLen int,
) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + truncateSuffix
}
