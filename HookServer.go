package ci

import (
	"net/http"
)

type HookServer struct {
	daemon Daemon
}

func NewHookServer(daemon Daemon) *HookServer { return &HookServer{daemon} }

//ServeHTTP defines the http server
func (s *HookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// any incoming request here is a valid hook, and triggers a refresh / build
	daemon := s.daemon
	daemon.Run()
	// later, with an execution queue model, I'd rather just schedule a refresh, and a build afterwards.
	// incoming hooks might come in groups. and only one refresh should be made.
	//
}
