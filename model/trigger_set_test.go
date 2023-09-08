package model

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
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
				Name: "test",
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
				Name: "test",
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
				Name: "test",
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
				Name: "test",
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
