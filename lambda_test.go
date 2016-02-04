package lambda

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

var client *docker.Client

func everythingIn(dir string) ([]FileLike, error) {
	arr := []FileLike{}
	first := false
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !first {
			first = true
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		arr = append(arr, f)

		if info.IsDir() {
			return filepath.SkipDir
		}

		return nil
	})
	return arr, err
}

func buildAndClean(name, base, handler, testdir string) error {
	if client == nil {
		panic("Need docker client")
	}
	defer client.RemoveImage(name)

	files, err := everythingIn(testdir)
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	defer os.Chdir(wd)
	os.Chdir(testdir)
	return CreateImage(name, base, handler, files...)
}

func TestCreateImageEmpty(t *testing.T) {
	err := CreateImage("iron-test/lambda-nodejs-empty", "iron/lambda-node", "test.run")
	if err == nil {
		t.Fatal("Expected error when no files passed")
	}
}

func TestCreateImageBasic(t *testing.T) {
	err := buildAndClean("iron-test/lambda-nodejs-basic", "iron/lambda-node", "test.run", "./nodejs-tests/basic")
	if err != nil {
		t.Fatal("CreateImage failed", err)
	}
}

func TestCreateImageWithDir(t *testing.T) {
	err := buildAndClean("iron-test/lambda-nodejs-withdir", "iron/lambda-node", "test.run", "./nodejs-tests/uuid")
	if err != nil {
		t.Fatal("CreateImage failed", err)
	}
}

func ensureBaseImage(name string) error {
	filteropts := docker.ListImagesOptions{
		Filter: name,
	}
	list, err := client.ListImages(filteropts)
	if len(list) > 0 {
		return nil
	}

	opts := docker.PullImageOptions{
		Repository: "iron/lambda-node",
	}

	var conf docker.AuthConfiguration
	err = client.PullImage(opts, conf)
	if err != nil {
		return err
	}

	return nil
}

func TestMain(m *testing.M) {
	flag.Parse()
	// Set up docker client to run clean up in individual tests.
	var err error
	client, err = docker.NewClientFromEnv()
	if err != nil {
		log.Fatal("Test could not connect to docker daemon", err)
	}

	// Grab node base image.
	if err := ensureBaseImage("iron/lambda-node"); err != nil {
		log.Fatal("Could not get nodejs base image to setup test.", err)
	}

	os.Exit(m.Run())
}
