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
					if idx == -1 {
						return "", errors.New(fmt.Sprintf("Expected timestamp at beginning of line, but could not find one it seems '%s'", line))
					}

					untimed := strings.TrimSpace(line[idx+1:])
					unprefix := strings.TrimPrefix(untimed, id)
					buf.WriteString(strings.TrimSpace(unprefix))
					buf.WriteRune('\n')
				}
				if err := scanner.Err(); err != nil {
					return "", err
				}
			}
		}
	}

	return "", errors.New(fmt.Sprintf("Don't know how to clean '%s'", output))
}

func cleanPython27AwsOutput(output string) (string, error) {
	var buf bytes.Buffer
	if strings.HasPrefix(output, "START RequestId:") {
		scanner := bufio.NewScanner(strings.NewReader(output))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "START RequestId:") {
				continue
			}
			if strings.HasPrefix(line, "END RequestId:") {
				return buf.String(), nil
			}
			buf.WriteString(line)
			buf.WriteRune('\n')
			if err := scanner.Err(); err != nil {
				return "", err
			}
		}
	}

	return "", errors.New(fmt.Sprintf("Don't know how to clean '%s'", output))
}

func clean(output, runtime string) (string, error) {
	switch runtime {
	case "nodejs":
		return cleanNodeJsAwsOutput(output)
	case "python2.7":
		return cleanPython27AwsOutput(output)
	default:
		return output, (error)(nil)
	}
}

func runOnLambda(l *lambda.Lambda, cw *cloudwatchlogs.CloudWatchLogs, wg *sync.WaitGroup, test *util.TestDescription, result chan<- io.Reader) {
	var output bytes.Buffer
	defer func() {
		result <- &output
		wg.Done()
	}()

	name := test.Name

	payload, _ := json.Marshal(test.Event)

	invoke_input := &lambda.InvokeInput{
		FunctionName:   aws.String(name),
		InvocationType: aws.String("Event"),
		Payload:        payload,
	}
	_, err := l.Invoke(invoke_input)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error invoking function %s %s", name, err))
		return
	}
	timeout := 30 * time.Second
	if test.Timeout != 0 {
		timeout = time.Duration(test.Timeout) * time.Second
	}

	time.Sleep(timeout)

	invocation_log, err := getLog(cw, name)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error getting log %s %s", name, err))
		return
	}

	final, err := clean(invocation_log, test.Runtime)

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
