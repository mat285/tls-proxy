package proxy

import (
	"fmt"
	"sync"

	"github.com/blend/go-sdk/logger"
)

type Proxy struct {
	Config         Config
	Log            logger.Log
	TLSServer      *TLSServer
	RedirectServer *HTTPRedirect

	lock    sync.Mutex
	running bool
	stopped chan struct{}
}

func NewProxyFromFile(file string) (*Proxy, error) {
	cfg, err := ReadConfigFile(file)
	if err != nil {
		return nil, err
	}
	return &Proxy{
		Config:  ConfigOrDefault(*cfg),
		Log:     logger.All(),
		running: false,
	}, nil
}

func (p *Proxy) Start() error {
	if p.running {
		return fmt.Errorf("already running")
	}
	p.lock.Lock()
	if p.running {
		p.lock.Unlock()
		return fmt.Errorf("already running")
	}

	var err error
	p.TLSServer, err = NewTLSServer(p.Log, p.Config.TLS)
	if err != nil {
		p.lock.Unlock()
		return err
	}
	p.RedirectServer = NewRedirect(p.Log, p.Config.Redirect)

	p.stopped = make(chan struct{})
	p.running = true
	errs := make(chan error, 2)
	go func() {
		errs <- p.TLSServer.Start()
	}()
	go func() {
		errs <- p.RedirectServer.Start()
	}()

	p.lock.Unlock()

	err1 := <-errs
	go p.TLSServer.Stop()
	go p.RedirectServer.Stop()
	err2 := <-errs

	p.lock.Lock()
	p.running = false
	close(p.stopped)
	p.lock.Unlock()

	if err1 == nil && err2 == nil {
		return nil
	}
	if err1 != nil {
		return err1
	}
	return err2
}

func (p *Proxy) Stop() error {
	if !p.running {
		return nil
	}
	go p.TLSServer.Stop()
	go p.RedirectServer.Stop()
	<-p.stopped
	return nil
}

func BindAddr(port uint16) string {
	return fmt.Sprintf("0.0.0.0:%d", port)
}
