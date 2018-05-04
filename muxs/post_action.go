package router

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"github.com/jdmelo/gotoolkit/utils"
)

type Request struct {
	RawReq      *http.Request
	Action      string
	RequestId   string
	UserId      string
	ClientToken string
	AdminToken  string
}

func NewPostActionMux(pattern, hostname string) *PostActionMux {

	cm := &PostActionMux{
		pattern:    pattern,
		handleFunc: make(map[string]HandleFunc),
		hostname:   hostname,
	}

	return cm
}

type HandleFunc func(w http.ResponseWriter, r *Request)

type PostActionMux struct {
	mu         sync.RWMutex
	handleFunc map[string]HandleFunc
	pattern    string
	hostname   string
}

func (p *PostActionMux) RegisterHandleFunc(action string, h HandleFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handleFunc[action] = h
}

func (p *PostActionMux) GetHandleFunc(action string) (HandleFunc, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	h, ok := p.handleFunc[action]
	if !ok {
		return nil, errors.New("handle function not found")
	}

	return h, nil
}

func (p *PostActionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		requestId   string
		clientToken string
		userId      string
		adminToken  string
	)
	if r.URL.Path != p.pattern {
		http.NotFound(w, r)
		return
	}

	action := r.FormValue("Action")
	if action == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	hFunc, err := p.GetHandleFunc(action)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// parser request id
	requestIds, _ := r.Header["Request-Id"]
	if len(requestIds) < 1 || requestIds[0] == "" {
		requestId = utils.GenerateUuid("req")
		w.Header().Set("Request-Id", requestId)
	} else {
		requestId = requestIds[0]
	}

	// parser client_token
	clientTokens, _ := r.Header["Client-Token"]
	if len(clientTokens) > 0 && clientTokens[0] != "" {
		clientToken = clientTokens[0]
	}

	// parser user id
	userIds, _ := r.Header["User-Id"]
	if len(userIds) > 0 && userIds[0] != "" {
		userId = userIds[0]
	}

	// parser adminToken user
	adminTokens, _ := r.Header["AdminToken-Token"]
	if len(adminTokens) > 0 && adminTokens[0] != "" {
		adminToken = adminTokens[0]
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, "request_id", requestId)
	r = r.WithContext(ctx)

	req := &Request{
		RawReq:      r,
		Action:      action,
		RequestId:   requestId,
		UserId:      userId,
		ClientToken: clientToken,
		AdminToken:  adminToken,
	}

	hFunc(w, req)
	// help to add default header
	w.Header().Set("X-Requst-Trace", p.hostname)

	return
}
