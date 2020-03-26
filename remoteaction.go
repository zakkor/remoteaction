package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/philandstuff/dhall-golang"
)

var (
	configPath = flag.String("c", "./config.json", "path to config file")
)

type Server struct {
	mu       sync.Mutex
	Config   Config
	MenuData []byte
	Commands map[string]string
	sem      chan bool
}

type RemoteMenu struct {
	Context  string         `json:"context"`
	Patterns []Pattern      `json:"patterns"`
	Actions  []RemoteAction `json:"actions"`
}

type RemoteAction struct {
	Name   string `json:"name"`
	Action string `json:"action"`
}

func main() {
	flag.Parse()

	s := &Server{}
	err := s.readConfig()
	if err != nil {
		log.Fatalln(err)
	}

	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.Timeout(2 * time.Second))
	r.Use(middleware.Logger)

	if len(s.Config.Auth.Username) > 0 || len(s.Config.Auth.Password) > 0 {
		r.Use(Authenticate(s.Config.Auth.Username, s.Config.Auth.Password))
	}

	// Serve menus JSON data.
	r.Get("/menus", s.getMenus)
	// Execute a command when an action happens.
	r.Post("/action", s.actionHandler)

	err = http.ListenAndServe(s.Config.Listen, r)
	if err != nil {
		log.Fatalln(err)
	}
}

func (s *Server) readConfig() error {
	// Read config file data and unmarshal into struct
	configData, err := ioutil.ReadFile(*configPath)
	if err != nil {
		return err
	}

	var config Config
	if strings.HasSuffix(*configPath, ".json") {
		err = json.Unmarshal(configData, &config)
	} else if strings.HasSuffix(*configPath, ".dhall") {
		err = dhall.Unmarshal(configData, &config)
	} else {
		err = errors.New("invalid config extension, dhall and json supported")
	}
	if err != nil {
		return err
	}

	var menus []RemoteMenu
	commands := map[string]string{}
	for i, menu := range config.Menus {
		var context string
		var patterns []Pattern
		var actions []Action
		switch {
		case menu.Link != nil:
			context = "link"
			patterns = menu.Link.Patterns
			actions = menu.Link.Actions

		case menu.Selection != nil:
			context = "selection"
			patterns = make([]Pattern, len(menu.Selection.Regexes))
			for i, r := range menu.Selection.Regexes {
				patterns[i] = Pattern{Regex: r}
			}
			actions = menu.Selection.Actions

		default:
			log.Fatalf("invalid menu type @ menu %d", i)
		}

		var ractions []RemoteAction
		for _, action := range actions {
			hash := sha256.Sum256([]byte(action.Script))
			hashStr := hex.EncodeToString(hash[:])

			ractions = append(ractions, RemoteAction{
				Name:   action.Name,
				Action: hashStr,
			})

			commands[hashStr] = action.Script
		}

		menus = append(menus, RemoteMenu{
			Context:  context,
			Patterns: patterns,
			Actions:  ractions,
		})

	}

	menuData, err := json.Marshal(menus)
	if err != nil {
		return err
	}

	executors := config.Executors
	if executors <= 0 {
		executors = 100
	}

	s.mu.Lock()
	s.Config = config
	s.MenuData = menuData
	s.Commands = commands
	s.sem = make(chan bool, executors)
	s.mu.Unlock()

	return nil
}

func (s *Server) getMenus(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	data := s.MenuData
	s.mu.Unlock()

	w.Write(data)
}

func (s *Server) actionHandler(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	s.mu.Lock()
	sem := s.sem
	script, ok := s.Commands[action]
	s.mu.Unlock()
	if !ok {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	data := r.URL.Query().Get("data")
	data, err := url.QueryUnescape(data)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	go func() {
		sem <- true
		defer func() { <-sem }()

		cmd := exec.Command("bash", "-c", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), "LINK="+data)
		cmd.Run()
	}()
}

func Authenticate(user, pass string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 {
				http.Error(w, http.StatusText(403), 403)
				return
			}

			if parts[0] != "Basic" {
				http.Error(w, http.StatusText(403), 403)
				return
			}

			data, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				http.Error(w, http.StatusText(403), 403)
				return
			}

			var ok bool
			authParts := strings.Split(string(data), ":")
			if len(authParts) == 1 {
				ok = authParts[0] == pass
			} else {
				ok = authParts[0] == user && authParts[1] == pass
			}

			if !ok {
				http.Error(w, http.StatusText(403), 403)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
