package main

import (
	"flag"

	"github.com/rakyll/command"
)

var (
	server = flag.String("s", "http://localhost:2020", "remote server address")
)

func main() {
	command.On("add",
		"<name> <remote> <branch>: adds a job on the ci-daemon", &addCmd{}, nil)
	command.On("remove",
		"<name>                  : removes a job", &removeCmd{}, nil)
	command.On("list",
		"                        : lists jobs on the server", &listCmd{}, nil)
	command.On("log",
		"<name>                  : logs details about a job", &logCmd{}, nil)

	command.ParseAndRun()

}
