package cylon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

type Server struct {
	ai         AI
	remoteRoot string
	localRoot  string
	done       chan bool
}

func NewServer(ai AI, remoteRoot string, localRoot string, done chan bool) *Server {
	s := &Server{
		ai:         ai,
		remoteRoot: remoteRoot,
		localRoot:  localRoot,
		done:       done,
	}

	return s
}

func (s *Server) CreateRouter() (*mux.Router, error) {
	r := mux.NewRouter()
	m := map[string]map[string]HttpApiFunc{
		"GET": {
			"/status": s.status,
		},
		"POST": {
			"/status": s.status,
			"/think":  s.think,
			"/end":    s.end,
			"/start":  s.start,
		},
	}

	for method, routes := range m {
		for route, handler := range routes {
			localRoute := route
			localHandler := handler
			localMethod := method
			f := makeHttpHandler(localMethod, localRoute, localHandler)

			r.Path(localRoute).Methods(localMethod).HandlerFunc(f)
		}
	}

	return r, nil
}

type joinMatchRequestMessage struct {
	Endpoint string `json:"endpoint"`
	Match    string `json:"match"`
}

func (s *Server) Join(match string) error {
	j := &joinMatchRequestMessage{
		Endpoint: s.localRoot,
		Match:    match,
	}

	u, err := url.Parse(s.remoteRoot)
	if err != nil {
		return err
	}

	u.Path = "join"

	log.Println("Making request to ", u.String())
	js, _ := json.Marshal(j)
	r, err := http.Post(u.String(), "application/json", bytes.NewBuffer(js))
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("expected HTTP 200 - OK, got %v", r.StatusCode))
	} else {
		log.Println("OK")
	}

	return nil
}

func makeHttpHandler(localMethod string, localRoute string, handlerFunc HttpApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeCorsHeaders(w, r)
		if err := handlerFunc(w, r, mux.Vars(r)); err != nil {
			httpError(w, err)
		}
	}
}

func writeCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
}

type HttpApiFunc func(w http.ResponseWriter, r *http.Request, vars map[string]string) error

func writeJSON(w http.ResponseWriter, code int, thing interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	val, err := json.Marshal(thing)
	w.Write(val)
	return err
}

func httpError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError

	if err != nil {
		http.Error(w, err.Error(), statusCode)
	}
}

func (s *Server) status(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) start(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) end(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	w.WriteHeader(http.StatusOK)
	s.done <- true
	return nil
}

func (s *Server) think(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	decoder := json.NewDecoder(r.Body)
	state := &RobotState{}

	err := decoder.Decode(state)
	if err != nil {
		log.Println(err)
		return err
	}

	commands := s.ai.Think(state)

	err = writeJSON(w, http.StatusOK, commands)
	if err != nil {
		log.Println(err)
	}

	return nil
}
