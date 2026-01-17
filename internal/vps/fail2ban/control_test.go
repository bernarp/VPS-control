package fail2ban

import (
	"regexp"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestControlService_parseIntField(t *testing.T) {
	s := &ControlService{
		logger: zap.NewNop(),
	}

	tests := []struct {
		name    string
		input   string
		pattern string
		want    int
	}{
		{
			name:    "currently failed",
			input:   "|- Currently failed:\t5\n",
			pattern: ReCurrentlyFailed,
			want:    5,
		},
		{
			name:    "total failed",
			input:   "|- Total failed:\t100\n",
			pattern: ReTotalFailed,
			want:    100,
		},
		{
			name:    "currently banned",
			input:   "|- Currently banned:\t25\n",
			pattern: ReCurrentlyBanned,
			want:    25,
		},
		{
			name:    "total banned",
			input:   "|- Total banned:\t259\n",
			pattern: ReTotalBanned,
			want:    259,
		},
		{
			name:    "no match",
			input:   "some random text",
			pattern: ReCurrentlyFailed,
			want:    0,
		},
		{
			name:    "empty input",
			input:   "",
			pattern: ReCurrentlyFailed,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := s.parseIntField(tt.input, tt.pattern)
				if got != tt.want {
					t.Errorf("parseIntField() = %d, want %d", got, tt.want)
				}
			},
		)
	}
}

func TestParseGlobalStatusOutput(t *testing.T) {
	output := "Status\n|- Number of jail:      3\n`- Jail list:   sshd, nginx-forbidden, postfix"

	s := &ControlService{
		logger: zap.NewNop(),
	}

	jailListRegex := regexp.MustCompile(ReJailList)
	matches := jailListRegex.FindStringSubmatch(output)
	if len(matches) < 2 {
		t.Error("expected jail list to be parsed")
		return
	}

	jails := strings.Split(matches[1], ",")
	if len(jails) != 3 {
		t.Errorf("expected 3 jails, got %d", len(jails))
	}

	_ = s
}

func TestParseJailDetailsOutput(t *testing.T) {
	output := "Status for the jail: sshd\n" +
		"|- Filter\n" +
		"|  |- Currently failed: 2\n" +
		"|  |- Total failed:     284\n" +
		"|  `- File list:        /var/log/auth.log\n" +
		"`- Actions\n" +
		"   |- Currently banned: 10\n" +
		"   |- Total banned:     259\n" +
		"   `- Banned IP list:   1.2.3.4 5.6.7.8 9.10.11.12"

	s := &ControlService{
		logger: zap.NewNop(),
	}

	currentlyFailed := s.parseIntField(output, ReCurrentlyFailed)
	if currentlyFailed != 2 {
		t.Errorf("CurrentlyFailed = %d, want 2", currentlyFailed)
	}

	totalFailed := s.parseIntField(output, ReTotalFailed)
	if totalFailed != 284 {
		t.Errorf("TotalFailed = %d, want 284", totalFailed)
	}

	currentlyBanned := s.parseIntField(output, ReCurrentlyBanned)
	if currentlyBanned != 10 {
		t.Errorf("CurrentlyBanned = %d, want 10", currentlyBanned)
	}

	totalBanned := s.parseIntField(output, ReTotalBanned)
	if totalBanned != 259 {
		t.Errorf("TotalBanned = %d, want 259", totalBanned)
	}
}

func TestParseBannedIPList(t *testing.T) {
	output := "   `- Banned IP list:   1.2.3.4 5.6.7.8 9.10.11.12"

	ipListRegex := regexp.MustCompile(ReBannedIPList)
	matches := ipListRegex.FindStringSubmatch(output)
	if len(matches) < 2 {
		t.Fatal("expected IP list to be parsed")
	}

	rawIps := strings.TrimSpace(matches[1])
	ips := strings.Fields(rawIps)

	if len(ips) != 3 {
		t.Errorf("expected 3 IPs, got %d", len(ips))
	}

	expectedIPs := []string{"1.2.3.4", "5.6.7.8", "9.10.11.12"}
	for i, ip := range expectedIPs {
		if ips[i] != ip {
			t.Errorf("IP[%d] = %q, want %q", i, ips[i], ip)
		}
	}
}

func TestParseBannedIPList_Empty(t *testing.T) {
	output := "   `- Banned IP list:   "

	ipListRegex := regexp.MustCompile(ReBannedIPList)
	matches := ipListRegex.FindStringSubmatch(output)
	if len(matches) < 2 {
		t.Fatal("expected IP list regex to match")
	}

	rawIps := strings.TrimSpace(matches[1])
	var ips []string
	if rawIps != "" {
		ips = strings.Fields(rawIps)
	}

	if len(ips) != 0 {
		t.Errorf("expected 0 IPs for empty list, got %d", len(ips))
	}
}

func TestJailDetailsDTO_Structure(t *testing.T) {
	dto := JailDetailsDTO{
		JailName:        "sshd",
		CurrentlyFailed: 5,
		TotalFailed:     100,
		CurrentlyBanned: 10,
		TotalBanned:     50,
		BannedIPList:    []string{"1.2.3.4", "5.6.7.8"},
	}

	if dto.JailName != "sshd" {
		t.Errorf("JailName = %q, want %q", dto.JailName, "sshd")
	}

	if len(dto.BannedIPList) != 2 {
		t.Errorf("BannedIPList length = %d, want 2", len(dto.BannedIPList))
	}
}

func TestFail2BanStatusDTO_Structure(t *testing.T) {
	dto := Fail2BanStatusDTO{
		JailCount: 3,
		JailList:  []string{"sshd", "nginx", "postfix"},
	}

	if dto.JailCount != 3 {
		t.Errorf("JailCount = %d, want 3", dto.JailCount)
	}

	if len(dto.JailList) != 3 {
		t.Errorf("JailList length = %d, want 3", len(dto.JailList))
	}
}

func TestErrorOutputPatterns(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		contains string
		expected bool
	}{
		{
			name:     "jail does not exist",
			output:   "Sorry but the jail 'invalid' Does not exist",
			contains: ErrOutputDoesNotExist,
			expected: true,
		},
		{
			name:     "jail not found",
			output:   "ERROR: Jail 'test' not found",
			contains: ErrOutputNotFound,
			expected: true,
		},
		{
			name:     "ip is not banned",
			output:   "1.2.3.4 is not banned",
			contains: ErrOutputIsNotBanned,
			expected: true,
		},
		{
			name:     "jail not found for unban",
			output:   "ERROR: Jail not found",
			contains: ErrOutputJailNotFound,
			expected: true,
		},
		{
			name:     "success output",
			output:   "1.2.3.4",
			contains: ErrOutputDoesNotExist,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := strings.Contains(tt.output, tt.contains)
				if result != tt.expected {
					t.Errorf(
						"Contains(%q, %q) = %v, want %v",
						tt.output, tt.contains, result, tt.expected,
					)
				}
			},
		)
	}
}

func TestConstants(t *testing.T) {
	if CmdSudo != "sudo" {
		t.Errorf("CmdSudo = %q, want %q", CmdSudo, "sudo")
	}

	if CmdFail2Ban != "fail2ban-client" {
		t.Errorf("CmdFail2Ban = %q, want %q", CmdFail2Ban, "fail2ban-client")
	}

	if ArgStatus != "status" {
		t.Errorf("ArgStatus = %q, want %q", ArgStatus, "status")
	}

	if ArgSet != "set" {
		t.Errorf("ArgSet = %q, want %q", ArgSet, "set")
	}

	if ArgUnbanIP != "unbanip" {
		t.Errorf("ArgUnbanIP = %q, want %q", ArgUnbanIP, "unbanip")
	}

	if ParamJailName != "name" {
		t.Errorf("ParamJailName = %q, want %q", ParamJailName, "name")
	}
}

func TestBanActionResponse_Structure(t *testing.T) {
	resp := BanActionResponse{
		Success: true,
		Message: "IP unbanned successfully",
	}

	if !resp.Success {
		t.Error("Success should be true")
	}

	if resp.Message != "IP unbanned successfully" {
		t.Errorf("Message = %q, want %q", resp.Message, "IP unbanned successfully")
	}
}

func TestBanActionRequest_Structure(t *testing.T) {
	req := BanActionRequest{
		Jail: "sshd",
		IP:   "192.168.1.100",
	}

	if req.Jail != "sshd" {
		t.Errorf("Jail = %q, want %q", req.Jail, "sshd")
	}

	if req.IP != "192.168.1.100" {
		t.Errorf("IP = %q, want %q", req.IP, "192.168.1.100")
	}
}
