package lambda

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fsouza/go-dockerclient"
)

type FileLike interface {
	io.Reader
	Stat() (os.FileInfo, error)
}

var ErrorNoFiles = errors.New("No files to add to image")

// Create a Dockerfile that adds each of the files to the base image. The
// expectation is that the base image sets up the current working directory
// inside the image correctly.  `handler` is set to be passed to node-lambda
// for now, but we may have to change this to accomodate other stacks.
func makeDockerfile(base string, handler string, files ...FileLike) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("FROM %s\n", base))

	for _, file := range files {
		// FIXME(nikhil): Validate path, no parent paths etc.
		info, err := file.Stat()
		if err != nil {
			return buf.Bytes(), err
		}
		buf.WriteString(fmt.Sprintf("ADD %s ./%s\n", info.Name(), info.Name()))
	}

	// FIXME(nikhil): Validate handler.
	buf.WriteString(fmt.Sprintf("CMD [\"%s\"]\n", handler))
	return buf.Bytes(), nil
}

func tarFile(tarrer *tar.Writer, file FileLike, info os.FileInfo) error {
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	if err := tarrer.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tarrer, file)
	return err
}

// using walk makes it impossible to test with fake files.
func tarDir(tarrer *tar.Writer, dir string) error {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = path

		if err := tarrer.WriteHeader(header); err != nil {
			return err
		}

		// Walk will get to contents of dir eventually.
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(tarrer, file)
		return err
	})
	return nil
}

func makeTar(dockerfile []byte, files ...FileLike) (io.Reader, error) {
	var tarred bytes.Buffer
	tarrer := tar.NewWriter(&tarred)

	now := time.Now()
	tarrer.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(len(dockerfile)), ModTime: now, AccessTime: now, ChangeTime: now})
	n, err := tarrer.Write(dockerfile)
	if err != nil {
		return nil, err
	}

	if n != len(dockerfile) {
		panic("Did not write all bytes")
	}

	for _, file := range files {
		info, err := file.Stat()
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			if err = tarDir(tarrer, info.Name()); err != nil {
				return nil, err
			}
		} else {
			if err = tarFile(tarrer, file, info); err != nil {
				return nil, err
			}
		}
	}

	return &tarred, nil
}

// Creates a docker image called `name`, using `base` as the base image.
// `handler` is the runtime-specific name to use for a lambda invocation (i.e.
// <module>.<function> for nodejs). `files` should be a list of files+dirs
// *relative to the current directory* that are to be included in the image.
func CreateImage(name string, base string, handler string, files ...FileLike) error {
	if len(files) == 0 {
		return ErrorNoFiles
	}

	df, err := makeDockerfile(base, handler, files...)
	if err != nil {
		return err
	}

	r, err := makeTar(df, files...)
	if err != nil {
		return err
	}

	var output bytes.Buffer

	opts := docker.BuildImageOptions{
		Name:         name,
		InputStream:  r,
		OutputStream: &output,
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}

	if err := client.BuildImage(opts); err != nil {
		return err
	}

	fmt.Println("Image output", output.String())
	return nil
}
