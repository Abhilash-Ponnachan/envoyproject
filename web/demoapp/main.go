package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// multiplexer to route req to hanlder functions
	muxHandler := http.NewServeMux()

	// handle root to serve main page
	muxHandler.HandleFunc("/", index)
	// handler to serve 'static' assets
	fs := http.FileServer(http.Dir("."))
	muxHandler.Handle("/assets/", fs)

	// handler func to test server with 'hello'
	muxHandler.HandleFunc("/hello", hello)

	// handler funcs for other path (namely APIs)
	muxHandler.HandleFunc("/api/info", info)

	// http server instance
	server := http.Server{
		Addr: fmt.Sprintf(":%s", config().Port),
		// our mux as handler for http requests
		Handler:      muxHandler,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 5,
		IdleTimeout:  time.Second * 10,
	}
	log.Printf("Starting http server on PORT: %s\n", config().Port)
	// start server and listen for req to serve
	// this is a 'blocking' call
	err := server.ListenAndServe()
	// check type of termination error
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}

}
