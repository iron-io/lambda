package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/iron-io/iron_go3/worker"
	"github.com/iron-io/lambda/test-suite/util"
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

	testLocations := make(map[string]string)
	for _, folder := range allFolders {

		d, err := util.ReadTestDescription(folder)
		if err != nil {
			return descs, fmt.Errorf("Could not load test: %s error: %s", folder, err)
		}
		key := d.Name
		if otherFolder, ok := testLocations[key]; ok {
			return descs, fmt.Errorf("Duplicate test name detected. Runtime: %s, Name: %s, Location1: %s, Location2: %s ", d.Runtime, d.Name, otherFolder, folder)
		}

		testLocations[key] = folder

		if filter == "" || strings.Contains(folder, filter) {
			descs = append(descs, d)
		}

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
	if len(tests) == 0 {
		log.Fatal("No tests to run")
	}

	// expected duration for all tests to run in a sequential way
	// after the fullTimeout expires no test result is accepted and `Timeout` message is reported for a test
	fullTimeout := 0
	for _, test := range tests {
		fullTimeout += test.Timeout + 5
	}

	concurrency := util.NewSemaphore(5) // using a limit, otherwise AWS fails with `ThrottlingException: Rate exceeded` on log retrieval

	endOfTime := time.Now().Add(time.Duration(fullTimeout) * time.Second)
	var testResults <-chan []string = nil
	for _, test := range tests {
		r := runTest(test, w, cw, l, endOfTime, concurrency)

		// forwarding messages from all tests to a single channel
		testResults = util.JoinChannels(testResults, r)
	}

	passed, failed := make([]string, 0, len(tests)), make([]string, 0, len(tests))

	for {
		lines, ok := <-testResults
		if !ok {
			break
		}

		if len(lines) > 0 {
			if strings.HasPrefix(lines[0], "PASS ") {
				passed = append(passed, lines[0])
			}
			if strings.HasPrefix(lines[0], "FAIL ") {
				failed = append(failed, lines[0])
			}
		}

		for _, line := range lines {
			log.Println(line)
		}
	}

	log.Println(fmt.Sprintf("Total %d passed and %d failed tests", len(passed), len(failed)))
	for _, line := range passed {
		log.Println(line)
	}
	for _, line := range failed {
		log.Println(line)
	}

	if len(failed) > 0 {
		os.Exit(1)
	}
}

//Returns a channel with a test run result and debug messages
func runTest(test *util.TestDescription, w *worker.Worker, cw *cloudwatchlogs.CloudWatchLogs, l *lambda.Lambda, waitEnd time.Time, s util.Semaphore) <-chan []string {
	result := make(chan []string)

	go func() {
		defer close(result)

		s.Lock()
		defer s.Unlock()

		testName := test.Name

		result <- []string{
			fmt.Sprintf("Starting test %s", testName),
		}

		endOfWait := time.NewTimer(waitEnd.Sub(time.Now()))

		awschan, awsdbg := runOnLambda(l, cw, test)
		ironchan, irondbg := runOnIron(w, test)

		// redirecting every debug message from test runs to result without waiting test results
		// before closing `result` aslo waits for closing of debug channels
		defer util.ForwardInBackground("DBG AWS Lambda "+testName+" ", awsdbg, result)()
		defer util.ForwardInBackground("DBG Iron "+testName+" ", irondbg, result)()

		// waiting for test results or for the timeout whichever occurs first
		var awss, irons *bytes.Buffer
		elapsed := false
		for !elapsed && (awschan != nil || ironchan != nil) {
			select {
			case data, ok := <-awschan:
				{
					if ok {
						if awss == nil {
							awss = &bytes.Buffer{}
						}
						awss.WriteString(data)
					} else {
						awschan = nil
					}
				}
			case data, ok := <-ironchan:
				{
					if ok {
						if irons == nil {
							irons = &bytes.Buffer{}
						}
						irons.WriteString(data)
					} else {
						ironchan = nil
					}
				}
			case <-endOfWait.C:
				elapsed = true
			}
		}

		delimiter := "=========================================="
		awsOutputStr, awsOutput := "No AWS lambda output", ""
		if awss != nil {
			awsOutput = string(awss.Bytes())
			awsOutputStr = fmt.Sprintf("AWS lambda output\n%s\n%s\n%s", delimiter, awsOutput, delimiter)
		}
		ironOutputStr, ironOutput := "No Iron output", ""
		if irons != nil {
			ironOutput = string(irons.Bytes())
			ironOutputStr = fmt.Sprintf("Iron output\n%s\n%s\n%s", delimiter, ironOutput, delimiter)
		}

		if elapsed {
			result <- []string{
				fmt.Sprintf("FAIL %s Timeout elapsed!", testName),
				awsOutputStr,
				ironOutputStr,
			}
			notifyFailure(testName)
		} else if awsOutput != ironOutput || awss == nil || irons == nil {
			result <- []string{
				fmt.Sprintf("FAIL %s Output does not match!", testName),
				awsOutputStr,
				ironOutputStr,
			}
			notifyFailure(testName)
		} else {
			if awss == nil {
				panic(testName + " " + awsOutputStr)
			}
			if irons == nil {
				panic(testName + " " + ironOutputStr)
			}
			result <- []string{
				fmt.Sprintf("PASS %s", testName),
			}
		}
	}()

	return result
}
