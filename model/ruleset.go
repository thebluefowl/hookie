package model

import (
	"errors"
	"net/http"
)

type Ruleset struct {
	Name     string   `yaml:"name"`
	Rules    []Rule   `yaml:"rules"`
	Operator Operator `yaml:"operator"`
}

// Match checks if the given request matches the rules defined in the Ruleset.
// It uses the Operator to determine how rules should be combined (AND or OR logic).
func (rs *Ruleset) Match(req *http.Request) (bool, error) {
	var errs []error // Collects errors from rule matching

	switch rs.Operator {
	case OperatorOr:
		// For OR logic, only one rule needs to match.
		for _, r := range rs.Rules {
			result, err := r.Match(req)

			// If there's an error, collect it.
			if err != nil {
				errs = append(errs, err)
			}

			// If a rule matches, return true immediately.
			if result {
				return true, errors.Join(errs...)
			}
		}
		// If no rules matched, return false.
		return false, errors.Join(errs...)

	default:
		// Default to AND logic: all rules must match.
		for _, r := range rs.Rules {
			result, err := r.Match(req)

			// If there's an error, collect it.
			if err != nil {
				errs = append(errs, err)
			}

			// If a rule doesn't match, return false immediately.
			if !result {
				return false, errors.Join(errs...)
			}
		}
		// If all rules matched, return true.
		return true, errors.Join(errs...)
	}
}
