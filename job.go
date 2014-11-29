package ci

import (
	"bytes"
	"fmt"
	"github.com/ericaro/ci/format"
	"github.com/ericaro/mrepo"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

//job is the main object in a ci. it represent a project to be build.
// it is configured by a unique name, a remote url (git url to checkout the project)
// and a branch to checkout.
type job struct {
	name   string
	remote string
	branch string
	// cmd      string    // the command executed as a CI (default `make`)
	// args     []string  // args of the ci command default `ci`

	//other fields are local one.
	at       *time.Timer
	refresh  execution // info about the refresh execution
	build    execution // info about the build execution
	execLock sync.Mutex
}

func RunJobNow(name, remote, branch string) {
	j := &job{name: name, remote: remote, branch: branch}
	j.doRun()
	fmt.Println(j.refresh.result.String())
	fmt.Println(j.build.result.String())
	fmt.Println("done")

}

//Marshal serialize all information into a format.Job object.
func (j *job) Marshal() *format.Job { return j.Status(true, true) }

//Status serialize information into a format.Job object.
//
// if withRrefresh, it will include the refresh output
//
// if withBuild if will include the build output
func (j *job) Status(withRefresh, withBuild bool) *format.Job {
	return &format.Job{
		Id: &format.Jobid{
			Name:   &j.name,
			Remote: &j.remote,
			Branch: &j.branch,
		},
		Refresh: j.refresh.Status(withRefresh),
		Build:   j.build.Status(withBuild),
	}
}

//Unmarshal initialise the current job with values from the format.Job message
func (j *job) Unmarshal(f *format.Job) error {

	id := f.Id
	j.name = id.GetName()
	j.remote = id.GetRemote()
	j.branch = id.GetBranch()

	if err := j.refresh.Unmarshal(f.GetRefresh()); err != nil {
		return err
	}
	if err := j.build.Unmarshal(f.GetBuild()); err != nil {
		return err
	}
	return nil
}

//Run schedules (or reschedule) a run
func (j *job) Run() {
	j.RunWithDelay(10 * time.Second)
}

func (j *job) RunWithDelay(delay time.Duration) {
	if j.at == nil { // never scheduled before
		log.Printf("%s Run scheduled in %v", j.name, delay)
		j.at = time.AfterFunc(delay, j.doRun)
	} else {
		stopped := j.at.Reset(delay) // reschedule for a delay (either restart it or cancel before restarting)
		if stopped {
			log.Printf("%s Run postponed in %v", j.name, delay)
		} else {
			log.Printf("%s Run scheduled in %v", j.name, delay)
		}
	}
}

//doRun really execute the run
func (j *job) doRun() {
	log.Printf("Pulling %s", j.name)
	j.Refresh()
	log.Printf("Building %s", j.name)
	j.Build()
}

//Refresh the current job.
//
// Skip if there is an ongoing job.
func (j *job) Refresh() {
	j.execLock.Lock()
	defer j.execLock.Unlock()

	//to start we refresh all information: buffer, and start time.
	j.refresh.result = new(bytes.Buffer)
	j.refresh.start = time.Now() // mark the job as started
	j.refresh.errcode = 0        // no semantic here... yet
	defer func() {               // we will update stuff at the end
		j.refresh.end = time.Now() // mark the job as ended at the end of this call.
	}()
	// do the job now and return
	if err := j.dorefresh(j.refresh.result); err != nil {
		j.refresh.errcode = -1 // no semantic here... yet
		fmt.Fprintln(j.refresh.result, err.Error())
	} else {
		j.refresh.errcode = 0
	}
	log.Printf("Done refreshing job %s", j.name)
}

//Build the current job
func (j *job) Build() {

	j.execLock.Lock()
	defer j.execLock.Unlock()

	// check that the version has changed
	/* temp deactivated  */
	if j.build.version == j.refresh.version {
		// currently uptodate, nothing to do
		log.Printf("job %s has already been built", j.name)
		return
	}
	/**/

	// I'm gonna run
	// I'm under the protection of the lock
	// mark the version has built
	j.build.result = new(bytes.Buffer)
	j.build.start = time.Now() // mark the job as started
	defer func() {
		j.build.version = j.refresh.version
		j.build.end = time.Now() // mark the job as ended at the end of this call.
	}()

	// do the job now and return
	if err := j.dobuild(j.build.result); err != nil {
		j.build.errcode = -1 // no semantic here... yet
		fmt.Fprintln(j.build.result, err.Error())
	} else {
		j.build.errcode = 0
	}
	log.Printf("Done building job %s", j.name)

}
func (j *job) dobuild(w io.Writer) error {

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "working dir: %s\n", wd)

	fmt.Fprintf(w, "%s $ make ci\n", filepath.Join(wd, j.name))
	return mrepo.Make(filepath.Join(wd, j.name), "ci", w)
}

//dorefresh actually run the refresh command, it is unsafe to call it without caution. It should only update errcode, and result
func (j *job) dorefresh(w io.Writer) error {

	// NB: this is  a mammoth function. I know it, but I wasn't able to
	// split it down before having done everything.
	//
	// Step by step I will extract subffunctions to appropriate set of objects
	//

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	var cloned bool
	_, err = os.Stat(j.name)
	if os.IsNotExist(err) { // target does not exist, make it.
		fmt.Fprintf(w, "job dir does not exists. Will create one: %s\n", j.name)
		result, err := mrepo.GitClone(wd, j.name, j.remote, j.branch)
		cloned = true //
		fmt.Fprintln(w, result)
		if err != nil {
			return err
		}
	}

	wk := mrepo.NewWorkspace(filepath.Join(wd, j.name))

	if !cloned {
		err := wk.PullTop(w)
		if err != nil {
			return err
		}
	}
	digest, err := wk.Refresh(w)
	if err != nil {
		return err
	}
	copy(j.refresh.version[:], digest)
	return nil
}
