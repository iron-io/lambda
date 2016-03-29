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
	"github.com/satori/go.uuid"
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
func makeDockerfile(base string, package_ string, handler string, files ...FileLike) ([]byte, error) {
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

	buf.WriteString("CMD [")
	if package_ != "" {
		buf.WriteString(fmt.Sprintf("\"%s\", ", package_))
	}
	// FIXME(nikhil): Validate handler.
	buf.WriteString(fmt.Sprintf("\"%s\"", handler))
	buf.WriteString("]\n")

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

type CreateImageOptions struct {
	Name          string
	Base          string
	Package       string // Used for Java, empty string for others.
	Handler       string
	OutputStream  io.Writer
	RawJSONStream bool
}

type PushImageOptions struct {
	NameVersion   string
	OutputStream  io.Writer
	RawJSONStream bool
}

// Creates a docker image called `name`, using `base` as the base image.
// `handler` is the runtime-specific name to use for a lambda invocation (i.e.
// <module>.<function> for nodejs). `files` should be a list of files+dirs
// *relative to the current directory* that are to be included in the image.
func CreateImage(opts CreateImageOptions, files ...FileLike) error {
	if len(files) == 0 {
		return ErrorNoFiles
	}

	df, err := makeDockerfile(opts.Base, opts.Package, opts.Handler, files...)
	if err != nil {
		return err
	}

	r, err := makeTar(df, files...)
	if err != nil {
		return err
	}

	buildopts := docker.BuildImageOptions{
		Name:          opts.Name,
		InputStream:   r,
		OutputStream:  opts.OutputStream,
		RawJSONStream: opts.RawJSONStream,
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	if err := client.BuildImage(buildopts); err != nil {
		return err
	}

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

	var allocatedMemory = int64(300 * 1024 * 1024)
	envs := []string{"PAYLOAD_FILE=/mnt/payload.json"}
	envs = append(envs, "AWS_LAMBDA_FUNCTION_NAME="+imageName)
	envs = append(envs, "AWS_LAMBDA_FUNCTION_VERSION=$LATEST")
	envs = append(envs, "TASK_ID="+uuid.NewV4().String())
	envs = append(envs, fmt.Sprintf("TASK_MAXRAM=%d", allocatedMemory))
	// Try to forward AWS credentials.
	{
		creds := credentials.NewEnvCredentials()
		v, err := creds.Get()
		if err == nil {
			envs = append(envs, "AWS_ACCESS_KEY_ID="+v.AccessKeyID)
			envs = append(envs, "AWS_SECRET_ACCESS_KEY="+v.SecretAccessKey)
		}
	}

	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Env:       envs,
			Memory:    allocatedMemory,
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
func RegisterWithIron(imageNameVersion string) error {
	tokens := strings.Split(imageNameVersion, ":")
	if len(tokens) != 2 || tokens[0] == "" || tokens[1] == "" {
		return errors.New("Invalid image name. Should be of the form \"name:version\".")
	}

	imageName := tokens[0]

	// Worker API doesn't have support for register yet, but we use it to extract the configuration.
	w := worker.New()
	url := fmt.Sprintf("https://%s/2/projects/%s/codes?oauth=%s", w.Settings.Host, w.Settings.ProjectId, w.Settings.Token)
	registerOpts := map[string]interface{}{
		"name":  imageName,
		"image": imageNameVersion,
		"env_vars": map[string]string{
			"AWS_LAMBDA_FUNCTION_NAME":    imageName,
			"AWS_LAMBDA_FUNCTION_VERSION": "1", // FIXME: swapi does not allow $ right now.
		},
	}

	// Try to forward AWS credentials.
	{
		creds := credentials.NewEnvCredentials()
		v, err := creds.Get()
		if err == nil {
			registerOpts["env_vars"].(map[string]string)["AWS_ACCESS_KEY_ID"] = v.AccessKeyID
			registerOpts["env_vars"].(map[string]string)["AWS_SECRET_ACCESS_KEY"] = v.SecretAccessKey
		}
	}

	marshal, err := json.Marshal(registerOpts)
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

func PushImage(in PushImageOptions) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	tokens := strings.Split(in.NameVersion, ":")
	if len(tokens) != 2 || tokens[0] == "" || tokens[1] == "" {
		return errors.New("Invalid image name. Should be of the form \"name:version\".")
	}

	imageName, version := tokens[0], tokens[1]

	opts := docker.PushImageOptions{
		Name:          imageName,
		Tag:           version,
		OutputStream:  in.OutputStream,
		RawJSONStream: in.RawJSONStream,
	}

	auths, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return err
	}

	var auth docker.AuthConfiguration
	for _, a := range auths.Configs {
		if strings.Contains(a.ServerAddress, "index.docker.io") {
			auth = a
			break
		}
	}

	if auth.ServerAddress == "" {
		return errors.New("No Docker Hub (index.docker.io) authorization found. Try `docker login`.")
	}

	return client.PushImage(opts, auth)
}
