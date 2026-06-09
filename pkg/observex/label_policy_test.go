package observex

import (
	"testing"
)

func TestDefaultLabelPolicy(t *testing.T) {
	p := NewDefaultLabelPolicy()

	// Denied labels must be rejected.
	for _, name := range DefaultDeniedLabels {
		if err := p.ValidateLabel(name); err == nil {
			t.Errorf("ValidateLabel(%q) expected error for denied label, got nil", name)
		}
	}

	// Valid labels must pass.
	validLabels := []string{"operation", "status", "source", "error_kind", "module"}
	for _, name := range validLabels {
		if err := p.ValidateLabel(name); err != nil {
			t.Errorf("ValidateLabel(%q) unexpected error: %v", name, err)
		}
	}
}

func TestLabelPolicyValidateLabel(t *testing.T) {
	p := NewDefaultLabelPolicy()

	tests := []struct {
		name    string
		label   string
		wantErr bool
	}{
		{"valid", "operation", false},
		{"denied_error", "error", true},
		{"denied_err", "err", true},
		{"denied_msg", "msg", true},
		{"denied_level", "level", true},
		{"empty", "", true},
		{"bad_case", "CamelCase", true},
		{"with_dot", "my.label", true},
		{"snake_case", "my_label", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.ValidateLabel(tt.label)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLabel(%q) error=%v, wantErr=%v", tt.label, err, tt.wantErr)
			}
		})
	}
}

func TestLabelPolicyValidateLabels(t *testing.T) {
	p := NewDefaultLabelPolicy()

	names := []string{"operation", "error", "status", "err"}
	errs := p.ValidateLabels(names)
	if len(errs) != 2 {
		t.Errorf("ValidateLabels expected 2 errors, got %d", len(errs))
	}
}

func TestLabelPolicyAllowedList(t *testing.T) {
	p := NewLabelPolicy([]string{"status", "source"}, nil)

	if err := p.ValidateLabel("status"); err != nil {
		t.Errorf("ValidateLabel(status) unexpected error: %v", err)
	}
	if err := p.ValidateLabel("custom_label"); err == nil {
		t.Errorf("ValidateLabel(custom_label) expected error for not-in-allowed, got nil")
	}
}

func TestLabelPolicyCustomDenied(t *testing.T) {
	p := NewLabelPolicy(nil, []string{"forbidden", "banned"})

	if err := p.ValidateLabel("forbidden"); err == nil {
		t.Errorf("ValidateLabel(forbidden) expected error, got nil")
	}
	if err := p.ValidateLabel("ok_label"); err != nil {
		t.Errorf("ValidateLabel(ok_label) unexpected error: %v", err)
	}
}
