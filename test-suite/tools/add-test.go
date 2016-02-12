package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/fsouza/go-dockerclient"
	"github.com/iron-io/iron_go3/worker"
	iron_lambda "github.com/iron-io/lambda"
	"github.com/iron-io/lambda-test-suite/util"
	"github.com/satori/go.uuid"
)

var imagePrefix string
var lambdaRole string

func makeZip(dir string) ([]byte, error) {
	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)
	_ = zipWriter
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

		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			log.Println(err)
			return err
		}
		p, _ := filepath.Rel(dir, path)
		hdr.Name = filepath.Join(p)
		if info.IsDir() {
			hdr.Name += "/"
		}

		w, err := zipWriter.CreateHeader(hdr)
		if err != nil {
			log.Println(err)
			return err
		}

		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}

			_, err = io.Copy(w, f)
			f.Close()
			return err
		}
		return nil
	})
	zipWriter.Close()
	return buffer.Bytes(), err
}

func createLambdaFunction(l *lambda.Lambda, code []byte, runtime, role, name, handler string) error {
	func_input := &lambda.CreateFunctionInput{
		Code:         &lambda.FunctionCode{ZipFile: code},
		Runtime:      aws.String(runtime),
		Role:         aws.String(role),
		FunctionName: aws.String(name),
		Handler:      aws.String(handler),
	}

	resp, err := l.CreateFunction(func_input)
	if err != nil {
		if err.(awserr.Error).Code() == "ResourceConflictException" {
			log.Println("Function already exists, trying to update code")
			input := &lambda.UpdateFunctionCodeInput{
				FunctionName: aws.String(name),
				ZipFile:      code,
			}

			resp, err = l.UpdateFunctionCode(input)
			if err != nil {
				log.Println("Could not update function code", err)
				return err
			}
		} else {
			return err
		}
	}

	_ = resp
	return nil
}

func addToLambda(dir string) error {
	desc, err := util.ReadTestDescription(dir)
	if err != nil {
		return err
	}

	zipContents, err := makeZip(dir)
	if err != nil {
		return err
	}

	s := session.New(&aws.Config{Region: aws.String("us-east-1"), Credentials: credentials.NewEnvCredentials()})

	l := lambda.New(s)

	err = createLambdaFunction(l, zipContents, desc.Runtime, lambdaRole, desc.Name, desc.Handler)
	return err
}

func makeImage(dir string, desc *util.TestDescription, imageNameVersion string) error {
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

func registerWithIron(imageName, imageNameVersion string) error {
	// Worker API doesn't have support for register yet, but we use it to extract the configuration.
	w := worker.New()
	url := fmt.Sprintf("https://%s/2/projects/%s/codes?oauth=%s", w.Settings.Host, w.Settings.ProjectId, w.Settings.Token)
	marshal, err := json.Marshal(map[string]string{
		"name":  imageName,
		"image": imageNameVersion,
	})

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	jsonWriter, err := mw.CreateFormField("data")
	if err != nil {
		log.Fatalf("This should never fail")
	}
	jsonWriter.Write(marshal)
	mw.Close()

	_, err = http.Post(url, mw.FormDataContentType(), &body)
	return err
}

func addToIron(dir string) error {
	desc, err := util.ReadTestDescription(dir)
	if err != nil {
		return err
	}

	version := uuid.NewV4().String()
	imageName := fmt.Sprintf("%s/%s", imagePrefix, desc.Name)
	imageNameVersion := fmt.Sprintf("%s:%s", imageName, version)

	err = makeImage(dir, desc, imageNameVersion)
	if err != nil {
		return err
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}

	authconfigs, err := docker.NewAuthConfigurationsFromDockerCfg()
	if err != nil {
		return err
	}

	var config docker.AuthConfiguration
	for _, v := range authconfigs.Configs {
		config = v
	}

	// Upload to Hub with correct version.
	pushopts := docker.PushImageOptions{
		Name:         imageName,
		Tag:          version,
		OutputStream: os.Stderr,
	}
	log.Println("PushOpts", pushopts)

	err = client.PushImage(pushopts, config)
	if err != nil {
		log.Println("PushImage error")
		return err
	}

	return registerWithIron(imageName, imageNameVersion)
}

func addTest(dir string) error {
	if err := addToLambda(dir); err != nil {
		return err
	}
	if err := addToIron(dir); err != nil {
		return err
	}
	return nil
}

func main() {
	if imagePrefix = os.Getenv("IRON_LAMBDA_TEST_IMAGE_PREFIX"); imagePrefix == "" {
		log.Fatalf("IRON_LAMBDA_TEST_IMAGE_PREFIX not set")
	}

	if lambdaRole = os.Getenv("IRON_LAMBDA_TEST_LAMBDA_ROLE"); lambdaRole == "" {
		log.Fatalf("IRON_LAMBDA_TEST_LAMBDA_ROLE not set")
	}

	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println(`Usage: ./add-test path/to/test [/more/paths...]

This will package all files and subdirectories except lambda.test as a test.`)
		return
	}

	for _, dir := range flag.Args() {
		if err := addTest(dir); err != nil {
			log.Fatal(err)
		}
	}
}
