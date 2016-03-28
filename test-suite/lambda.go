package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
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

// Processes all requests log lines inside the log and succedes only with the latest one
// The log line format:  [some data] [timestamp] [request_id] [some other data]
// Request start format: START RequestId: [request_id] [some data]
// Request end format:   END RequestId: [request_id] [some data]
// AWS report format:    REPORT RequestId: [request_id] [some data]
func cleanLambda(output string) (string, error) {
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

		line, isOk := util.RemoveTimestampAndRequestIdFromAwsLogLine(line, requestId)
		if isOk {
			buf.WriteString(line)
			buf.WriteRune('\n')
		}
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

func findFirstRequestIdFromLog(output string) string {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		//processing START, END and REPORT log lines
		requestIdFieldIndex := indexOf("RequestId:", fields)
		if (requestIdFieldIndex > 0) && (requestIdFieldIndex+1 < len(fields)) {
			requestIdInLine := fields[requestIdFieldIndex+1]
			prefix := fields[requestIdFieldIndex-1]

			if prefix == "START" {
				return requestIdInLine
			}
		}
	}

	return ""
}

func isEndOfRequestMarkerPresent(output, requestId string) bool {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		//processing START, END and REPORT log lines
		requestIdFieldIndex := indexOf("RequestId:", fields)
		if (requestIdFieldIndex > 0) && (requestIdFieldIndex+1 < len(fields)) {
			requestIdInLine := fields[requestIdFieldIndex+1]
			prefix := fields[requestIdFieldIndex-1]

			if prefix == "END" && requestIdInLine == requestId {
				return true
			}
		}
	}
	return false
}

func getLogAndDetectRequestComplete(logGetter func() (string, error), old_output, requestId string) (string, string, bool, error) {

	output, err := logGetter()
	if err != nil {
		return "", "", false, err
	}

	log := strings.TrimPrefix(output, old_output)

	if requestId == "" {
		requestId = findFirstRequestIdFromLog(log)
	}

	if requestId != "" {
		marker := isEndOfRequestMarkerPresent(strings.TrimPrefix(output, old_output), requestId)
		return output, requestId, marker, nil
	}

	return log, "", false, nil
}

//Returns a result and a debug channels. The channels are closed on test run finalization
func runOnLambda(l *lambda.Lambda, cw *cloudwatchlogs.CloudWatchLogs, test *util.TestDescription) (<-chan string, <-chan string) {
	result := make(chan string, 1)
	debug := make(chan string, 1)
	go func() {
		defer close(result)
		defer close(debug)

		name := test.Name

		debug <- "Getting old log"
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
		debug <- "Enqueuing task"
		_, err = l.Invoke(invoke_input)
		if err != nil {
			debug <- fmt.Sprintf("Error invoking function %s ", err)
			return
		}

		timeout := time.Duration(test.Timeout) * time.Second

		debug <- "Waiting for task"
		now := time.Now()
		elapsed := time.After(timeout)
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		log := ""
		completed := false
		requestId := ""
		logGetter := func() (string, error) {
			return getLog(cw, name)
		}
		
		// waiting for test.Timeout or for the full log whichever occurs first
	logWaitLoop:
		for {
			select {
			case <-elapsed:
				break logWaitLoop
			case <-ticker.C:
				log, requestId, completed, err = getLogAndDetectRequestComplete(logGetter, old_invocation_log, requestId)
				if err != nil {
					debug <- fmt.Sprintf("Error getting log %s ", err)
					return
				}
				if completed {
					break logWaitLoop
				}
			}
		}

		if !completed {
			log, requestId, completed, err = getLogAndDetectRequestComplete(logGetter, old_invocation_log, requestId)
			if err != nil {
				debug <- fmt.Sprintf("Error getting log %s ", err)
				return
			}
			if !completed {
				if requestId != "" {
					debug <- fmt.Sprintf("Request Id: %s", requestId)
				}
				debug <- time.Now().Sub(now).String()
				logLines := strings.Split(log, "\n")
				switch len(logLines) {
				case 0:
					debug <- "Log for current test run is empty"
				case 1:
					debug <- fmt.Sprintf("Log does not contain entries for current test run:\n%s", logLines[0])
				default:
					debug <- fmt.Sprintf("Log does not contain entries for current test run:\n%s\n...\n%s", logLines[0], logLines[len(logLines)-1])
				}
				return
			}
		}
		debug <- fmt.Sprintf("Request Id: %s", requestId)
		final, err := cleanLambda(log)

		if err != nil {
			debug <- fmt.Sprintf("Error cleaning log  %s", err)
			return
		}

		result <- final
	}()
	return result, debug
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
		// allow 3 minute local vs server time out of sync
		StartTime: aws.Int64(time.Now().Add(-3*time.Minute).Unix() * 1000),
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
