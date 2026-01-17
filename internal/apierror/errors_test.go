package apierror

import (
	"testing"

	"go.uber.org/zap"
)

func TestInit_ValidYAML(t *testing.T) {
	yamlData := `
errors:
  INTERNAL_ERROR:
    status: 500
    message: "Custom server error"
  INVALID_REQUEST:
    status: 400
    message: "Custom bad request"
`
	origMsg := Errors.INTERNAL_ERROR.Message
	defer func() {
		Errors.INTERNAL_ERROR.Message = origMsg
	}()

	err := Init([]byte(yamlData), zap.NewNop())
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if Errors.INTERNAL_ERROR.Message != "Custom server error" {
		t.Errorf("Message = %q, want %q", Errors.INTERNAL_ERROR.Message, "Custom server error")
	}

	if Errors.INTERNAL_ERROR.Status != 500 {
		t.Errorf("Status = %d, want 500", Errors.INTERNAL_ERROR.Status)
	}
}

func TestInit_InvalidYAML(t *testing.T) {
	invalidYaml := "invalid: yaml: content: [[["

	err := Init([]byte(invalidYaml), zap.NewNop())
	if err == nil {
		t.Error("expected error for invalid YAML content")
	}
}

func TestInit_PartialYAML(t *testing.T) {
	yamlData := `
errors:
  RATE_LIMIT_EXCEEDED:
    status: 429
    message: "Slow down!"
`
	err := Init([]byte(yamlData), zap.NewNop())
	if err != nil {
		t.Errorf("partial YAML should not fail: %v", err)
	}

	if Errors.RATE_LIMIT_EXCEEDED.Message != "Slow down!" {
		t.Errorf("Expected updated message, got %q", Errors.RATE_LIMIT_EXCEEDED.Message)
	}
}

func TestAppError_Chaining(t *testing.T) {
	original := Errors.PERMISSION_DENIED

	_ = original.WithMeta("some-meta").Wrap(nil)

	if original.Meta != nil {
		t.Error("Registry error Meta was modified during chaining")
	}
}

func TestErrorRegistry_Defaults(t *testing.T) {
	if Errors.INTERNAL_ERROR == nil || Errors.MALICIOUS_INPUT_DETECTED == nil {
		t.Fatal("Some error definitions in registry are nil")
	}

	if Errors.INTERNAL_ERROR.Code != "INTERNAL_ERROR" {
		t.Errorf("Expected code INTERNAL_ERROR, got %q", Errors.INTERNAL_ERROR.Code)
	}
}
