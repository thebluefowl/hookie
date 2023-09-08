package model

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComparatorEqual_Compare(t *testing.T) {
	c := &ComparatorEqual{}

	// 1. Test for matching strings
	target := PropertyValue{Value: "hello"}
	val := "hello"
	result, err := c.Compare(target, val)
	assert.True(t, result, "Expected strings to match and return true")
	assert.Nil(t, err, "Expected no error for matching strings")

	// 2. Test for non-matching strings
	val2 := "world"
	result, err = c.Compare(target, val2)
	assert.False(t, result, "Expected strings to not match and return false")
	assert.Nil(t, err, "Expected no error for non-matching strings")

	// 3. Test for non-string type
	val3 := 123
	result, err = c.Compare(target, val3)
	expectedError := fmt.Errorf("%w for %v", ErrUnsupportedComparator, val3)
	assert.False(t, result, "Expected false for non-string type")
	assert.Equal(t, expectedError, err, "Expected ErrUnsupportedComparator error for non-string type")
}

func TestComparatorNotEqual_Compare(t *testing.T) {
	c := &ComparatorNotEqual{}

	// 1. Test for non-matching strings
	target := PropertyValue{Value: "hello"}
	val := "world"
	result, err := c.Compare(target, val)
	assert.True(t, result, "Expected strings to not match and return true")
	assert.Nil(t, err, "Expected no error for non-matching strings")

	// 2. Test for matching strings
	val2 := "hello"
	result, err = c.Compare(target, val2)
	assert.False(t, result, "Expected strings to match and return false")
	assert.Nil(t, err, "Expected no error for matching strings")

	// 3. Test for non-string type
	val3 := 123
	result, err = c.Compare(target, val3)
	expectedError := fmt.Errorf("%w for %v", ErrUnsupportedComparator, val3)
	assert.False(t, result, "Expected false for non-string type")
	assert.Equal(t, expectedError, err, "Expected ErrUnsupportedComparator error for non-string type")
}

func TestComparatorContains_Compare(t *testing.T) {
	c := &ComparatorContains{}

	// 1. Test for string contains
	target := PropertyValue{Value: "hello world"}
	val := "world"
	result, err := c.Compare(target, val)
	assert.True(t, result, "Expected target string to contain value string")
	assert.Nil(t, err, "Expected no error for string contains")

	// 2. Test for url.Values contains
	target2 := PropertyValue{Key: "key1", Value: "value1"}
	val2 := url.Values{"key1": []string{"value1"}}
	result, err = c.Compare(target2, val2)
	assert.True(t, result, "Expected url.Values to contain target key-value pair")
	assert.Nil(t, err, "Expected no error for url.Values contains")

	// 3. Test for http.Header contains
	target3 := PropertyValue{Key: "HeaderKey", Value: "HeaderValue"}
	val3 := http.Header{"HeaderKey": []string{"HeaderValue"}}
	result, err = c.Compare(target3, val3)
	assert.True(t, result, "Expected http.Header to contain target key-value pair")
	assert.Nil(t, err, "Expected no error for http.Header contains")

	// 4. Test for unsupported type
	val4 := 123
	result, err = c.Compare(target, val4)
	assert.False(t, result, "Expected false for unsupported type")
	assert.Equal(t, ErrUnsupportedComparator, err, "Expected ErrUnsupportedComparator error for unsupported type")
}

func TestComparatorNotContains_Compare(t *testing.T) {
	c := &ComparatorNotContains{}

	// 1. Test for string not contains
	target := PropertyValue{Value: "hello world"}
	val := "mars"
	result, err := c.Compare(target, val)
	assert.True(t, result, "Expected target string to not contain value string")
	assert.Nil(t, err, "Expected no error for string not contains")

	// Negative case for string
	valNeg := "world"
	result, err = c.Compare(target, valNeg)
	assert.False(t, result, "Expected target string to contain value string, thus returning false")
	assert.Nil(t, err, "Expected no error for string contains")

	// 2. Test for url.Values not contains
	target2 := PropertyValue{Key: "key1", Value: "value1"}
	val2 := url.Values{"key2": []string{"value2"}}
	result, err = c.Compare(target2, val2)
	assert.True(t, result, "Expected url.Values to not contain target key-value pair")
	assert.Nil(t, err, "Expected no error for url.Values not contains")

	// Negative case for url.Values
	valNeg2 := url.Values{"key1": []string{"value1"}}
	result, err = c.Compare(target2, valNeg2)
	assert.False(t, result, "Expected url.Values to contain target key-value pair, thus returning false")
	assert.Nil(t, err, "Expected no error for url.Values contains")

	// 3. Test for http.Header not contains
	target3 := PropertyValue{Key: "HeaderKey", Value: "HeaderValue"}
	val3 := http.Header{"OtherHeaderKey": []string{"OtherValue"}}
	result, err = c.Compare(target3, val3)
	assert.True(t, result, "Expected http.Header to not contain target key-value pair")
	assert.Nil(t, err, "Expected no error for http.Header not contains")

	// Negative case for http.Header
	valNeg3 := http.Header{"HeaderKey": []string{"HeaderValue"}}
	result, err = c.Compare(target3, valNeg3)
	assert.False(t, result, "Expected http.Header to contain target key-value pair, thus returning false")
	assert.Nil(t, err, "Expected no error for http.Header contains")

	// 4. Test for unsupported type
	val4 := 123
	result, err = c.Compare(target, val4)
	assert.False(t, result, "Expected false for unsupported type")
	assert.Equal(t, ErrUnsupportedComparator, err, "Expected ErrUnsupportedComparator error for unsupported type")
}
