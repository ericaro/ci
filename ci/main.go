package main

import (
	"flag"
	"fmt"
	"github.com/ericaro/ci/format"
	"log"
	"os"
	"strings"
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
		fmt.Printf("LOG: %s  %s %s \n\n", id.GetName(), id.GetRemote(), id.GetBranch())

		refresh := job.GetRefresh()
		refreshStart := time.Unix(refresh.GetStart(), 0)
		refreshEnd := time.Unix(refresh.GetEnd(), 0)
		build := job.GetBuild()
		buildStart := time.Unix(build.GetStart(), 0)
		buildEnd := time.Unix(build.GetEnd(), 0)

		switch {
		case refreshEnd.Before(refreshStart):
			fmt.Printf("REFRESHING for %s\n\n", time.Since(refreshStart))
		case refresh.GetErrcode() != 0:
			fmt.Printf("REFRESH FAILED %s ago\n\n", time.Since(refreshEnd))
		default:
			fmt.Printf("REFRESH SUCCESS %s ago\n\n", time.Since(refreshEnd))
		}
		fmt.Printf("\n%s\n", strings.Replace(refresh.GetResult(), "\n", "\n    ", -1))

		fmt.Printf("\n\n")
		if refreshStart.Before(refreshEnd) { // if refresh has finished, print the build

			switch {
			case buildEnd.Before(buildStart):
				fmt.Printf("BUILDING for %s\n\n", time.Since(buildStart))
			case build.GetErrcode() != 0:
				fmt.Printf("BUILD FAILED %s ago\n\n", time.Since(buildEnd))
			default:
				fmt.Printf("BUILD SUCCESS %s ago\n\n", time.Since(buildEnd))
			}
			fmt.Printf("\n%s\n", strings.Replace(build.GetResult(), "\n", "\n    ", -1))

			if buildStart.Before(buildEnd) { // if build has finished, print a summary

				if refresh.GetErrcode() == 0 {
					fmt.Printf("\033[00;32mrefresh success  %s ago\033[00m (%s)\n", time.Since(refreshEnd), refreshEnd.Sub(refreshStart))
				} else {
					fmt.Printf("\033[00;31mrefresh failed   %s ago\033[00m (%s)\n", time.Since(refreshEnd), refreshEnd.Sub(refreshStart))
				}
				if build.GetErrcode() == 0 {
					fmt.Printf("\033[00;32mbuild   success  %s ago\033[00m (%s)\n", time.Since(buildEnd), buildEnd.Sub(buildStart))
				} else {
					fmt.Printf("\033[00;31mbuild   failed   %s ago\033[00m (%s)\n", time.Since(buildEnd), buildEnd.Sub(buildStart))
				}
			}
		}

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
