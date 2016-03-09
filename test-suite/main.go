package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/iron-io/iron_go3/worker"
	"github.com/iron-io/lambda/test-suite/util"
	"github.com/satori/go.uuid"
	"github.com/sendgrid/sendgrid-go"
)

func getSubDirs(basePath string) ([]string, error) {
	infos, err := ioutil.ReadDir(basePath)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0)
	for _, info := range infos {
		if !info.IsDir() {
			continue
		}
		subDirPath := filepath.Join(basePath, info.Name())
		result = append(result, subDirPath)
	}

	return result, nil
}

func loadTests(filter string) ([]*util.TestDescription, error) {
	testsRoot := "./tests"
	// assume test location <testsRoot>/<lang>/<test>/lambda.test
	descs := []*util.TestDescription{}

	langFolders, err := getSubDirs(testsRoot)
	if err != nil {
		return descs, err
	}

	allFolders := make([]string, 0)
	for _, folder := range langFolders {
		testFolders, err := getSubDirs(folder)
		if err != nil {
			return descs, err
		}
		allFolders = append(allFolders, testFolders...)
	}

	for _, folder := range allFolders {
		if filter != "" {
			if !strings.Contains(folder, filter) {
				continue
			}
		}

		d, err := util.ReadTestDescription(folder)
		if err != nil {
			return descs, fmt.Errorf("Could not load test: %s error: %s", folder, err)
		}
		descs = append(descs, d)
	}
	return descs, nil
}

func notifyFailure(name string) {
	var sgApiKey string
	if sgApiKey = os.Getenv("SENDGRID_API_KEY"); sgApiKey == "" {
		log.Println("SendGrid support not enabled.")
		return
	}

	var taskID string
	if taskID = os.Getenv("TASK_ID"); taskID == "" {
		log.Println("No task ID, not running on IronWorker. No emails will be sent.")
		return
	}

	message := sendgrid.NewMail()
	message.AddTos([]string{
		"dev@iron.io",
	})
	message.SetFromName("Lambda Test Suite")
	message.SetFrom("lambda-test-suite-notifications@iron.io")
	message.SetSubject(fmt.Sprintf("TEST-FAILURE %s", name))
	message.SetText(fmt.Sprintf(`The following test failed due to divergence between IronWorker and AWS Lambda output:

	%s: %s

Please check the task log for task ID %s for full output. DO NOT reply to this message.`, time.Now(), name, taskID))

	client := sendgrid.NewSendGridClientWithApiKey(sgApiKey)
	if err := client.Send(message); err != nil {
		log.Println("Error sending email", err)
	}
}

func main() {
	helpRequested := flag.Bool("h", false, "Show help")
	flag.Parse()
	if *helpRequested {
		fmt.Fprintln(os.Stderr, `Usage: ./lambda-test-suite [filter]
Runs all tests. If filter is passed, only runs tests matching filter. Filter is applied to entire path relative to tests/ directory.`)
		return
	}

	var filter string
	if flag.NArg() > 0 {
		filter = flag.Arg(0)
	}

	// Verify iron and aws connections.
	w := worker.New()
	_, err := w.TaskList()
	if err != nil {
		log.Fatal("Could not connect to iron.io API", err)
	}

	s := session.New(&aws.Config{Region: aws.String("us-east-1"), Credentials: credentials.NewEnvCredentials()})

	l := lambda.New(s)
	_, err = l.ListFunctions(&lambda.ListFunctionsInput{})
	if err != nil {
		log.Fatal("Could not connect to Lambda API", err)
	}

	cw := cloudwatchlogs.New(s)
	_, err = cw.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{})
	if err != nil {
		log.Fatal("Could not connect to CloudWatch API", err)
	}

	log.Print("All API connections successful.")

	tests, err := loadTests(filter)
	if err != nil {
		log.Fatal(err)
	}

	endMarker := uuid.NewV4().String()
	testResults := make(chan []string)
	endMarkerCount := 0
	for _, test := range tests {
		endMarkerCount++
		go runTest(endMarker, testResults, test, w, cw, l)
	}

	for endMarkerCount > 0 {
		lines := <-testResults
		for _, line := range lines {
			if line == endMarker {
				endMarkerCount--
			} else {
				log.Println(line)
			}
		}
	}
}

func runTest(endMarker string, result chan<- []string, test *util.TestDescription, w *worker.Worker, cw *cloudwatchlogs.CloudWatchLogs, l *lambda.Lambda) {
	defer func() {
		result <- []string{endMarker}
	}()

	awschan := make(chan io.Reader, 1)
	ironchan := make(chan io.Reader, 1)

	var wg sync.WaitGroup
	wg.Add(2)
	go runOnLambda(l, cw, &wg, test, awschan)
	go runOnIron(w, &wg, test, ironchan)

	wg.Wait()

	awsreader := <-awschan
	awss, _ := ioutil.ReadAll(awsreader)

	ironreader := <-ironchan
	irons, _ := ioutil.ReadAll(ironreader)

	if !bytes.Equal(awss, irons) {
		delimiter := "=========================================="
		result <- []string{
			fmt.Sprintf("FAIL %s Output does not match!", test.Name),
			fmt.Sprintf("AWS lambda output\n%s\n%s\n%s", delimiter, awss, delimiter),
			fmt.Sprintf("Iron output\n%s\n%s\n%s", delimiter, irons, delimiter),
		}
		notifyFailure(test.Name)
	} else {
		result <- []string{
			fmt.Sprintf("PASS %s", test.Name),
		}
	}
}
