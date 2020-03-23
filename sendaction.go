package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"text/template"

	"github.com/go-chi/chi"
)

type Config struct {
	Menus    []Menu            `json:"menus"`
	Commands map[string]string `json:"commands"`
}

type Menu struct {
	Contexts []string `json:"contexts"`
	Action   string   `json:"action"`
	Name     string   `json:"name"`
	Patterns []string `json:"patterns"`
	Regexes  []string `json:"regexes"`
}

func main() {
	configPath := flag.String("c", "path to config file", "./config.json")
	flag.Parse()

	// Read config file data and unmarshal into struct
	configData, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalln(err)
	}
	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalln(err)
	}

	menusData, err := json.Marshal(config.Menus)
	if err != nil {
		log.Fatalln(err)
	}

	r := chi.NewRouter()
	// Serve menus JSON data.
	r.Get("/menus", func(w http.ResponseWriter, r *http.Request) {
		w.Write(menusData)
	})
	// Execute a command when an action happens.
	r.Post("/action", func(w http.ResponseWriter, r *http.Request) {
		// TODO: better error handling in here

		action := r.URL.Query().Get("action")
		data := r.URL.Query().Get("data")
		data, err := url.QueryUnescape(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cmdTempl, ok := config.Commands[action]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Parse and execute template
		t := template.Must(template.New(action).Parse(cmdTempl))
		var buf bytes.Buffer
		err = t.Execute(&buf, data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Split template output into [cmd, args...]
		parts := strings.Split(buf.String(), " ")
		cmd := parts[0]
		var args []string
		if len(parts) > 1 {
			args = parts[1:]
		}

		out, err := exec.Command(cmd, args...).Output()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(out)
	})
	http.ListenAndServe(":6969", r)
}
