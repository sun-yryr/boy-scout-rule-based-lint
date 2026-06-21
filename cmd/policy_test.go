package cmd

import (
	"testing"

	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/baseline"
)

func TestValidatePolicy(t *testing.T) {
	tests := []struct {
		name    string
		policy  string
		wantErr bool
	}{
		{name: "off", policy: "off"},
		{name: "file", policy: "file"},
		{name: "hunk", policy: "hunk"},
		{name: "invalid", policy: "bogus", wantErr: true},
		{name: "scope unavailable", policy: "scope", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePolicy(tt.policy)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validatePolicy(%q) err = %v, wantErr %v", tt.policy, err, tt.wantErr)
			}
		})
	}
}

func TestResolveCheckPolicy_DefaultOffWithoutConfig(t *testing.T) {
	bl := baseline.New()

	policy, err := resolveCheckPolicy("off", false, bl)
	if err != nil {
		t.Fatalf("resolveCheckPolicy() err = %v", err)
	}
	if policy != "off" {
		t.Fatalf("resolveCheckPolicy() = %q, want off", policy)
	}
}

func TestResolveCheckPolicy_InvalidBaselinePolicy(t *testing.T) {
	bl := &baseline.Baseline{
		Config: &baseline.Config{BoyScoutPolicy: "bogus"},
	}

	_, err := resolveCheckPolicy("off", false, bl)
	if err == nil {
		t.Fatal("resolveCheckPolicy() err = nil, want error")
	}
}
