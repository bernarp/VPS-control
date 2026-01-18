package pm2

import (
	"VPS-control/internal/vps"
)

var _ ProcessLister = (*ListService)(nil)

// ListService provides lightweight PM2 process information retrieval.
// Instead of using 'pm2 jlist' which returns heavy JSON objects with full process metadata,
// this service directly reads PM2 PID files from ~/.pm2/pids/*.pid and queries /proc filesystem.
// This approach is significantly faster and reduces memory overhead.
type ListService struct {
	base *vps.BaseVpsService
}

func NewListService(base *vps.BaseVpsService) *ListService {
	return &ListService{base: base}
}

// GetProcessesBasic retrieves minimal process information (name, PID, active status).
// Uses direct file system reads instead of PM2 API for better performance.
// Groups processes by parent PID (PPID) to maintain PM2 cluster structure.
func (s *ListService) GetProcessesBasic() (ProcessBasicGrouped, error) {
	// Bash script that:
	// 1. Iterates through ~/.pm2/pids/*.pid files
	// 2. Reads PID from each file
	// 3. Gets parent PID from /proc
	// 4. Extracts process name from filename (removes -N.pid suffix)
	// 5. Checks if process is active by verifying /proc/$pid exists
	// 6. Groups results by PPID and outputs as JSON
	script := `unset g; declare -A g; for f in ~/.pm2/pids/*.pid; do pid=$(cat "$f"); ppid=$(ps -o ppid= -p "$pid" 2>/dev/null | tr -d ' '); name=$(basename "$f" | sed 's/-[0-9]*\.pid//'); [ -d /proc/$pid ] && active=true || active=false; g[$ppid]+="${g[$ppid]:+,}{\"name\":\"$name\",\"pid\":$pid,\"active\":$active}"; done; echo "{"; f=1; for pp in "${!g[@]}"; do [ $f -eq 1 ] || echo ","; f=0; printf '"%s":[%s]' "$pp" "${g[$pp]}"; done; echo "}"`

	var result ProcessBasicGrouped
	if err := s.base.RunScript(script, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetProcessesWithCwd extends GetProcessesBasic by adding current working directory.
// Useful for identifying which project/folder each process is running from.
// Still much faster than 'pm2 jlist' as it only reads specific /proc entries.
func (s *ListService) GetProcessesWithCwd() (ProcessWithCwdGrouped, error) {
	// Extended version that also reads /proc/$pid/cwd symlink
	// to determine the working directory of each process
	script := `unset g; declare -A g; for f in ~/.pm2/pids/*.pid; do pid=$(cat "$f"); ppid=$(ps -o ppid= -p "$pid" 2>/dev/null | tr -d ' '); name=$(basename "$f" | sed 's/-[0-9]*\.pid//'); cwd=$(readlink /proc/$pid/cwd 2>/dev/null); [ -d /proc/$pid ] && active=true || active=false; g[$ppid]+="${g[$ppid]:+,}{\"name\":\"$name\",\"pid\":$pid,\"cwd\":\"$cwd\",\"active\":$active}"; done; echo "{"; f=1; for pp in "${!g[@]}"; do [ $f -eq 1 ] || echo ","; f=0; printf '"%s":[%s]' "$pp" "${g[$pp]}"; done; echo "}"`

	var result ProcessWithCwdGrouped
	if err := s.base.RunScript(script, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetProcessesFull retrieves complete process information including resource usage.
// Adds memory %, CPU %, and start time by querying 'ps' and /proc/stat.
// This is the most comprehensive option but still faster than PM2's native methods
// because it avoids deserializing PM2's internal state and large metadata objects.
func (s *ListService) GetProcessesFull() (ProcessFullGrouped, error) {
	// Full version that additionally queries:
	// - Memory usage (%) via ps -o %mem
	// - CPU usage (%) via ps -o %cpu
	// - Process start time by reading /proc/$pid stat file modification time
	// All data is aggregated in a single pass for efficiency
	script := `unset g; declare -A g; for f in ~/.pm2/pids/*.pid; do pid=$(cat "$f"); stats=$(ps -o ppid=,%mem=,%cpu= -p "$pid" 2>/dev/null); ppid=$(echo $stats | awk '{print $1}'); mem=$(echo $stats | awk '{print $2}'); cpu=$(echo $stats | awk '{print $3}'); name=$(basename "$f" | sed 's/-[0-9]*\.pid//'); cwd=$(readlink /proc/$pid/cwd 2>/dev/null); start=$(date -d @$(stat -c %Y /proc/$pid 2>/dev/null) -Iseconds 2>/dev/null); [ -d /proc/$pid ] && active=true || active=false; g[$ppid]+="${g[$ppid]:+,}{\"name\":\"$name\",\"pid\":$pid,\"cwd\":\"$cwd\",\"mem\":$mem,\"cpu\":$cpu,\"started_at\":\"$start\",\"active\":$active}"; done; echo "{"; f=1; for pp in "${!g[@]}"; do [ $f -eq 1 ] || echo ","; f=0; printf '"%s":[%s]' "$pp" "${g[$pp]}"; done; echo "}"`

	var result ProcessFullGrouped
	if err := s.base.RunScript(script, &result); err != nil {
		return nil, err
	}
	return result, nil
}
