package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/ericaro/ci/format"
)

type listCmd struct{}

func (cmd *listCmd) Flags(fs *flag.FlagSet) *flag.FlagSet { return fs }
func (cmd *listCmd) Run(args []string) {
	c := format.NewClient(*server)

	if len(args) != 0 {
		fmt.Printf("list command requires no arguments. Got %v\n", len(args))
		flag.Usage()
		os.Exit(-1)
	}

	req := &format.Request{
		List: &format.ListRequest{}, // no options available in the command line, yet
	}

	resp, err := c.Proto(req)
	if err != nil {
		log.Fatal(err.Error())
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "Status", "Name", "Remote", "Branch", "Version")
	for _, s := range resp.List.Jobs {
		id := s.Id
		refresh := s.Refresh
		build := s.Build
		refreshFailed := refresh.GetErrcode() != 0
		buildFailed := build.GetErrcode() != 0

		uptodate := refresh.GetVersion() == build.GetVersion()
		var status string
		switch {
		case refresh.GetEnd() < refresh.GetStart():
			status = "Pulling"
		case build.GetEnd() < build.GetStart():
			status = "Building"
		case !uptodate:
			status = "Need Build"
		case refreshFailed:
			status = "Pulling Failed"
		case buildFailed:
			status = "Building Failed"
		default:
			status = "Success"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", status, id.GetName(), id.GetRemote(), id.GetBranch(), refresh.GetVersion())
	}
	w.Flush()
}
