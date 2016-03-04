package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	iron_lambda "github.com/iron-io/lambda/lambda"
)

type Payload map[string]interface{}

type TestDescription struct {
	Handler     string
	Name        string
	Runtime     string
	Event       Payload
	Description string // Completely ignored by test harness, just useful to convey intent of test.
	Timeout     int
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
	normalizedRuntime := strings.Replace(desc.Runtime, ".", "_", -1)
	desc.Name = fmt.Sprintf("lambda-test-suite-%s-%s", normalizedRuntime, desc.Name)
	return &desc, nil
}

func MakeImage(dir string, desc *TestDescription, imageNameVersion string) error {
	files := make([]iron_lambda.FileLike, 0)
	defer func() {
		for _, file := range files {
			file.(*os.File).Close()
		}
	}()

	first := false
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Skip dir itself.
		if !first {
			first = true
			return nil
		}

		if info.Name() == "lambda.test" {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		files = append(files, f)

		if info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return err
	}

	// FIXME(nikhil): Use some configuration username.
	err = iron_lambda.CreateImage(imageNameVersion, "iron/lambda-"+desc.Runtime, desc.Handler, files...)
	return err
}
