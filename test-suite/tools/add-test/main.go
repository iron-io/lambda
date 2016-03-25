package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/docker/docker/pkg/jsonmessage"
	iron_lambda "github.com/iron-io/lambda/lambda"
	"github.com/iron-io/lambda/test-suite/util"
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

func createLambdaFunction(l *lambda.Lambda, code []byte, runtime, role, name, handler string, timeout int) error {
	func_input := &lambda.CreateFunctionInput{
		Code:         &lambda.FunctionCode{ZipFile: code},
		Runtime:      aws.String(runtime),
		Role:         aws.String(role),
		FunctionName: aws.String(name),
		Handler:      aws.String(handler),
		Timeout:      aws.Int64(int64(timeout)),
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

			configInput := &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(name),
				Handler:      aws.String(handler),
				Timeout:      aws.Int64(int64(timeout)),
			}
			resp, err = l.UpdateFunctionConfiguration(configInput)
			if err != nil {
				log.Println("Could not update function configuration", err)
				return err
			}

		} else {
			return err
		}
	}

	log.Println(resp)
	return nil
}

func addToLambda(dir string) error {
	desc, err := util.ReadTestDescription(dir)
	if err != nil {
		return err
	}

	var zipContents []byte
	if desc.Runtime == "java8" {
		zipContents, err = ioutil.ReadFile(filepath.Join(dir, "test-build.jar"))
		if err != nil {
			return err
		}
	} else {
		zipContents, err = makeZip(dir)
		if err != nil {
			return err
		}
	}

	s := session.New(&aws.Config{Region: aws.String("us-east-1"), Credentials: credentials.NewEnvCredentials()})

	l := lambda.New(s)

	err = createLambdaFunction(l, zipContents, desc.Runtime, lambdaRole, desc.Name, desc.Handler, desc.Timeout)
	return err
}

type DockerJsonWriter struct {
	under io.Writer
	w     io.Writer
}

func NewDockerJsonWriter(under io.Writer) *DockerJsonWriter {
	r, w := io.Pipe()
	go func() {
		err := jsonmessage.DisplayJSONMessagesStream(r, under, 1, true, nil)
		log.Fatal(err)
	}()
	return &DockerJsonWriter{under, w}
}

func (djw *DockerJsonWriter) Write(p []byte) (int, error) {
	return djw.w.Write(p)
}

func addToIron(dir string) error {
	desc, err := util.ReadTestDescription(dir)
	if err != nil {
		return err
	}

	version := uuid.NewV4().String()
	imageNameVersion := fmt.Sprintf("%s/%s:%s", imagePrefix, desc.Name, version)

	err = util.MakeImage(dir, desc, imageNameVersion)
	if err != nil {
		return err
	}

	opts := iron_lambda.PushImageOptions{imageNameVersion, NewDockerJsonWriter(os.Stdout), true}
	err = iron_lambda.PushImage(opts)
	if err != nil {
		return err
	}

	return iron_lambda.RegisterWithIron(imageNameVersion)
}

type RegisterOn struct {
	aws  bool
	iron bool
}

func addTest(dir string, ro RegisterOn) error {
	if ro.aws {
		if err := addToLambda(dir); err != nil {
			return err
		}
	}
	if ro.iron {
		if err := addToIron(dir); err != nil {
			return err
		}
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

	noAws := flag.Bool("no-aws", false, "Do not register with AWS.")
	noIron := flag.Bool("no-iron", false, "Do not register with IronWorker.")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println(`Usage: ./add-test path/to/test [/more/paths...]

This will package all files and subdirectories except lambda.test as a test.`)
		flag.PrintDefaults()
		return
	}

	p := RegisterOn{aws: !*noAws, iron: !*noIron}
	for _, dir := range flag.Args() {
		if err := addTest(dir, p); err != nil {
			log.Fatal(err)
		}
	}
}
