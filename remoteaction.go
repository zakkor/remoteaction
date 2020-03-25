package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/alessio/shellescape"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Server struct {
	Config Config
}

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
	var (
		configPath = flag.String("c", "./config.json", "path to config file")
		user       = flag.String("user", "", "username to use for authentication")
		pass       = flag.String("pass", "", "password to use for authentication")
	)
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

	s := &Server{Config: config}

	menusData, err := json.Marshal(config.Menus)
	if err != nil {
		log.Fatalln(err)
	}

	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.Timeout(2 * time.Second))

	if *pass != "" {
		r.Use(AuthOnly(*user, *pass))
	}

	// Serve menus JSON data.
	r.Get("/menus", func(w http.ResponseWriter, r *http.Request) {
		w.Write(menusData)
	})

	// Execute a command when an action happens.
	r.Post("/action", s.ActionHandler)

	err = http.ListenAndServe(":6969", r)
	if err != nil {
		log.Fatalln(err)
	}
}

func (s *Server) ActionHandler(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	data := r.URL.Query().Get("data")
	data, err := url.QueryUnescape(data)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	data = shellescape.Quote(data)

	cmdTempl, ok := s.Config.Commands[action]
	if !ok {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// Parse and execute template
	t := template.Must(template.New(action).Parse(cmdTempl))
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	cmd := buf.String()
	log.Println(cmd)
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Write(out)
}

func AuthOnly(user, pass string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t := r.Header.Get("Authorization")
			if t == "" {
				http.Error(w, http.StatusText(403), 403)
				return
			}
			st := strings.Split(t, "Basic")
			if len(st) != 2 {
				http.Error(w, http.StatusText(403), 403)
				return
			}
			t = strings.TrimSpace(st[1])
			data, err := base64.StdEncoding.DecodeString(t)
			if err != nil {
				http.Error(w, http.StatusText(403), 403)
				return
			}
			sd := strings.Split(string(data), ":")
			var gotuser, gotpass string
			// If only one part was specified, we assume it is the password
			if len(sd) == 1 {
				gotpass = sd[0]
			} else if len(sd) == 2 {
				gotuser = sd[0]
				gotpass = sd[1]
			} else {
				http.Error(w, http.StatusText(403), 403)
				return
			}

			if gotuser != user || gotpass != pass {
				http.Error(w, http.StatusText(403), 403)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
