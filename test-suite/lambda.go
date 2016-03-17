package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/iron-io/lambda/test-suite/util"
)

func indexOf(a string, list []string) int {
	for i, b := range list {
		if b == a {
			return i
		}
	}
	return -1
}

func cleanNodeJsAwsOutput(output string) (string, error) {
	var buf bytes.Buffer
	if strings.HasPrefix(output, "START RequestId:") {
		scanner := bufio.NewScanner(strings.NewReader(output))
		if scanner.Scan() {
			firstLine := scanner.Text()
			fields := strings.Fields(firstLine)
			if len(fields) > 2 {
				id := fields[2]
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					if strings.HasPrefix(line, "END") {
						return buf.String(), nil
					}

					// Remove timestamp
					idx := strings.IndexByte(line, 'Z')
					if idx >= 0 {
						untimed := strings.TrimSpace(line[idx+1:])
						unprefix := strings.TrimPrefix(untimed, id)
						buf.WriteString(strings.TrimSpace(unprefix))
						buf.WriteRune('\n')
					} else {
						buf.WriteString(line)
						buf.WriteRune('\n')
					}
				}
				if err := scanner.Err(); err != nil {
					return "", err
				}
			}
		}
	}

	return "", errors.New(fmt.Sprintf("Don't know how to clean '%s'", output))
}

// Processes all requests log lines inside the log and succedes only with the latest one
// The log line format:  [some data] [timestamp] [request_id] [some other data]
// Request start format: START RequestId: [request_id] [some data]
// Request end format:   END RequestId: [request_id] [some data]
// AWS report format:    REPORT RequestId: [request_id] [some data]
func cleanPython27AwsOutput(output string) (string, error) {
	var buf bytes.Buffer
	var requestId string = ""
	knownRequestIds := make(map[string]bool, 0)
	requestProcessed := false
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		//processing START, END and REPORT log lines
		requestIdFieldIndex := indexOf("RequestId:", fields)
		if (requestIdFieldIndex > 0) && (requestIdFieldIndex+1 < len(fields)) {
			requestIdInLine := fields[requestIdFieldIndex+1]
			prefix := fields[requestIdFieldIndex-1]
			requestIdIsKnown := knownRequestIds[requestIdInLine]
			skip := true
			switch prefix {
			case "START":
				requestId = requestIdInLine
				knownRequestIds[requestId] = true
				requestIdIsKnown = true

				// in case of multiple requests in the log we should use only the last one
				buf.Reset()
				requestProcessed = false
			case "END":
				if requestIdIsKnown {
					requestId = ""
					requestProcessed = true
				}
			case "REPORT":
				{
					// no specific action needed
				}
			default:
				skip = false
			}

			if skip {
				if !requestIdIsKnown {
					return "", errors.New(fmt.Sprintf("Unknown request_id '%s' in the line '%s' of the log '%s'", requestIdInLine, line, output))
				}
				continue
			}
		}

		line = util.RemoveTimestampAndRequestIdFromLogLine(line, requestId)

		buf.WriteString(line)
		buf.WriteRune('\n')
		if err := scanner.Err(); err != nil {
			return "", err
		}
	}

	if !requestProcessed {
		return "", errors.New(fmt.Sprintf("Don't know how to clean '%s'", output))
	} else {
		return buf.String(), nil
	}

}

func cleanAwsGeneric(old_output, output string) (string, error) {
	if old_output == output {
		return "", errors.New("No change in the log")
	}

	if strings.HasPrefix(output, old_output) {
		return output[len(old_output):], nil
	}
	return output, nil
}

func cleanAws(old_output, output, runtime string) (string, error) {
	output, err := cleanAwsGeneric(old_output, output)
	if err != nil {
		return "", err
	}

	switch runtime {
	case "nodejs":
		return cleanNodeJsAwsOutput(output)
	case "python2.7":
		return cleanPython27AwsOutput(output)
	case "java8":
		return cleanNodeJsAwsOutput(output)
	default:
		return output, nil
	}
}

func runOnLambda(l *lambda.Lambda, cw *cloudwatchlogs.CloudWatchLogs, wg *sync.WaitGroup, test *util.TestDescription, result chan<- io.Reader) {
	var output bytes.Buffer
	defer func() {
		result <- &output
		wg.Done()
	}()

	name := test.Name

	old_invocation_log, err := getLog(cw, name)
	if err != nil {
		old_invocation_log = ""
	}

	payload, _ := json.Marshal(test.Event)

	invoke_input := &lambda.InvokeInput{
		FunctionName:   aws.String(name),
		InvocationType: aws.String("Event"),
		Payload:        payload,
	}
	_, err = l.Invoke(invoke_input)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error invoking function %s %s", name, err))
		return
	}

	latency := 1 //1 second for network and infrastructure timeouts
	timeout := time.Duration(test.Timeout+latency) * time.Second

	time.Sleep(timeout)

	invocation_log, err := getLog(cw, name)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error getting log %s %s", name, err))
		return
	}

	final, err := cleanAws(old_invocation_log, invocation_log, test.Runtime)

	if err != nil {
		output.WriteString(fmt.Sprintf("Error cleaning log %s %s", name, err))
		return
	}

	output.WriteString(final)
}

func getLog(cw *cloudwatchlogs.CloudWatchLogs, name string) (string, error) {
	groupPrefix := aws.String("/aws/lambda/" + name)
	groups, err := cw.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{LogGroupNamePrefix: groupPrefix})
	if err != nil {
		return "", err
	}

	if len(groups.LogGroups) < 1 {
		return "", errors.New(fmt.Sprintf("No log group found for %s", name))
	}

	group := groups.LogGroups[0]
	// We don't handle the case where lambda functions may share prefixes but we get the list of groups back in non-lexicographic order. Reminder in case that ever happens.
	if *group.LogGroupName != *groupPrefix {
		log.Fatal("Got group matching prefix but not actual", groupPrefix, group.LogGroupName)
	}

	streams, err := cw.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: group.LogGroupName,
		Descending:   aws.Bool(true),
		OrderBy:      aws.String("LastEventTime"),
	})

	if err != nil {
		return "", err
	}

	if len(streams.LogStreams) < 1 {
		return "", errors.New(fmt.Sprintf("No log streams found for %s", name))
	}

	stream := streams.LogStreams[0]

	get_log_input := &cloudwatchlogs.GetLogEventsInput{
		LogStreamName: stream.LogStreamName,
		LogGroupName:  group.LogGroupName,
		StartFromHead: aws.Bool(true),
	}

	events, err := cw.GetLogEvents(get_log_input)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	for _, event := range events.Events {
		output.WriteString(*event.Message)
	}

	return output.String(), nil
}
