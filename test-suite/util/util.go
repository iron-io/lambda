package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type Payload map[string]interface{}

type TestDescription struct {
	Handler     string
	Name        string
	Runtime     string
	Event       Payload
	Description string // Completely ignored by test harness, just useful to convey intent of test.
}

func ReadTestDescription(dir string) (*TestDescription, error) {
	c, err := ioutil.ReadFile(filepath.Join(dir, "lambda.test"))
	if err != nil {
		return nil, err
	}

	var desc TestDescription
	err = json.Unmarshal(c, &desc)
	if err != nil {
		return nil, err
	}

	desc.Name = fmt.Sprintf("lambda-test-suite-%s-%s", desc.Runtime, desc.Name)
	return &desc, nil
}
