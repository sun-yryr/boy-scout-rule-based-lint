package cmd

import (
	"fmt"

	"github.com/sun-yryr/boy-scout-rule-based-lint/internal/baseline"
)

var validPolicies = map[string]bool{
	"off":   true,
	"file":  true,
	"hunk":  true,
	"scope": true,
}

func validatePolicy(policy string) error {
	if !validPolicies[policy] {
		return fmt.Errorf("invalid boy scout policy %q: valid values are off, file, hunk, scope", policy)
	}
	if policy == "scope" {
		return fmt.Errorf("boy scout policy scope is not yet available")
	}
	return nil
}

func resolveCheckPolicy(cliPolicy string, policyChanged bool, bl *baseline.Baseline) (string, error) {
	if policyChanged {
		if err := validatePolicy(cliPolicy); err != nil {
			return "", err
		}
		return cliPolicy, nil
	}

	if bl.Config != nil && bl.Config.BoyScoutPolicy != "" {
		if err := validatePolicy(bl.Config.BoyScoutPolicy); err != nil {
			return "", err
		}
		return bl.Config.BoyScoutPolicy, nil
	}

	return "off", nil
}

func resolveCheckBaseRef(cliBaseRef string, baseRefChanged bool, bl *baseline.Baseline) string {
	if baseRefChanged {
		return cliBaseRef
	}
	if bl.Config != nil {
		return bl.Config.BaseRef
	}
	return ""
}
