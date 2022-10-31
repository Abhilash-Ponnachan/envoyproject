package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
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

// handler func for get info API
func info(w http.ResponseWriter, r *http.Request) {
	dt := struct {
		ReqHdrs  http.Header
		ClientIP string
		ServerIP string
		EnvSub   []string
	}{
		ReqHdrs:  r.Header,
		ClientIP: readClientIP(r),
		ServerIP: getHostIP(),
		EnvSub:   filterEnvVars(),
	}
	js, err := json.Marshal(dt)
	log.Printf("js = %s\n", js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
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

// helper functions ...
func getHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		log.Println(err)
		return "<<ERROR getting hostname>>"
	}
	return hostName
}

func readClientIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

func getHostIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return fmt.Sprintf("Error in getIP : %s \n", err.Error())
	}
	defer conn.Close()

	return conn.LocalAddr().String()
}

// filtered subset of env vars
func filterEnvVars() []string {
	evs := os.Environ()
	fevs := []string{}
	for _, ev := range evs {
		kv := strings.Split(ev, "=")
		if len(kv) > 0 {
			switch kv[0] {
			case portEnv,
				appNameEnv,
				bgColorEnv,
				fgColorEnv,
				"HOST":
				fevs = append(fevs, ev)
			}
		}
	}
	return fevs
}
