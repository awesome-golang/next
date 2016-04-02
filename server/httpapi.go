package server

import (
	"encoding/json"
	"net/http"
	"time"

	"gopkg.in/logex.v1"

	"github.com/chzyer/next/crypto"
	"github.com/chzyer/next/ip"
	"github.com/chzyer/next/uc"
	"github.com/chzyer/next/util/clock"
)

type HttpApiConfig struct {
	AesKey []byte

	CertFile string
	KeyFile  string
}

type HttpApi struct {
	listen   string
	cfg      *HttpApiConfig
	users    *uc.Users
	clock    *clock.Clock
	server   *http.Server
	delegate HttpDelegate
}

type HttpDelegate interface {
	AllocIP() *ip.IP
	GetGateway() *ip.IPNet
	GetMTU() int
	GetDataChannel() int
	OnNewUser(userId int)
}

func NewHttpApi(listen string, users *uc.Users, ct *clock.Clock, cfg *HttpApiConfig, delegate HttpDelegate) *HttpApi {
	server := &http.Server{
		Addr:           listen,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 10,
	}
	return &HttpApi{
		cfg:      cfg,
		clock:    ct,
		server:   server,
		listen:   listen,
		users:    users,
		delegate: delegate,
	}
}

type replyError struct {
	Error string `json:"error"`
}

func (h *HttpApi) replyError(w http.ResponseWriter, err interface{}) {
	w.WriteHeader(400)
	switch t := err.(type) {
	case error:
		h.reply(w, replyError{t.Error()})
	case string:
		h.reply(w, replyError{t})
	}
}

func (h *HttpApi) reply(w http.ResponseWriter, obj interface{}) {
	ret, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	crypto.EncodeAes(ret, ret, h.cfg.AesKey, nil)

	n, err := w.Write(ret)
	if n != len(ret) {
		logex.Errorf("short write: %v, want: %v", n, len(ret))
		return
	}
	if err != nil {
		logex.Error(err)
	}
}

func (h *HttpApi) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/", h.Auth)
	mux.HandleFunc("/time/", h.Time)
	h.server.Handler = mux
	return h.server.ListenAndServe()
}
