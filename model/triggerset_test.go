package model

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestRuleset_Match(t *testing.T) {
	tests := []struct {
		name       string
		TriggerSet *TriggerSet
		req        *http.Request
		wantErr    bool
		wantResult bool
	}{
		{
			name: "AND match",
			TriggerSet: &TriggerSet{
				Triggers: []Trigger{
					{Property: PropertyPath, comparator: &ComparatorEqual{}, Value: PropertyValue{Value: "/test"}},
					{Property: PropertyMethod, comparator: &ComparatorEqual{}, Value: PropertyValue{Value: "GET"}},
				},
				Operator: OperatorAnd,
			},
			req:        &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "/test"}},
			wantErr:    false,
			wantResult: true,
		},
		{
			name: "AND no match",
			TriggerSet: &TriggerSet{
				Triggers: []Trigger{
					{Property: PropertyPath, comparator: &ComparatorEqual{}, Value: PropertyValue{Value: "/test"}},
					{Property: PropertyMethod, comparator: &ComparatorEqual{}, Value: PropertyValue{Value: "GET"}},
				},
				Operator: OperatorAnd,
			},
			req:        &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/test"}},
			wantErr:    false,
			wantResult: false,
		},
		{
			name: "OR match",
			TriggerSet: &TriggerSet{
				Triggers: []Trigger{
					{Property: PropertyPath, comparator: &ComparatorEqual{}, Value: PropertyValue{Value: "/test"}},
					{Property: PropertyMethod, comparator: &ComparatorEqual{}, Value: PropertyValue{Value: "GET"}},
				},
				Operator: OperatorOr,
			},
			req:        &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/test"}},
			wantErr:    false,
			wantResult: true,
		},
		{
			name: "Header mismatch error",
			TriggerSet: &TriggerSet{
				Triggers: []Trigger{
					{Property: PropertyHeader, comparator: &ComparatorEqual{}, Value: PropertyValue{Key: "key", Value: "value"}},
				},
			},
			req:        &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/test"}},
			wantErr:    true,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.TriggerSet.Match(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantResult, result)
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
triggers:
  - property: path
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
triggers:
  - property: path
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
triggers:
  - property: query
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
triggers:
  - property: header
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
