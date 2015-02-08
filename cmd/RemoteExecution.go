package cmd

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ericaro/ci/format"
)

func GetRemoteExecution(remoteurl string, req *format.Request) (b, r *RemoteExecution) {
	c := format.NewClient(remoteurl)
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

type RemoteExecution struct {
	x          *format.Execution
	start, end time.Time
	duration   time.Duration
	since      time.Duration
	name       string
}

func newExec(px *format.Execution, name string) *RemoteExecution {
	x := RemoteExecution{
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

func (x *RemoteExecution) StartAfter(z *RemoteExecution) bool {
	return x.start.After(z.end)
}

// Done returns true when the execution is done
func (x *RemoteExecution) Done() bool { return x.end.After(x.start) }

//Print returns a terminal friendly string representation of the current execution
func (x *RemoteExecution) Print() string {
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

//Tail returns changes between x (the assumed previous RemoteExecution) and 'n' the new one
func (x *RemoteExecution) Tail(n *RemoteExecution) string {
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
func (x *RemoteExecution) Summary() string {
	if x.x.GetErrcode() == 0 {
		return fmt.Sprintf("%s \033[00;32msuccess\033[00m in %s, %s ago", x.name, x.duration, x.since)
	} else {
		return fmt.Sprintf("%s \033[00;31mfailed\033[00m in %s, %s ago", x.name, x.duration, x.since)
	}
}
