package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ericaro/ci/format"
)

type RemoveCmd struct{ Server *string }

func (cmd *RemoveCmd) Flags(fs *flag.FlagSet) *flag.FlagSet { return fs }
func (cmd *RemoveCmd) Run(args []string) {
	c := format.NewClient(*cmd.Server)

	//ci add job remote branch
	// TODO(ea) check arg count
	if len(args) != 1 {
		fmt.Printf("remove command requires 1 arguments. Got %v\n", len(args))
		flag.Usage()
		os.Exit(-1)
	}

	job := args[0]
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

}
