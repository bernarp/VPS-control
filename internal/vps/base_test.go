package vps

import (
	"encoding/json"
	"runtime"
	"testing"
)

func TestBaseVpsService_Creation(t *testing.T) {
	svc := NewBaseVpsService()
	if svc == nil {
		t.Fatal("NewBaseVpsService() returned nil")
	}
}

func TestBaseVpsService_RunScript_EmptyOutput(t *testing.T) {
	svc := NewBaseVpsService()

	var result interface{}
	err := svc.RunScript(":", &result)
	if err != nil {
		t.Errorf("RunScript with empty output should not error: %v", err)
	}
}

func TestBaseVpsService_RunScript_ValidJSON(t *testing.T) {
	svc := NewBaseVpsService()

	type TestResult struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	var result TestResult
	err := svc.RunScript(`echo '{"name":"test","value":42}'`, &result)
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	if result.Name != "test" {
		t.Errorf("Name = %q, want %q", result.Name, "test")
	}

	if result.Value != 42 {
		t.Errorf("Value = %d, want 42", result.Value)
	}
}

func TestBaseVpsService_RunScript_InvalidJSON(t *testing.T) {
	svc := NewBaseVpsService()

	var result map[string]interface{}
	err := svc.RunScript(`echo 'not valid json'`, &result)
	if err == nil {
		t.Error("RunScript should fail for invalid JSON output")
	}
}

func TestBaseVpsService_RunScript_CommandError(t *testing.T) {
	svc := NewBaseVpsService()

	var result interface{}
	err := svc.RunScript("exit 1", &result)
	if err == nil {
		t.Error("RunScript should fail when command exits with error")
	}
}

func TestBaseVpsService_RunScript_CommandNotFound(t *testing.T) {
	svc := NewBaseVpsService()

	var result interface{}
	err := svc.RunScript("nonexistentcommand12345", &result)
	if err == nil {
		t.Error("RunScript should fail for non-existent command")
	}
}

func TestBaseVpsService_ExecuteSimple_Success(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows: 'true' command not available")
	}

	svc := NewBaseVpsService()
	err := svc.ExecuteSimple("true")
	if err != nil {
		t.Errorf("ExecuteSimple(true) failed: %v", err)
	}
}

func TestBaseVpsService_ExecuteSimple_Success_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows")
	}

	svc := NewBaseVpsService()
	// cmd /c exit 0 - успешное завершение на Windows
	err := svc.ExecuteSimple("cmd", "/c", "exit", "0")
	if err != nil {
		t.Errorf("ExecuteSimple(cmd /c exit 0) failed: %v", err)
	}
}

func TestBaseVpsService_ExecuteSimple_WithArgs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows: 'echo' as standalone command not available")
	}

	svc := NewBaseVpsService()
	err := svc.ExecuteSimple("echo", "hello", "world")
	if err != nil {
		t.Errorf("ExecuteSimple(echo, args) failed: %v", err)
	}
}

func TestBaseVpsService_ExecuteSimple_WithArgs_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping on non-Windows")
	}

	svc := NewBaseVpsService()
	err := svc.ExecuteSimple("cmd", "/c", "echo", "hello", "world")
	if err != nil {
		t.Errorf("ExecuteSimple(cmd /c echo) failed: %v", err)
	}
}

func TestBaseVpsService_ExecuteSimple_CommandFailed(t *testing.T) {
	svc := NewBaseVpsService()

	var err error
	if runtime.GOOS == "windows" {
		err = svc.ExecuteSimple("cmd", "/c", "exit", "1")
	} else {
		err = svc.ExecuteSimple("false")
	}

	if err == nil {
		t.Error("ExecuteSimple should return error for failed command")
	}
}

func TestBaseVpsService_ExecuteSimple_CommandNotFound(t *testing.T) {
	svc := NewBaseVpsService()

	err := svc.ExecuteSimple("nonexistentcommand12345")
	if err == nil {
		t.Error("ExecuteSimple should fail for non-existent command")
	}
}

func TestBaseVpsService_RunScript_ArrayOutput(t *testing.T) {
	svc := NewBaseVpsService()

	var result []string
	err := svc.RunScript(`echo '["a","b","c"]'`, &result)
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("result length = %d, want 3", len(result))
	}

	expected := []string{"a", "b", "c"}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("result[%d] = %q, want %q", i, result[i], v)
		}
	}
}

func TestBaseVpsService_RunScript_MapOutput(t *testing.T) {
	svc := NewBaseVpsService()

	var result map[string][]map[string]interface{}
	script := `echo '{"group1":[{"name":"proc1","pid":123}]}'`
	err := svc.RunScript(script, &result)
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	group, ok := result["group1"]
	if !ok {
		t.Fatal("expected 'group1' key in result")
	}

	if len(group) != 1 {
		t.Errorf("group1 length = %d, want 1", len(group))
	}

	if group[0]["name"] != "proc1" {
		t.Errorf("name = %v, want 'proc1'", group[0]["name"])
	}
}

func TestJSONUnmarshalBehavior(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		valid bool
	}{
		{"valid json object", []byte(`{"key":"value"}`), true},
		{"valid json array", []byte(`[1,2,3]`), true},
		{"empty object", []byte(`{}`), true},
		{"empty array", []byte(`[]`), true},
		{"null", []byte(`null`), true},
		{"invalid json", []byte(`not json`), false},
		{"empty bytes", []byte(``), false},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				var result interface{}
				err := json.Unmarshal(tt.input, &result)
				if tt.valid && err != nil {
					t.Errorf("expected valid JSON, got error: %v", err)
				}
				if !tt.valid && err == nil {
					t.Error("expected error for invalid JSON")
				}
			},
		)
	}
}

func TestBaseVpsService_RunScript_NestedJSON(t *testing.T) {
	svc := NewBaseVpsService()

	type Nested struct {
		Inner struct {
			Value string `json:"value"`
		} `json:"inner"`
	}

	var result Nested
	script := `echo '{"inner":{"value":"deep"}}'`
	err := svc.RunScript(script, &result)
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	if result.Inner.Value != "deep" {
		t.Errorf("Inner.Value = %q, want %q", result.Inner.Value, "deep")
	}
}
