package pm2

import (
	"VPS-control/internal/apierror"
	"fmt"
	"os/exec"
	"strconv"
)

var _ ProcessController = (*ControlService)(nil)

type ControlService struct {
	listSvc ProcessLister
}

func NewControlService(listSvc ProcessLister) *ControlService {
	return &ControlService{
		listSvc: listSvc,
	}
}

func (s *ControlService) Restart(target string) (string, error) {
	return s.executeAction(ActionRestart, target)
}

func (s *ControlService) Start(target string) (string, error) {
	return s.executeAction(ActionStart, target)
}

func (s *ControlService) Stop(target string) (string, error) {
	return s.executeAction(ActionStop, target)
}

func (s *ControlService) executeAction(
	action Action,
	target string,
) (string, error) {
	processes, err := s.listSvc.GetProcessesBasic()
	if err != nil {
		return "", err
	}

	var targetProc *ProcessBasicDTO
	pid, isPidErr := strconv.Atoi(target)

	for _, group := range processes {
		for _, proc := range group {
			if (isPidErr == nil && proc.PID == pid) || proc.Name == target {
				targetProc = &proc
				break
			}
		}
	}

	if targetProc == nil {
		return "", apierror.Errors.PM2_PROCESS_NOT_FOUND
	}

	if action == ActionStart && targetProc.Active {
		return targetProc.Name, apierror.Errors.PROCESS_ALREADY_RUNNING
	}
	if action == ActionStop && !targetProc.Active {
		return targetProc.Name, apierror.Errors.PROCESS_ALREADY_STOPPED
	}

	// targetProc.Name приходит из pm2 jlist, а не от пользователя напрямую.
	// Процесс валидируется через GetProcessesBasic() — только существующие имена.
	cmd := exec.Command("pm2", string(action), targetProc.Name) //nolint:gosec // validated against pm2 process list
	if err := cmd.Run(); err != nil {
		return targetProc.Name, fmt.Errorf("pm2 execution failed: %w", err)
	}

	return targetProc.Name, nil
}
