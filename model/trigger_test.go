package model

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
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

func TestRule_UnmarshalYAML(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		wantErr  bool
		wantType interface{}
	}{
		{
			name: "valid rule with equal comparator",
			data: []byte(`name: example-ruleset
rules:
  - name: rule1
    property: path
    comparator: equal
    value:
      value: value1
operator: AND`),
			wantErr:  false,
			wantType: &ComparatorEqual{},
		},
		{
			name: "valid rule with not_equal comparator",
			data: []byte(`name: example-ruleset
rules:
  - name: rule1
    property: path
    comparator: not_equal
    value:
      key: X-Header
      value: value1
operator: AND`),
			wantErr:  false,
			wantType: &ComparatorNotEqual{},
		},
		{
			name: "valid rule with not_equal comparator",
			data: []byte(`name: example-ruleset
rules:
  - name: rule1
    property: query
    comparator: contains
    value:
      key: X-Header
      value: value1
operator: AND`),
			wantErr:  false,
			wantType: &ComparatorContains{},
		},
		{
			name: "valid rule with not_equal comparator",
			data: []byte(`name: example-ruleset
rules:
  - name: rule1
    property: header
    comparator: not_contains
    value:
      key: X-Header
      value: value1
operator: AND`),
			wantErr:  false,
			wantType: &ComparatorNotContains{},
		},
		{
			name:    "valid rule with not_equal comparator",
			data:    []byte(`xxx---`),
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ruleset := &TriggerSet{}
			err := yaml.Unmarshal(tt.data, ruleset)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tt.wantType, ruleset.Triggers[0].comparator)
			}
		})
	}

}
func TestProperty_Value(t *testing.T) {
	tests := []struct {
		property Property
		req      *http.Request
		expected interface{}
	}{
		{
			property: PropertyPath,
			req:      &http.Request{URL: &url.URL{Path: "/testpath"}},
			expected: "/testpath",
		},
		{
			property: PropertyHost,
			req:      &http.Request{Host: "example.com"},
			expected: "example.com",
		},
		{
			property: PropertyMethod,
			req:      &http.Request{Method: http.MethodGet},
			expected: http.MethodGet,
		},
		{
			property: PropertyHeader,
			req: &http.Request{
				Header: http.Header{
					"Authorization": []string{"Bearer token"},
				},
			},
			expected: http.Header{
				"Authorization": []string{"Bearer token"},
			},
		},
		{
			property: PropertyQuery,
			req: &http.Request{
				URL: &url.URL{
					RawQuery: "key=value",
				},
			},
			expected: url.Values{
				"key": []string{"value"},
			},
		},
		{
			property: Property("xxx"), // An invalid property to test the default return case
			req:      &http.Request{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.property), func(t *testing.T) {
			got := tt.property.Value(tt.req)
			assert.Equal(t, tt.expected, got)
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
