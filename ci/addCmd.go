package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ericaro/ci/format"
)

type addCmd struct{}

func (cmd *addCmd) Flags(fs *flag.FlagSet) *flag.FlagSet { return fs }
func (cmd *addCmd) Run(args []string) {
	c := format.NewClient(*server)
	//ci add job remote branch
	// TODO(ea) check arg count

	if len(args) != 3 {
		fmt.Printf("add command requires 3 arguments. Got %v\n", len(args))
		flag.Usage()
		os.Exit(-1)
	}
	job, remote, branch := args[0], args[1], args[2]
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
}
