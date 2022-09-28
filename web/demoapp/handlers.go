package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

// reference and cache the templates
var templates = template.Must(
	template.ParseFiles(
		fmt.Sprintf("%s/index.html", config().TemplatesDir)))

// handler func index / default page
func index(w http.ResponseWriter, r *http.Request) {
	d := struct {
		Title   string
		AppName string
		Host    string
		BgColor string
		FgColor string
	}{
		Title:   config().AppName,
		AppName: config().AppName,
		Host:    getHostName(),
		BgColor: config().BgColor,
		FgColor: config().FgColor,
	}
	err := templates.ExecuteTemplate(w, "index", &d)
	if err != nil {
		http.Error(w, fmt.Sprintf("index: couldn't parse template: %v", err), http.StatusInternalServerError)
		return
	}
	// automatically returns OK
}

// handler func hello for testing (/hello?name=alex)
func hello(w http.ResponseWriter, r *http.Request) {
	q, ok := r.URL.Query()["name"]
	if ok {
		fmt.Fprintf(w, "<h1>Salut, Bonjour  %s!</h1>", string(q[0]))
	} else {
		fmt.Fprint(w, "<h1>Salut, Bonjour!</h1>")
	}
}

func getHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		log.Println(err)
		return "<<ERROR getting hostname>>"
	}
	return hostName
}
