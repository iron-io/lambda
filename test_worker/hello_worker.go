package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	contents, _ := ioutil.ReadFile("/task/task_payload.json")
	fmt.Println(string(contents))
	fmt.Println("Hello from IronWorker!\n")
	fmt.Println(os.Getenv("TREBEK"))
}
