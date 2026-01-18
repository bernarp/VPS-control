package pm2

type ProcessLister interface {
	GetProcessesBasic() (ProcessBasicGrouped, error)
	GetProcessesWithCwd() (ProcessWithCwdGrouped, error)
	GetProcessesFull() (ProcessFullGrouped, error)
}
type ProcessController interface {
	Restart(target string) (string, error)
	Start(target string) (string, error)
	Stop(target string) (string, error)
	executeAction(
		action Action,
		target string,
	) (string, error)
}
