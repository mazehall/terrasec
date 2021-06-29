package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	version = "v0.1.0 terraform-backend-gopass"
)

type Http struct {
	Port       string
	Kind       string
	ConfigFile string
}

func (h Http) Run() (url string, err error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tfstate", h.handleState)
	mux.HandleFunc("/", versionPrint)
	srv := &http.Server{}
	srv.Handler = mux

	errorCh := make(chan error, 1)
	addrCh := make(chan string, 1)
	addr := "127.0.0.1" + h.Port
	go func() {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			errorCh <- err
		}
		addrCh <- l.Addr().String()
		errorCh <- http.Serve(l, mux)
	}()
	for {
		select {
		case addr = <-addrCh:
		case serverError := <-errorCh:
			return "", serverError
		case <-time.After(time.Millisecond * 300):
			return addr, nil
		}
	}
}

func versionPrint(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, version)
}

func (h Http) handleState(response http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "LOCK":
		h.lockState(request, response)
	case "UNLOCK":
		h.unlockState(request, response)
	case http.MethodGet:
		h.getState(request, response)
	case http.MethodPost:
		h.postState(request, response)
	case http.MethodDelete:
		h.deleteState(request, response)
	default:
		response.WriteHeader(http.StatusNotImplemented)
	}
}

func (h Http) lockState(request *http.Request, response http.ResponseWriter) {
	// fmt.Println("Lock state")
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	// fmt.Println(string(body))
	r, err := GetRepo(h.ConfigFile, h.Kind)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := r.lockState(body); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusOK)
}

func (h Http) unlockState(request *http.Request, response http.ResponseWriter) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	// fmt.Println(string(body))
	r, err := GetRepo(h.ConfigFile, h.Kind)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := r.unlockState(body); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	response.WriteHeader(http.StatusOK)
}

func (h Http) getState(request *http.Request, response http.ResponseWriter) {
	r, err := GetRepo(h.ConfigFile, h.Kind)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	content, err := r.getState()
	if err != nil {
		fmt.Println(err)
		response.WriteHeader(http.StatusNoContent)
		return
	}
	// fmt.Println(content)
	var j interface{}
	if err = json.Unmarshal(content, &j); err != nil {
		fmt.Println(err)
		response.WriteHeader(http.StatusNoContent)
		return
	}
	if _, err := response.Write(content); err != nil {
		fmt.Println(err)
		response.WriteHeader(http.StatusNoContent)
		return
	}
}

func (h Http) postState(request *http.Request, response http.ResponseWriter) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	// fmt.Println(string(body))
	r, err := GetRepo(h.ConfigFile, h.Kind)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := r.saveState(body); err != nil {
		fmt.Println(err)
		response.WriteHeader(http.StatusNoContent)
		return
	}

	response.WriteHeader(http.StatusOK)
}

func (h Http) deleteState(request *http.Request, response http.ResponseWriter) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(string(body))

	response.WriteHeader(http.StatusOK)
}
