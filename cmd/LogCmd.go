package cmd

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ericaro/ci/format"
)

type LogCmd struct {
	Server *string
	tail   *bool
}

func (c *LogCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	c.tail = fs.Bool("tail", false, "print the current job, and then poll for updates.")
	return fs
}
func (c *LogCmd) Run(args []string) {
	if len(args) != 1 {
		fmt.Printf("log command requires 1 arguments. Got %v\n", len(args))
		flag.Usage()
		os.Exit(-1)
	}

	jobname := args[0]

	req := &format.Request{
		Log: &format.LogRequest{
			Jobname: &jobname,
		},
	}
	b, r := GetRemoteExecution(*c.Server, req)

	fmt.Println(r.Print(), "\n")

	if r.Done() { // if refresh has finished, print the build
		fmt.Println(r.Summary())

		fmt.Println(b.Print(), "\n")
		if b.Done() { // if build has finished, print a summary
			fmt.Println(b.Summary())
		}
	}

	if *c.tail {
		for _ = range time.Tick(2 * time.Second) {

			newb, newr := GetRemoteExecution(*c.Server, req)

			fmt.Print(r.Tail(newr))
			fmt.Print(b.Tail(newb))
			b, r = newb, newr
		}
	} else { // when not in tail mode, always print out the summary for both refresh, and build
		if b.StartAfter(r) {
			fmt.Println("\n\n", r.Summary(), "\n", b.Summary())
		} else {
			fmt.Println("\n\n", b.Summary(), "\n", r.Summary())
		}
	}
}
