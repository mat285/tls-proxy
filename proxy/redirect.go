package proxy

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/blend/go-sdk/logger"
)

type HTTPRedirect struct {
	Log    logger.Log
	Config RedirectConfig
	Server *http.Server
}

func NewRedirect(log logger.Log, cfg RedirectConfig) *HTTPRedirect {
	h := &HTTPRedirect{
		Log:    log,
		Config: cfg,
	}
	server := &http.Server{
		Addr:    BindAddr(cfg.Port),
		Handler: h,
	}
	h.Server = server
	return h
}

func (hr *HTTPRedirect) Start() error {
	hr.Log.Infof("Starting Redirect Server on %s", hr.Server.Addr)
	hr.Server.Handler = hr
	return hr.Server.ListenAndServe()
}

func (hr *HTTPRedirect) Stop() error {
	hr.Log.Infof("Stopping Redirect Server")
	err := hr.Server.Shutdown(context.Background())
	hr.Log.Infof("Redirect Server stopped")
	return err

}

func (hr *HTTPRedirect) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	original := req.URL.String()
	req.URL.Scheme = "https"
	host, err := hr.replacePort(req.Host, hr.Config.UpstreamPort)
	if err != nil {
		hr.Log.Errorf("Bad host for redirect request %s", original)
		if len(req.Host) == 0 {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	req.URL.Host = host
	req.Host = host
	hr.Log.Infof("Redirecting request for %s to %s", original, req.URL.String())
	http.Redirect(rw, req, req.URL.String(), http.StatusMovedPermanently)
}

func (hr *HTTPRedirect) replacePort(host string, port uint16) (string, error) {
	return replacePort(host, port)
}

func replacePort(host string, port uint16) (string, error) {
	if len(host) == 0 {
		return "", fmt.Errorf("bad host")
	}
	parts := strings.Split(host, ":")
	if len(parts) == 0 {
		return fmt.Sprintf(":%d", port), nil
	}
	if len(parts[0]) == 0 {
		return "", fmt.Errorf("bad host")
	}
	return fmt.Sprintf("%s:%d", parts[0], port), nil
}
