package model

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrEmptyRuleValue = errors.New("empty rule value")
)

type Property string

const (
	PropertyPath   Property = "path"
	PropertyBody   Property = "body"
	PropertyHost   Property = "host"
	PropertyMethod Property = "method"
	PropertyHeader Property = "header"
	PropertyQuery  Property = "query"
)

func (p Property) Value(req *http.Request) (value interface{}) {
	switch p {
	case PropertyPath:
		return req.URL.Path
	case PropertyHost:
		return req.Host
	case PropertyMethod:
		return req.Method
	case PropertyHeader:
		return req.Header
	case PropertyQuery:
		return req.URL.Query()
	}
	return nil
}

type Operator string

const (
	OperatorAnd Operator = "and"
	OperatorOr  Operator = "or"
)

type Target struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Rule struct {
	Name       string     `yaml:"name"`
	Property   Property   `yaml:"property"`
	Comparator string     `yaml:"comparator"`
	comparator Comparator `yaml:"-"`
	Value      Target     `yaml:"value"`
}

func (r *Rule) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Rule
	if err := unmarshal((*plain)(r)); err != nil {
		return err
	}
	switch r.Comparator {
	case Equal:
		r.comparator = &ComparatorEqual{}
	case NotEqual:
		r.comparator = &ComparatorNotEqual{}
	case Contains:
		r.comparator = &ComparatorContains{}
	case NotContains:
		r.comparator = &ComparatorNotContains{}
	}

	return r.Validate()
}

func (r *Rule) Validate() error {
	switch r.Property {
	case PropertyPath, PropertyHost, PropertyMethod:
		if r.Value.Value == "" {
			return ErrEmptyRuleValue
		}
	case PropertyHeader, PropertyQuery:
		if r.Value.Key == "" || r.Value.Value == "" {
			return ErrEmptyRuleValue
		}
		if r.Comparator != Contains && r.Comparator != NotContains {
			return fmt.Errorf("unsupported Comparator for property %s", r.Property)
		}
	}
	return nil
}

func (r *Rule) Match(req *http.Request) (bool, error) {
	value := r.Property.Value(req)
	return r.comparator.Compare(r.Value, value)
}
