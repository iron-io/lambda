package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/iron-io/iron_go3/config"
	"github.com/iron-io/iron_go3/mq"
)

func printMessages(msgs []mq.Message) {
	for _, msg := range msgs {
		fmt.Printf("%s %q\n", msg.Id, msg.Body)
	}
}

func printReservedMessages(msgs []mq.Message) {
	for _, msg := range msgs {
		fmt.Printf("%s %s %q\n", msg.Id, msg.ReservationId, msg.Body)
	}
}

// BLANKS name: url.com/endpoint
func printSubscribers(info mq.QueueInfo) {
	for _, subscriber := range info.Push.Subscribers {
		fmt.Printf("%s%s\n", BLANKS, subscriber.URL)
	}
}

// This is based on the format of func printMessages([]*mq.Message)
func readIds() ([]string, error) {
	var ids []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		if len(message) > 19 {
			id := message[:19] // We want the first 19 characters of the line, since an id is 19 characters long
			ids = append(ids, id)
		}
	}
	return ids, scanner.Err()
}

// Check if stdout is being piped
func isPipedOut() bool {
	fi, _ := os.Stdout.Stat()
	return (fi.Mode() & os.ModeNamedPipe) == os.ModeNamedPipe
}

func isPipedIn() bool {
	fi, _ := os.Stdin.Stat()
	return (fi.Mode() & os.ModeNamedPipe) == os.ModeNamedPipe
}

// TODO: Figure out the region for the hud url
// seriously though
// this is super duper hacky
// it only works with the public cluster mq-aws-us-east-1-1
func getHudTag(settings config.Settings) (string, error) {
	res, err := http.Get("https://auth.iron.io/1/clusters?oauth=" + settings.Token)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	clusters := struct {
		Clusters []struct {
			Tag string `json:"tag"`
			URL string `json:"url"`
		} `json:"clusters"`
	}{}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	err = json.Unmarshal(b, &clusters)
	if err != nil {
		return "", err
	}
	queueHost := settings.Host
	for _, cluster := range clusters.Clusters {
		if cluster.URL == queueHost {
			return cluster.Tag, err
		}
	}
	return "", fmt.Errorf("no hud tags found")
}

func printQueueHudURL(prefix string, q mq.Queue) {
	if tag, err := getHudTag(q.Settings); err == nil {
		fmt.Printf("%sVisit hud-e.iron.io/mq/%s/projects/%s/queues/%s for more info.\n", prefix,
			tag,
			q.Settings.ProjectId,
			q.Name)
	}
}

func mqProjectName(settings config.Settings) (string, error) {
	res, err := http.Get("https://auth.iron.io/1/projects/" + settings.ProjectId + "?oauth=" + settings.Token)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	projects := struct {
		Project struct {
			Name string `json:"name"`
		} `json:"project"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&projects)
	if err != nil {
		return "", err
	}
	return projects.Project.Name, nil
}
