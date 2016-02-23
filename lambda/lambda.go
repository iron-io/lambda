package lambda

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/fsouza/go-dockerclient"
	"github.com/iron-io/iron_go3/worker"
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
		buf.WriteString(fmt.Sprintf("ADD [\"%s\", \"./%s\"]\n", info.Name(), info.Name()))
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

		// tarDir is called with an absolute path. `path` is relative to dir.
		// In the Docker image, we want to add the files at the 'top level'.
		// This means, the tar entry header must be relative to the base of the dir.
		//
		// For example, a node project is
		// - file1.js
		// - node_modules
		//
		// tarDir gets called with /abs/path/to/node_modules `path` will be the
		// absolute path to each entry. We want to convert a path `sub` to a tar entry of
		// `node_modules/sub`.
		p, _ := filepath.Rel(dir, path)
		header.Name = filepath.Join(filepath.Base(dir), p)

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
			// os.File.Name() is the path passed to os.Open, convert it to absolute path.
			p, err := filepath.Abs(file.(*os.File).Name())
			if err != nil {
				return nil, err
			}

			if err = tarDir(tarrer, p); err != nil {
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

func getClient() (*docker.Client, error) {
	return docker.NewClientFromEnv()
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
		Pull:         true,
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	if err := client.BuildImage(opts); err != nil {
		return err
	}

	fmt.Println("Image output", output.String())
	return nil
}

func ImageExists(imageName string) (bool, error) {
	client, err := getClient()
	if err != nil {
		return false, err
	}

	images, err := client.ListImages(docker.ListImagesOptions{Filter: imageName})
	if err != nil {
		return false, err
	}

	return len(images) > 0, nil
}

func RunImageWithPayload(imageName string, payload string) error {
	// FIXME(nikhil): Should we bother validating JSON here?

	// Write payload to temp file.
	fp, _ := filepath.Abs("./")
	payloadDir, err := ioutil.TempDir(fp, "iron-lambda-")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(payloadDir)
	}()

	payloadFilePath := filepath.Join(payloadDir, "payload.json")

	err = ioutil.WriteFile(payloadFilePath, []byte(payload), 0644)
	if err != nil {
		return errors.New(fmt.Sprintf("Error writing payload to file: %s", err.Error()))
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Env:       []string{"PAYLOAD_FILE=/mnt/payload.json"},
			Memory:    1024 * 1024 * 1024,
			CPUShares: 2,
			Hostname:  "Hello",
			Image:     imageName,
			Volumes: map[string]struct{}{
				"/mnt": {},
			},
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{payloadDir + ":/mnt:ro"},
		},
	}

	container, err := client.CreateContainer(opts)
	if err != nil {
		fmt.Println("CreateContainer error")
		return err
	}

	defer func() {
		client.RemoveContainer(docker.RemoveContainerOptions{
			ID: container.ID, RemoveVolumes: true, Force: true,
		})
	}()

	err = client.StartContainer(container.ID, nil)
	if err != nil {
		fmt.Println("StartContainer error")
		return err
	}

	attachOpts := docker.AttachToContainerOptions{
		Container:    container.ID,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		Logs:         true,
		Stream:       true,
		Stdout:       true,
		Stderr:       true,
	}

	err = client.AttachToContainer(attachOpts)
	if err != nil {
		return err
	}

	exitCode, err := client.WaitContainer(container.ID)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return errors.New(fmt.Sprintf("Container exited with non-zero exit code %d", exitCode))
	}

	return nil
}

// Registers public docker image named `imageNameVersion` as a IronWorker called `imageName`.
// For example,
//	  RegisterWithIron("foo/myimage:1", credentials.NewEnvCredentials()) will register a worker called "foo/myimage" that will use Docker Image "foo/myimage:1".
// The AWS credentials are required to configure environment variables for the image so that the AWS APIs can be used successfully.
func RegisterWithIron(imageNameVersion string, awsCredentials *credentials.Credentials) error {
	tokens := strings.Split(imageNameVersion, ":")
	if len(tokens) != 2 || tokens[0] == "" || tokens[1] == "" {
		return errors.New("Invalid image name. Should be of the form \"name:version\".")
	}

	imageName := tokens[0]

	creds, err := awsCredentials.Get()
	if err != nil {
		return errors.New(fmt.Sprintf("Could not extract AWS credentials to register environment variables with IronWorker: %s", err))
	}

	// Worker API doesn't have support for register yet, but we use it to extract the configuration.
	w := worker.New()
	url := fmt.Sprintf("https://%s/2/projects/%s/codes?oauth=%s", w.Settings.Host, w.Settings.ProjectId, w.Settings.Token)
	marshal, err := json.Marshal(map[string]interface{}{
		"name":  imageName,
		"image": imageNameVersion,
		"env_vars": map[string]string{
			"AWS_ACCESS_KEY_ID":           creds.AccessKeyID,
			"AWS_SECRET_ACCESS_KEY":       creds.SecretAccessKey,
			"AWS_LAMBDA_FUNCTION_NAME":    imageName,
			"AWS_LAMBDA_FUNCTION_VERSION": "1", // FIXME: swapi does not allow $ right now.
		},
	})

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	jsonWriter, err := mw.CreateFormField("data")
	if err != nil {
		log.Fatalf("This should never fail")
	}
	jsonWriter.Write(marshal)
	mw.Close()

	resp, err := http.Post(url, mw.FormDataContentType(), &body)
	if err == nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%s readall %s", imageName, err)
		}
		log.Println("Register", imageName, "with iron, response:", string(b))
	}
	return err
}

// Use AWS credentials from environment variables.
func RegisterWithIronDefaultAWSCredentials(imageNameVersion string) error {
	awsCredentials := credentials.NewEnvCredentials()
	return RegisterWithIron(imageNameVersion, awsCredentials)
}

func PushImage(imageNameVersion string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	tokens := strings.Split(imageNameVersion, ":")
	if len(tokens) != 2 || tokens[0] == "" || tokens[1] == "" {
		return errors.New("Invalid image name. Should be of the form \"name:version\".")
	}

	imageName, version := tokens[0], tokens[1]

	opts := docker.PushImageOptions{
		Name:         imageName,
		Tag:          version,
		OutputStream: os.Stdout,
	}

	auths, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return err
	}

	// FIXME(nikhil): Is there a nicer way to pick a config?
	var auth docker.AuthConfiguration
	if len(auths.Configs) > 0 {
		for _, a := range auths.Configs {
			auth = a
			break
		}
	} else {
		return errors.New("No docker authorizations found. Try `docker login`.")
	}

	return client.PushImage(opts, auth)
}
