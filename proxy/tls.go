package proxy

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/blend/go-sdk/logger"
)

type TLSServer struct {
	Log          logger.Log
	Config       TLSConfig
	Server       *http.Server
	ReverseProxy *httputil.ReverseProxy
}

func NewTLSServer(log logger.Log, cfg TLSConfig) (*TLSServer, error) {
	target, err := url.Parse(cfg.Upstream)
	if err != nil {
		return nil, err
	}
	rev := httputil.NewSingleHostReverseProxy(target)
	t := &TLSServer{
		Config:       cfg,
		ReverseProxy: rev,
		Log:          log,
	}
	serv := &http.Server{
		Addr:    BindAddr(cfg.Port),
		Handler: t,
	}
	t.Server = serv
	return t, nil
}

func (s *TLSServer) Start() error {
	s.Log.Infof("Starting TLS Server on %s", s.Server.Addr)
	return s.Server.ListenAndServeTLS(s.Config.CertFile, s.Config.KeyFile)
}
func (s *TLSServer) Stop() error {
	s.Log.Infof("Stopping TLS Server")
	err := s.Server.Shutdown(context.Background())
	s.Log.Infof("TLS Server stopped")
	return err
}

func (s *TLSServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.Log.Infof("Proxying request for %s", req.URL.String())
	s.ReverseProxy.ServeHTTP(rw, req)
}

func ParsePemEd25519PrivateKey(data []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(data)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	typed, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("wrong key type")
	}
	return typed, nil
}

func ParsePemEd25519PublicKey(data []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(data)
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	typed, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("wrong key type")
	}
	return typed, nil
}
