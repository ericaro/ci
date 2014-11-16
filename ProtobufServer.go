package ci

import (
	"github.com/ericaro/ci/format"
	"net/http"
)

//ProtobufServer is an independent http server that just exposes an http protobuf protocol
type ProtobufServer struct {
	daemon Daemon
}

func NewProtobufServer(daemon Daemon) *ProtobufServer { return &ProtobufServer{daemon} }

func (s *ProtobufServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// read any command
	q := new(format.Request)
	err := format.RequestDecode(q, r)
	if err != nil {
		http.Error(w, "http body must be of type protobuf Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// process the pb request
	resp := s.Execute(q)

	err = format.ResponseWriterEncode(w, resp)
	if err != nil {
		http.Error(w, "unexpected error formatting the response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

//Execute actually run the service transform a request into a response.
// This is not a generic request/response protocol, the request is actually specific
// to the ci operations.
func (s *ProtobufServer) Execute(q *format.Request) *format.Response {
	daemon := s.daemon
	switch {
	case q.List != nil:
		l := daemon.ListJobs(q.List.GetRefreshResult(), q.List.GetRefreshResult())
		return &format.Response{List: l}

	case q.Log != nil:
		j := daemon.JobDetails(q.Log.GetJobname())
		return &format.Response{Log: j}

	case q.Add != nil:
		j := q.Add.Id
		err := daemon.AddJob(j.GetName(), j.GetRemote(), j.GetBranch())
		if err != nil {
			msg := err.Error()
			return &format.Response{Error: &msg}
		}
		// shedule a run after an Add
		daemon.Run()
		return &format.Response{}
	case q.Remove != nil:
		err := daemon.RemoveJob(q.Remove.GetJobname())
		if err != nil {
			msg := err.Error()
			return &format.Response{Error: &msg}
		}
		return &format.Response{}
	}
	return nil
}
