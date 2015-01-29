package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ericaro/ci/format"
)

type logCmd struct {
	tail *bool
}

func (cmd *logCmd) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.tail = fs.Bool("tail", false, "print the current job, and then poll for updates.")
	return fs
}
func (cmd *logCmd) Run(args []string) {
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
	b, r := cmd.GetJob(req)

	fmt.Println(r.Print(), "\n")

	if r.Done() { // if refresh has finished, print the build
		fmt.Println(r.Summary())

		fmt.Println(b.Print(), "\n")
		if b.Done() { // if build has finished, print a summary
			fmt.Println(b.Summary())
		}
	}

	if *cmd.tail {
		for _ = range time.Tick(2 * time.Second) {

			newb, newr := cmd.GetJob(req)

			fmt.Print(r.Tail(newr))
			fmt.Print(b.Tail(newb))
			b, r = newb, newr
		}
	} else { // when not in tail mode, always print out the summary for both refresh, and build
		if b.start.After(r.end) {
			fmt.Println("\n\n", r.Summary(), "\n", b.Summary())
		} else {
			fmt.Println("\n\n", b.Summary(), "\n", r.Summary())
		}
	}
}

func (cmd *logCmd) GetJob(req *format.Request) (b, r *exec) {
	c := format.NewClient(*server)
	resp, err := c.Proto(req)
	if err != nil {
		log.Fatal(err.Error())
	}
	job := resp.GetLog().GetJob()
	// now present the resp
	//
	r = newExec(job.GetRefresh(), "refresh")
	b = newExec(job.GetBuild(), "build")
	return
}

type exec struct {
	x          *format.Execution
	start, end time.Time
	duration   time.Duration
	since      time.Duration
	name       string
}

func newExec(px *format.Execution, name string) *exec {
	x := exec{
		x:    px,
		name: name,
	}
	x.start = time.Unix(px.GetStart(), 0)
	x.end = time.Unix(px.GetEnd(), 0)

	x.duration = x.end.Sub(x.start)
	if x.end.Before(x.start) {
		x.since = time.Since(x.start)
	} else {
		x.since = time.Since(x.end)
	}

	return &x
}

// Done returns true when the execution is done
func (x *exec) Done() bool { return x.end.After(x.start) }

//Print returns a terminal friendly string representation of the current execution
func (x *exec) Print() string {
	buf := new(bytes.Buffer)
	switch {

	case x.end.Before(x.start): // active
		fmt.Fprintf(buf, "%s started %s ago.\n", x.name, x.since)

	case x.x.GetErrcode() != 0:
		fmt.Fprintf(buf, "%s \033[00;31mfailed\033[00m %s ago\n\n", x.name, x.since)

	default:
		fmt.Fprintf(buf, "%s \033[00;32msuccess\033[00m %s ago\n\n", x.name, x.since)
	}

	txt := x.x.GetResult()
	fmt.Fprintf(buf, "\n%s", strings.Replace(txt, "\n", "\n    ", -1))

	return buf.String()
}

//Tail returns changes between x (the assumed previous exec) and 'n' the new one
func (x *exec) Tail(n *exec) string {
	// ALG:
	// if same start then this is a diff otherwise this is a basic print
	if x.Done() && !n.Done() {
		// the new one is printed
		return n.Print()
	}
	if x.Done() && n.Done() { // there is nothing new to display
		return ""
	}
	// now we are in an interesting case:
	txt := x.x.GetResult()
	is := n.x.GetResult()
	tail := strings.TrimPrefix(is, txt)
	tail = strings.Replace(tail, "\n", "\n    ", -1)

	if !x.Done() && n.Done() { // the new has finished
		tail += "\n" + n.Summary() + "\n"
	}
	return tail
}

// Summary returns a small summary of the execution (status, duration and time since ended)
func (x *exec) Summary() string {
	if x.x.GetErrcode() == 0 {
		return fmt.Sprintf("%s \033[00;32msuccess\033[00m in %s, %s ago", x.name, x.duration, x.since)
	} else {
		return fmt.Sprintf("%s \033[00;31mfailed\033[00m in %s, %s ago", x.name, x.duration, x.since)
	}
}
