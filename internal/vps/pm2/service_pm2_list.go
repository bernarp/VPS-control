package pm2

import (
	"DiscordBotControl/internal/vps"
)

type ListService struct {
	base *vps.BaseVpsService
}

func NewListService(base *vps.BaseVpsService) *ListService {
	return &ListService{base: base}
}

func (s *ListService) GetProcessesBasic() (ProcessBasicGrouped, error) {
	script := `unset g; declare -A g; for f in ~/.pm2/pids/*.pid; do pid=$(cat "$f"); ppid=$(ps -o ppid= -p "$pid" 2>/dev/null | tr -d ' '); name=$(basename "$f" | sed 's/-[0-9]*\.pid//'); [ -d /proc/$pid ] && active=true || active=false; g[$ppid]+="${g[$ppid]:+,}{\"name\":\"$name\",\"pid\":$pid,\"active\":$active}"; done; echo "{"; f=1; for pp in "${!g[@]}"; do [ $f -eq 1 ] || echo ","; f=0; printf '"%s":[%s]' "$pp" "${g[$pp]}"; done; echo "}"`

	var result ProcessBasicGrouped
	if err := s.base.RunScript(script, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ListService) GetProcessesWithCwd() (ProcessWithCwdGrouped, error) {
	script := `unset g; declare -A g; for f in ~/.pm2/pids/*.pid; do pid=$(cat "$f"); ppid=$(ps -o ppid= -p "$pid" 2>/dev/null | tr -d ' '); name=$(basename "$f" | sed 's/-[0-9]*\.pid//'); cwd=$(readlink /proc/$pid/cwd 2>/dev/null); [ -d /proc/$pid ] && active=true || active=false; g[$ppid]+="${g[$ppid]:+,}{\"name\":\"$name\",\"pid\":$pid,\"cwd\":\"$cwd\",\"active\":$active}"; done; echo "{"; f=1; for pp in "${!g[@]}"; do [ $f -eq 1 ] || echo ","; f=0; printf '"%s":[%s]' "$pp" "${g[$pp]}"; done; echo "}"`

	var result ProcessWithCwdGrouped
	if err := s.base.RunScript(script, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ListService) GetProcessesFull() (ProcessFullGrouped, error) {
	script := `unset g; declare -A g; for f in ~/.pm2/pids/*.pid; do pid=$(cat "$f"); stats=$(ps -o ppid=,%mem=,%cpu= -p "$pid" 2>/dev/null); ppid=$(echo $stats | awk '{print $1}'); mem=$(echo $stats | awk '{print $2}'); cpu=$(echo $stats | awk '{print $3}'); name=$(basename "$f" | sed 's/-[0-9]*\.pid//'); cwd=$(readlink /proc/$pid/cwd 2>/dev/null); start=$(date -d @$(stat -c %Y /proc/$pid 2>/dev/null) -Iseconds 2>/dev/null); [ -d /proc/$pid ] && active=true || active=false; g[$ppid]+="${g[$ppid]:+,}{\"name\":\"$name\",\"pid\":$pid,\"cwd\":\"$cwd\",\"mem\":$mem,\"cpu\":$cpu,\"started_at\":\"$start\",\"active\":$active}"; done; echo "{"; f=1; for pp in "${!g[@]}"; do [ $f -eq 1 ] || echo ","; f=0; printf '"%s":[%s]' "$pp" "${g[$pp]}"; done; echo "}"`

	var result ProcessFullGrouped
	if err := s.base.RunScript(script, &result); err != nil {
		return nil, err
	}
	return result, nil
}
