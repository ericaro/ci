package main

import (
	"flag"

	"github.com/rakyll/command"
)

var (
	// you can build you own version of the exe using
	//go install -ldflags '-X main.DefaultCIServer http://yourip:2020'

	DefaultCIServer = "http://localhost:2020"
	server          = flag.String("s", DefaultCIServer, "remote server address")
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
