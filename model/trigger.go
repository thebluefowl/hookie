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

type PropertyValue struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type Trigger struct {
	Name       string        `yaml:"name"`
	Property   Property      `yaml:"property"`
	Comparator string        `yaml:"comparator"`
	comparator Comparator    `yaml:"-"`
	Value      PropertyValue `yaml:"value"`
}

func (t *Trigger) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Trigger
	if err := unmarshal((*plain)(t)); err != nil {
		return err
	}
	switch t.Comparator {
	case Equal:
		t.comparator = &ComparatorEqual{}
	case NotEqual:
		t.comparator = &ComparatorNotEqual{}
	case Contains:
		t.comparator = &ComparatorContains{}
	case NotContains:
		t.comparator = &ComparatorNotContains{}
	}

	return t.Validate()
}

func (t *Trigger) Validate() error {
	switch t.Property {
	case PropertyPath, PropertyHost, PropertyMethod:
		if t.Value.Value == "" {
			return ErrEmptyRuleValue
		}
	case PropertyHeader, PropertyQuery:
		if t.Value.Key == "" || t.Value.Value == "" {
			return ErrEmptyRuleValue
		}
		if t.Comparator != Contains && t.Comparator != NotContains {
			return fmt.Errorf("unsupported Comparator for property %s", t.Property)
		}
	}
	return nil
}

func (t *Trigger) Match(req *http.Request) (bool, error) {
	value := t.Property.Value(req)
	return t.comparator.Compare(t.Value, value)
}
