package main

// TODO(reed): separate into own client package

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

const AWS_US_EAST_HOST = `worker-aws-us-east-1.iron.io`

// The idea is:
//  validate arguments
//  if ^ goes well, config
//  if ^ goes well, run
//
//  ...and if anything goes wrong, help()
type Command interface {
	Args(...string) error // validate arguments
	Config()              // configure env variables
	Help() string         // custom command help, TODO(reed): export? really?
	Run()                 // cmd specific
}

type baseCmd struct {
	Token      string `json:"token"`
	ProjectID  string `json:"project_id"`
	Host       string `json:"host"`
	Protocol   string `json:"protocol"`
	Port       int    `json:"port"`
	APIVersion int    `json:"api_version"`
}

func (bc *baseCmd) Config() {
	// TODO(reed): better way to change zero value?
	bc.Host = AWS_US_EAST_HOST
	bc.Protocol = "https"
	bc.Port = 443
	bc.APIVersion = 2

	switch {
	// TODO(reed): env variables, etc.
	case exists(os.ExpandEnv("$HOME") + "/.iron.json"):
		body, _ := ioutil.ReadFile(os.ExpandEnv("$HOME") + "/.iron.json")
		json.Unmarshal(body, &bc)
		fallthrough
	case exists("iron.json"):
		body, _ := ioutil.ReadFile("iron.json")
		json.Unmarshal(body, &bc)
	}
}

func exists(fname string) bool {
	_, err := os.Stat(fname)
	return !os.IsNotExist(err)
}

func (bc baseCmd) baseURL() string {
	return bc.Protocol + "://" + bc.Host + ":" + strconv.Itoa(bc.Port) + "/" + strconv.Itoa(bc.APIVersion)
}

// return results
func (bc baseCmd) get(url string) (string, error) {
	// TODO(reed): query string params aren't gonna work w/ multiples.
	resp, err := http.Get(bc.baseURL() + url + "?oauth=" + bc.Token)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
}

// return results
func (bc baseCmd) post(url string) string {
	return ""
}
