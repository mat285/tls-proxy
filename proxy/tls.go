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
	"strconv"

	"github.com/blend/go-sdk/logger"
)

type TLSServer struct {
	Log          logger.Log
	Config       TLSConfig
	Server       *http.Server
	ReverseProxy *httputil.ReverseProxy
	upstreamPort uint16
}

func NewTLSServer(log logger.Log, cfg TLSConfig) (*TLSServer, error) {
	target, err := url.Parse(cfg.Upstream)
	if err != nil {
		return nil, err
	}
	up := uint16(0)
	parsed, err := strconv.ParseUint(target.Port(), 10, 16)
	if err == nil {
		up = uint16(parsed)
	}
	log.Debugf("proxying to target %s", target.String())
	rev := httputil.NewSingleHostReverseProxy(target)
	t := &TLSServer{
		Config:       cfg,
		ReverseProxy: rev,
		Log:          log,
		upstreamPort: up,
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
	if s.upstreamPort != 0 {
		host, err := replacePort(req.Host, s.upstreamPort)
		if err == nil {
			req.Host = host
		}
	}
	s.ReverseProxy.ServeHTTP(rw, req)
}

func (s *TLSServer) rewritePort(pr *httputil.ProxyRequest) {

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
