package kubernetes

import (
	"strings"
	"testing"
)

func TestValidateDeviceID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
		errMsg  string // substring expected in error message
	}{
		// --- valid IDs ---
		{name: "valid hyphenated", id: "dev-01", wantErr: false},
		{name: "valid single alpha", id: "a", wantErr: false},
		{name: "valid single digit", id: "0", wantErr: false},
		{name: "valid alphanumeric", id: "abc123", wantErr: false},
		{name: "valid multi-hyphen", id: "a-b-c", wantErr: false},

		// --- empty ---
		{name: "empty string", id: "", wantErr: true, errMsg: "cannot be empty"},

		// --- too long ---
		{name: "exactly 240 chars", id: strings.Repeat("a", 240), wantErr: false},
		{name: "241 chars", id: strings.Repeat("a", 241), wantErr: true, errMsg: "too long"},
		{name: "300 chars", id: strings.Repeat("b", 300), wantErr: true, errMsg: "too long"},

		// --- uppercase ---
		{name: "uppercase first char", id: "Dev-01", wantErr: true, errMsg: "must be lowercase"},
		{name: "all uppercase", id: "DEV", wantErr: true, errMsg: "must be lowercase"},

		// --- invalid characters ---
		{name: "underscore", id: "dev_01", wantErr: true, errMsg: "invalid characters"},
		{name: "dot", id: "dev.01", wantErr: true, errMsg: "invalid characters"},
		{name: "space", id: "dev 01", wantErr: true, errMsg: "invalid characters"},
		{name: "at sign", id: "dev@01", wantErr: true, errMsg: "invalid characters"},

		// --- leading/trailing hyphen ---
		{name: "leading hyphen", id: "-dev", wantErr: true, errMsg: "invalid characters"},
		{name: "trailing hyphen", id: "dev-", wantErr: true, errMsg: "invalid characters"},
		{name: "only hyphen", id: "-", wantErr: true, errMsg: "invalid characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeviceID(tt.id)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
