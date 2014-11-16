package main

import (
	"flag"
	"fmt"
	"github.com/ericaro/ci/format"
	"log"
	"os"
	"text/tabwriter"
	"time"
)

var (
	server = flag.String("s", "http://localhost:2020", "remote server address")
)

func main() {
	flag.Usage = usage
	flag.Parse()
	// real basic command line right now

	c := format.NewClient(*server)

	cmd := flag.Arg(0)

	switch cmd {

	case "add":
		//ci add job remote branch
		// TODO(ea) check arg count

		if flag.NArg() != 4 {
			fmt.Printf("%q command requires 3 arguments. Got %v\n", cmd, flag.NArg()-1)
			flag.Usage()
			os.Exit(-1)
		}
		job := flag.Arg(1)
		remote := flag.Arg(2)
		branch := flag.Arg(3)
		req := &format.Request{
			Add: &format.AddRequest{
				Id: &format.Jobid{
					Name:   &job,
					Remote: &remote,
					Branch: &branch,
				},
			},
		}

		resp, err := c.Proto(req)
		if err != nil {
			log.Fatal(err.Error())
		}
		if resp.Error != nil {
			fmt.Printf("%s\n", *resp.Error)
		} else {
			fmt.Printf("added %s %s %s\n", job, remote, branch)
		}
	case "remove":
		//ci add job remote branch
		// TODO(ea) check arg count
		if flag.NArg() != 2 {
			fmt.Printf("%q command requires 1 arguments. Got %v\n", cmd, flag.NArg()-1)
			flag.Usage()
			os.Exit(-1)
		}
		job := flag.Arg(1)
		req := &format.Request{
			Remove: &format.RemoveRequest{
				Jobname: &job,
			},
		}

		resp, err := c.Proto(req)
		if err != nil {
			log.Fatal(err.Error())
		}
		if resp.Error != nil {
			fmt.Printf("%s\n", *resp.Error)
		} else {
			fmt.Printf("removed %s\n", job)
		}

	case "log":
		if flag.NArg() != 2 {
			fmt.Printf("%q command requires 1 arguments. Got %v\n", cmd, flag.NArg()-1)
			flag.Usage()
			os.Exit(-1)
		}
		jobname := flag.Arg(1)
		req := &format.Request{
			Log: &format.LogRequest{
				Jobname: &jobname,
			},
		}
		resp, err := c.Proto(req)
		if err != nil {
			log.Fatal(err.Error())
		}
		job := resp.GetLog().GetJob()
		// now present the resp
		//
		id := job.GetId()
		fmt.Printf("%s (%s %s)\n", id.GetName(), id.GetRemote(), id.GetBranch())
		refresh := job.GetRefresh()
		refreshLast := time.Unix(refresh.GetEnd(), 0)

		var s string
		switch {
		case refresh.GetEnd() < refresh.GetStart():
			s = "Pulling since"
		case refresh.GetErrcode() != 0:
			s = "Pulling Failed on"
		default:
			s = "Pull Success on"
		}
		fmt.Printf("%s %s\n%s\n", s, refreshLast.String(), refresh.GetResult())

		build := job.GetBuild()
		buildLast := time.Unix(build.GetEnd(), 0)
		switch {
		case build.GetEnd() < build.GetStart():
			s = "Building since"
		case build.GetErrcode() != 0:
			s = "Building Failed on"
		default:
			s = "Build Success on"
		}
		fmt.Printf("\n%s %s\n%s\n", s, buildLast.String(), build.GetResult())

	case "list":
		if flag.NArg() != 1 {
			fmt.Printf("%q command requires no arguments. Got %v\n", cmd, flag.NArg()-1)
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

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", "Status", "Name", "Remote", "Branch", "Refreshed", "Build Date", "Commit")
		for _, s := range resp.List.Jobs {
			id := s.Id
			refresh := s.Refresh
			build := s.Build
			refreshLast := time.Unix(refresh.GetEnd(), 0)
			refreshFailed := refresh.GetErrcode() != 0
			buildLast := time.Unix(build.GetEnd(), 0)
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

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", status, id.GetName(), id.GetRemote(), id.GetBranch(), refreshLast.String(), buildLast.String(), refresh.GetVersion())
		}
		w.Flush()
	}

}
