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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/iron-io/iron_go3/worker"
	"github.com/iron-io/lambda-test-suite/util"
)

func loadTests(filter string) ([]*util.TestDescription, error) {
	descs := []*util.TestDescription{}
	infos, err := ioutil.ReadDir("./tests/node")
	if err != nil {
		return descs, err
	}

	for _, info := range infos {
		p := filepath.Join("./tests/node", info.Name())
		if filter != "" {
			if !strings.Contains(p, filter) {
				continue
			}
		}

		d, err := util.ReadTestDescription(p)
		if err != nil {
			return descs, err
		}
		descs = append(descs, d)
	}
	return descs, nil
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
	for _, test := range tests {
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
			fmt.Printf("FAIL %s Output does not match!\n", test.Name)
			fmt.Printf("AWS lambda output '%s'\n", awss)
			fmt.Printf("Iron output '%s'\n", irons)
		} else {
			fmt.Printf("PASS %s\n", test.Name)
		}
	}
}
