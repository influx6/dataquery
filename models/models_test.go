package datamodel

import (
	"testing"
	"time"

	"github.com/influx6/flux"
)

type zum struct {
	D int `model:"d_o"`
}

//TestModelAttr creates a basic test case of a modelattr
func TestModelAttr(t *testing.T) {

	na := NewModelAttr("name", "name", &zum{1})

	dt := &zum{3}

	if err := na.Validate(dt); err != nil {
		flux.FatalFailed(t, "Validation failed: %s", err.Error())
	}

	flux.LogPassed(t, "Validation Passed: %s", dt)
}

// TestModelPassAndFailure sets up a new map model and validates that the wrong data fails
func TestModelPassAndFailure(t *testing.T) {

	mo := NewModels(map[string]interface{}{
		"name": "",
		"age":  0,
		"date": time.Now(),
	})

	key, err := mo.Validate(map[string]interface{}{
		"name": "alex ewetumo",
		"age":  20,
		"date": time.Now(),
	})

	if err != nil {
		flux.FatalFailed(t, "Validation key failed: %s : %s", key, err.Error())
	}

	flux.LogPassed(t, "Validation passed!")

	key, err = mo.Validate(map[string]interface{}{
		"name": "alex ewetumo",
		"age":  20,
		"date": time.Duration(300) * time.Minute,
	})

	if err == nil {
		flux.FatalFailed(t, "Validation key did not failed: %s", "date")
	}

	flux.LogPassed(t, "Validation failed as expected: %s : %s", key, err.Error())
}

// TestLists checks if we can also assign and validates lists of values
func TestLists(t *testing.T) {
	mo := NewModels(map[string]interface{}{
		"name": []string{},
	})

	key, err := mo.Validate(map[string]interface{}{
		"name": []string{"alex ewetumo"},
	})

	if err != nil {
		flux.FatalFailed(t, "Validation key failed: %s : %s", key, err.Error())
	}

	flux.LogPassed(t, "Validation passed!")
}

// TestInterLists tests wether we can assign a []string to an array type
func TestInterLists(t *testing.T) {
	mo := NewModels(map[string]interface{}{
		"name": []interface{}{},
	})

	key, err := mo.Validate(map[string]interface{}{
		"name": []string{"alex ewetumo"},
	})

	if err != nil {
		flux.LogPassed(t, "Validation passed! We cant assign %s to %s!", key, err.Error())
		return
	}

	flux.FatalFailed(t, "Validation key did not fail!")
}

// TestModelStructAttr if we can generates a map of attributes/fields of a inited struct
func TestModelStructAttr(t *testing.T) {

	b := &zum{1}

	bo, err := NewModelStruct(b, "model")

	if err != nil {
		flux.FatalFailed(t, "Unable to create modelAttr for %s", err.Error())
	}

	flux.LogPassed(t, "Successfully generated modelstruct with tag: %s %s", b, bo)
}

// TestModelStructTypeAttr if we can generates a map of field types form the struct type itself
func TestModelStructTypeAttr(t *testing.T) {

	b := &zum{1}

	bo, err := NewModelStruct(b, "model")

	if err != nil {
		flux.FatalFailed(t, "Unable to create modelAttr for %s", err.Error())
	}

	flux.LogPassed(t, "Successfully generated modelstruct with tag: %s %s", b, bo)
}

// TestStructModels generates a modelattr and Model corresponding for validating values
func TestStructModels(t *testing.T) {

	b := &zum{1}

	bo, err := NewStructModels(b, "model")

	if err != nil {
		flux.FatalFailed(t, "Unable to create modelAttr for %s", err.Error())
	}

	flux.LogPassed(t, "Successfully generated modelstruct with tag: %s %s", b, bo)

	key, err := bo.Validate(map[string]interface{}{
		"d_o": 20,
	})

	if err != nil {
		flux.FatalFailed(t, "Validation key failed: %s : %s", key, err.Error())
	}

	flux.LogPassed(t, "Validation passed!")
}
