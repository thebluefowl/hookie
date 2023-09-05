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
		ruleset    *Ruleset
		req        *http.Request
		wantErr    bool
		wantResult bool
	}{
		{
			name: "AND match",
			ruleset: &Ruleset{
				Name: "test",
				Rules: []Rule{
					{Property: PropertyPath, comparator: &ComparatorEqual{}, Value: Target{Value: "/test"}},
					{Property: PropertyMethod, comparator: &ComparatorEqual{}, Value: Target{Value: "GET"}},
				},
				Operator: OperatorAnd,
			},
			req:        &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "/test"}},
			wantErr:    false,
			wantResult: true,
		},
		{
			name: "AND no match",
			ruleset: &Ruleset{
				Name: "test",
				Rules: []Rule{
					{Property: PropertyPath, comparator: &ComparatorEqual{}, Value: Target{Value: "/test"}},
					{Property: PropertyMethod, comparator: &ComparatorEqual{}, Value: Target{Value: "GET"}},
				},
				Operator: OperatorAnd,
			},
			req:        &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/test"}},
			wantErr:    false,
			wantResult: false,
		},
		{
			name: "OR match",
			ruleset: &Ruleset{
				Name: "test",
				Rules: []Rule{
					{Property: PropertyPath, comparator: &ComparatorEqual{}, Value: Target{Value: "/test"}},
					{Property: PropertyMethod, comparator: &ComparatorEqual{}, Value: Target{Value: "GET"}},
				},
				Operator: OperatorOr,
			},
			req:        &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/test"}},
			wantErr:    false,
			wantResult: true,
		},
		{
			name: "Header mismatch error",
			ruleset: &Ruleset{
				Name: "test",
				Rules: []Rule{
					{Property: PropertyHeader, comparator: &ComparatorEqual{}, Value: Target{Key: "key", Value: "value"}},
				},
			},
			req:        &http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/test"}},
			wantErr:    true,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.ruleset.Match(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantResult, result)
		})
	}
}
