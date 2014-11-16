package main

import (
	"flag"
	"fmt"
	"github.com/ericaro/ci"
	"log"
	"net/http"
	"os"
)

var (
	dbfile = flag.String("o", "ci.db", "override the default local file name")
	port   = flag.Int("p", 2020, "override the default local port")
)

func main() {
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("port:%v\n", port)

	log.Fatal(ListenAndServe(wd, *dbfile, *port))
}

func ListenAndServe(wd, dbfile string, port int) (err error) {
	daemon, err := ci.NewDaemon(wd, dbfile)
	if err != nil {
		log.Printf("error.startup:%q", err.Error())
		return err
	}

	//launch the hook server in an independent gorutine.
	go func() {
		port := 8080
		hook := ci.NewHookServer(daemon)
		log.Printf("startup.hookserver:%v", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), hook))
	}()

	log.Printf("startup.protoserver:%v", port)
	pbs := ci.NewProtobufServer(daemon)
	return http.ListenAndServe(fmt.Sprintf(":%v", port), pbs)

}
