package model

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		rule      *Trigger
		wantError error
	}{
		{
			name: "PropertyPath with empty value should error",
			rule: &Trigger{
				Property: PropertyPath, Value: PropertyValue{Value: ""},
			},
			wantError: ErrEmptyRuleValue,
		},
		{
			name: "PropertyHost with valid value should pass",
			rule: &Trigger{
				Property: PropertyHost, Value: PropertyValue{Value: "validValue"},
			},
			wantError: nil,
		},
		{
			name: "PropertyMethod with empty value should error",
			rule: &Trigger{
				Property: PropertyMethod, Value: PropertyValue{Value: ""},
			},
			wantError: ErrEmptyRuleValue,
		},
		{
			name: "PropertyHeader with empty key should error",
			rule: &Trigger{
				Property: PropertyHeader, Value: PropertyValue{Key: "", Value: "validValue"},
			},
			wantError: ErrEmptyRuleValue,
		},
		{
			name: "PropertyHeader with unsupported comparator should error",
			rule: &Trigger{
				Property: PropertyHeader, Value: PropertyValue{Key: "validKey", Value: "validValue"},
				Comparator: "InvalidComparator"},
			wantError: fmt.Errorf("unsupported Comparator for property %s", PropertyHeader),
		},
		{
			name: "PropertyQuery with empty value should error",
			rule: &Trigger{
				Property: PropertyQuery, Value: PropertyValue{Key: "validKey", Value: ""},
			},
			wantError: ErrEmptyRuleValue,
		},
		{
			name: "PropertyQuery with valid key and value but unsupported comparator should error",
			rule: &Trigger{
				Property: PropertyQuery, Value: PropertyValue{Key: "validKey", Value: "validValue"},
				Comparator: "InvalidComparator"},
			wantError: fmt.Errorf("unsupported Comparator for property %s", PropertyQuery),
		},
		// You can add more test cases if needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			assert.Equal(t, tt.wantError, err)
		})
	}
}

func TestRule_Match(t *testing.T) {
	rule := &Trigger{
		Property:   PropertyPath,
		Value:      PropertyValue{Value: "/testpath"},
		comparator: &ComparatorEqual{},
	}

	req := &http.Request{URL: &url.URL{Path: "/testpath"}}
	result, err := rule.Match(req)
	assert.NoError(t, err)
	assert.True(t, result)
}
