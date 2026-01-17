package vps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

type BaseVpsService struct{}

func NewBaseVpsService() *BaseVpsService {
	return &BaseVpsService{}
}

func (s *BaseVpsService) RunScript(
	script string,
	target any,
) error {
	cmd := exec.Command("bash", "-c", script)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execution error: %s", errOut.String())
	}
	if out.Len() == 0 {
		return nil
	}
	return json.Unmarshal(out.Bytes(), target)
}

func (s *BaseVpsService) ExecuteSimple(
	name string,
	args ...string,
) error {
	cmd := exec.Command(name, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %s %v failed: %w", name, args, err)
	}
	return nil
}
