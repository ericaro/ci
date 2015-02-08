package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	Usage = `USAGE %[1]s [-options] <command> <args...>
			
DESCRIPTION:

  %[1]s is a client to manage a ci-daemon server.

  <command> can be:

    - add <name> <remote> <branch>: adds a job on the ci-daemon
    - remove <name>               : removes a job
    - list                        : lists jobs on the server
    - log <name>                  : logs details about a job

OPTIONS:

`
	Example = `
EXAMPLE

To add a repo to be built:

  %[1]s add mrepo git@github.com:ericaro/mrepo.git master

To check a build progress:

  %[1]s log mrepo

`
)

func usage() {
	fmt.Printf(Usage, os.Args[0])
	flag.PrintDefaults()
	fmt.Printf(Example, os.Args[0])
}
