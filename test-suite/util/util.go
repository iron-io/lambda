package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	iron_lambda "github.com/iron-io/lambda/lambda"
)

type TestDescription struct {
	Handler     string
	Name        string
	Runtime     string
	Event       interface{}
	Description string // Completely ignored by test harness, just useful to convey intent of test.

	// The test's timeout in seconds, valid timeout as imposed by Lambda
	// is between 1 and 300 inclusive.
	// If no Timeout is specified the 60 sec default is used
	Timeout int
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

	if desc.Timeout == 0 {
		desc.Timeout = 60
	} else if desc.Timeout < 1 {
		desc.Timeout = 1
	} else if desc.Timeout > 300 {
		desc.Timeout = 300
	}

	return &desc, nil
}

func MakeImage(dir string, desc *TestDescription, imageNameVersion string) error {
	files := make([]iron_lambda.FileLike, 0)
	defer func() {
		for _, file := range files {
			file.(*os.File).Close()
		}
	}()

	hasTestJar := false
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

		if info.Name() == "test-build.jar" {
			hasTestJar = true
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

	if desc.Runtime == "java8" && !hasTestJar {
		return errors.New("One of the files MUST be test-build.jar for Java tests.")
	}

	opts := iron_lambda.CreateImageOptions{imageNameVersion, "iron/lambda-" + desc.Runtime, "", desc.Handler, os.Stdout, false}
	// FIXME(nikhil): Use some configuration username.
	if desc.Runtime == "java8" {
		opts.Package = "test-build.jar"
	}
	err = iron_lambda.CreateImage(opts, files...)
	return err
}

func RemoveTimestampAndRequestIdFromAwsLogLine(line, requestId string) (string, bool) {
	return removeTimestampAndRequestIdFromLogLine(line, requestId, awsRequestIdRegexp)
}
func RemoveTimestampAndRequestIdFromIronLogLine(line, requestId string) (string, bool) {
	return removeTimestampAndRequestIdFromLogLine(line, requestId, ironRequestIdRegexp)
}

var ironRequestIdRegexp *regexp.Regexp = regexp.MustCompile("^[0-9a-fA-F]{24}$")
var awsRequestIdRegexp *regexp.Regexp = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")
var timestampRegexp *regexp.Regexp = regexp.MustCompile("^20\\d{2}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}\\.\\d{3,6}Z")

func removeTimestampAndRequestIdFromLogLine(line, requestId string, requestIdRegexp *regexp.Regexp) (string, bool) {
	sep := "\t"
	parts := strings.Split(line, sep)

	// assume timestamp is before request_id
	for i, p := range parts {
		if p == requestId {
			hasTimeStamp := i > 0 && timestampRegexp.MatchString(parts[i-1])
			if hasTimeStamp {
				parts = append(parts[:i-1], parts[i+1:]...)
			} else {
				parts = append(parts[:i], parts[i+1:]...)
			}
			return strings.Join(parts, sep), true
		}
	}

	//remove a log line from another request_id
	for _, p := range parts {
		if requestIdRegexp.MatchString(p) {
			return "", false
		}
	}
	return line, true
}
