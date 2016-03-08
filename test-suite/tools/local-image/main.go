package main

// Create a local docker image for each test on the command line.
//
// Usage: go run ./local-image.go [path/to/test]...

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/iron-io/lambda/test-suite/util"
)

var imagePrefix string

func makeLocalDocker(dir string) error {
	desc, err := util.ReadTestDescription(dir)
	if err != nil {
		return err
	}
	imageNameVersion := fmt.Sprintf("%s/%s:%s", imagePrefix, desc.Name, "latest")
	if err != nil {
		return err
	}

	err = util.MakeImage(dir, desc, imageNameVersion)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if imagePrefix = os.Getenv("IRON_LAMBDA_TEST_IMAGE_PREFIX"); imagePrefix == "" {
		log.Fatalf("IRON_LAMBDA_TEST_IMAGE_PREFIX not set")
	}

	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println(`Usage: ./local-image path/to/test [/more/paths...]

This will package all files and subdirectories except lambda.test as a test and create a local docker image.`)
		return
	}

	for _, dir := range flag.Args() {
		if err := makeLocalDocker(dir); err != nil {
			log.Fatal(err)
		}
	}
}
