// package ci handle ci job execution, persistence, andwsebservice.
package ci

import (
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/ericaro/ci/format"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

//Daemon defines the API for a Continuous Integration Server.
type Daemon interface {
	AddJob(path, remote, branch string) error
	RemoveJob(path string) error
	ListJobs(refreshResult, buildResult bool) *format.ListResponse
	JobDetails(job string) *format.LogResponse
	Run()
	Marshal() *format.Server
	Unmarshal(*format.Server) error
}

func NewDaemon(wd, dbfile string) (daemon Daemon, err error) {

	//Creates the daemon
	daemon = &ci{wd: wd, jobs: make(map[string]*job)}

	// read from disk if needed
	_, err = os.Stat(dbfile)

	if err == nil || !os.IsNotExist(err) {
		log.Printf("daemon.loading:%q", dbfile)
		// read the file
		file, err := os.Open(dbfile)
		if err != nil {
			log.Printf("error.daemon.opening:%q", err.Error())
			return daemon, err
		}
		defer file.Close()

		b, err := ioutil.ReadAll(file)
		if err != nil {
			log.Printf("error.daemon.reading:%q", err.Error())
			return daemon, err
		}

		// now protobuf read the content
		f := new(format.Server)
		err = proto.Unmarshal(b, f)
		if err != nil {
			log.Printf("error.daemon.parsing:%q", err.Error())
			return daemon, err
		}

		// the format is read, init the server
		err = daemon.Unmarshal(f)
		if err != nil {
			log.Printf("error.daemon.loading:%q", err.Error())
			return daemon, err
		}
	}
	// now the ci is fully created or unmarshaled
	//just log the job found
	for i, n := range daemon.ListJobs(false, false).GetJobs() {
		log.Printf("    daemon.job[%v]:%q,\n", i, n.GetId().GetName())
	}
	log.Printf("daemon.ready")

	// register a syscall hook to persist it on exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		for _ = range c {
			// sig is a ^C, handle it
			b, err := proto.Marshal(daemon.Marshal())
			if err != nil {
				log.Printf("error.daemon.marshaling:%q", err.Error())
				os.Exit(-1)
			}

			err = ioutil.WriteFile(filepath.Join(wd, dbfile), b, os.ModePerm)
			if err != nil {
				log.Printf("error.daemon.writefile:%q", err.Error())
				os.Exit(-1)
			}

			log.Printf("daemon.persisted")
			os.Exit(0)
		}
	}()
	return daemon, nil
}

// ci is a collection of jobs. It implements Server
type ci struct {
	jobs map[string]*job // path -> job
	wd   string          // absolute path to the working dir
}

// return a message describing the full details of a job.
func (c *ci) JobDetails(job string) *format.LogResponse {
	j := c.jobs[job]
	return &format.LogResponse{
		Job: j.Status(true, true),
	}
}

//ListJobs return a format.ListResponse describing all jobs.
// refreshResult = true means to add the output of the refresh action.
func (c *ci) ListJobs(refreshResult, buildResult bool) *format.ListResponse {

	js := make([]*format.Job, 0, len(c.jobs))
	for _, j := range c.jobs {
		js = append(js, j.Status(refreshResult, buildResult))
	}

	return &format.ListResponse{
		Jobs: js,
	}
}

//Run launch a run on every job
func (c *ci) Run() {

	for _, j := range c.jobs {
		j.Run() // I don't need to fork here, because Run() already handles that.
	}
}

func (c *ci) AddJob(path, remote, branch string) error {
	if _, exists := c.jobs[path]; exists {
		return fmt.Errorf("a job with this name already exists.")
	}
	c.jobs[path] = &job{name: path,
		remote: remote,
		branch: branch,
	}
	return nil
}

func (c *ci) RemoveJob(path string) error {
	if _, exists := c.jobs[path]; exists {

		//remove from the daemon server
		delete(c.jobs, path)

		//remove from local filesystem
		if err := os.RemoveAll(path); err != nil {
			if os.IsNotExist(err) {
				return nil //ok
			} else {
				return fmt.Errorf("cannot removing job's local directory: %s", err.Error())

			}
		}

	}
	return nil
}

// the main feature for a ci is to edit jobs, and persist them.

func (c *ci) Marshal() *format.Server {
	jobs := make([]*format.Job, 0, 100)
	for _, j := range c.jobs {
		jobs = append(jobs, j.Marshal())
	}
	return &format.Server{Jobs: jobs}
}

func (c *ci) Unmarshal(f *format.Server) error {

	// clean up the current object
	c.jobs = make(map[string]*job)

	for _, j := range f.Jobs {

		jb := job{}
		jb.Unmarshal(j)
		c.jobs[jb.name] = &jb

	}
	return nil
}
