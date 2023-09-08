package model

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	Equal       = "equal"
	NotEqual    = "not_equal"
	Contains    = "contains"
	NotContains = "not_contains"
)

var (
	ErrUnsupportedComparator = errors.New("unsupported comparator")
)

type Comparator interface {
	Compare(PropertyValue, interface{}) (bool, error)
}

type ComparatorEqual struct{}

func (c *ComparatorEqual) Compare(target PropertyValue, value interface{}) (bool, error) {
	switch v := value.(type) {
	case string:
		return target.Value == v, nil
	default:
		return false, fmt.Errorf("%w for %v", ErrUnsupportedComparator, value)
	}
}

type ComparatorNotEqual struct{}

func (c *ComparatorNotEqual) Compare(target PropertyValue, value interface{}) (bool, error) {
	switch v := value.(type) {
	case string:
		return target.Value != v, nil
	default:
		return false, fmt.Errorf("%w for %v", ErrUnsupportedComparator, value)
	}
}

type ComparatorContains struct{}

func (c *ComparatorContains) Compare(target PropertyValue, value interface{}) (bool, error) {
	switch v := value.(type) {
	case string:
		return strings.Contains(v, target.Value), nil
	case url.Values:
		values := map[string][]string(v)
		return checkValues(values, target.Key, target.Value), nil
	case http.Header:
		headers := map[string][]string(v)
		return checkValues(headers, target.Key, target.Value), nil
	default:
		return false, ErrUnsupportedComparator
	}
}

type ComparatorNotContains struct{}

func (c *ComparatorNotContains) Compare(target PropertyValue, value interface{}) (bool, error) {
	switch v := value.(type) {
	case string:
		return !strings.Contains(target.Value, v), nil
	case url.Values:
		return !checkValues(v, target.Key, target.Value), nil
	case http.Header:
		return !checkValues(v, target.Key, target.Value), nil
	default:
		return false, ErrUnsupportedComparator
	}
}

func checkValues(m map[string][]string, key, val string) bool {
	for _, y := range m[key] {
		if y == val {
			return true
		}
	}
	return false
}
