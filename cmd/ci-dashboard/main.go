package main

import (
	"flag"
	"fmt"
	"github.com/ericaro/ci/format"
	"log"
	"math"
	"net/http"
	"sort"
	"time"
)

var (
	server = flag.String("s", "http://localhost:2020", "remote server address")
	title  = flag.String("t", "CI Dashboard", "CI title")
	port   = flag.Int("p", 8080, "http port to listen to")
	prop   = flag.Float64("prop", 4, "cell width ~= prop*cell height")
)

func main() {
	flag.Parse()

	log.Printf("dashboard available at http://locahost:%v\n", *port)

	d := new(Dashboard)
	d.Title = *title
	http.ListenAndServe(fmt.Sprintf(":%v", *port), d)
}

func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/favicon.ico" {
		_, err := w.Write(favicon)
		if err != nil {
			log.Printf("Error Rendering favico: %s", err.Error())
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
			log.Printf("%s 501 %s", r.Method, r.URL.String())
			return
		}

	}
	d.JobMatrix = d.FillJobMatrix()
	err := dashboard.Execute(w, d)
	if err != nil {
		log.Printf("Error Rendering template: %s", err.Error())
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		log.Printf("%s 501 %s", r.Method, r.URL.String())
		return
	}
	log.Printf("%s 200 %s", r.Method, r.URL.String())
}

func (d *Dashboard) FillJobMatrix() (jobs [][]Job) {
	jobin, err := GetJobs()
	if err != nil {
		log.Printf("error getting jobs: %s", err.Error())
		return //empty matrix, if no server communication
	}
	//skip trival case
	if len(jobin) == 0 {
		return nil
	}
	//format all jobs into a local type list
	joblist := make([]Job, 0, len(jobin))

	for _, v := range jobin {
		joblist = append(joblist, Job{
			Status:  Status(v), //todo fill it
			Name:    v.GetId().GetName(),
			Version: v.GetBuild().GetVersion(),
		})
	}

	//sort by anme
	sort.Sort(byName(joblist))

	// now build up the job matrix

	//nrows is the "squarest" number of rows
	flen := float64(len(joblist))
	// square would be n so that n*n~ flen
	// but I want to ensure a "good" proportion:
	// n* prop *n ~ flen => n² ~ flen/prop
	fn := math.Floor(math.Sqrt(flen / *prop))

	nrows := int(math.Floor(fn)) // unfortunately
	if nrows <= 0 {
		nrows = 1
	}

	// a temproary row is filled up to nrows items
	var row = make([]Job, 0, nrows)
	for i, v := range joblist {
		if i > 0 && i%nrows == 0 { // we have reach the end of the row,
			jobs = append(jobs, row)
			row = make([]Job, 0, nrows)
		}
		row = append(row, v)
	}
	jobs = append(jobs, row)
	return
}

func Status(j *format.Job) string {

	var zero, rstart, rend, bstart, bend time.Time

	rstart = time.Unix(j.Refresh.GetStart(), 0)
	rend = time.Unix(j.Refresh.GetEnd(), 0)
	bstart = time.Unix(j.Build.GetStart(), 0)
	bend = time.Unix(j.Build.GetEnd(), 0)

	switch {
	case zero.Equal(rstart) || zero.Equal(bstart) || rstart.After(rend) || bstart.After(bend):
		return "running"
	case j.GetRefresh().GetErrcode() == 0 && j.GetBuild().GetErrcode() == 0:
		return "success"
	default:
		return "failed"
	}
}

//GetJobs just make the http request
func GetJobs() ([]*format.Job, error) {
	req := &format.Request{List: &format.ListRequest{}}

	c := format.NewClient(*server)
	resp, err := c.Proto(req)
	if err != nil {
		return nil, err
	}
	return resp.GetList().GetJobs(), nil
}

//byName to sort any slice of Execution by their Name !
type byName []Job

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
