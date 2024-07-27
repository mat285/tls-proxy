package proxy

import (
	"fmt"
	"strings"
	"sync"

	"github.com/blend/go-sdk/logger"
)

type Proxy struct {
	Config         Config
	Log            logger.Log
	TLSServer      Runnable
	RedirectServer Runnable

	lock    sync.Mutex
	running bool
	stopped chan struct{}
}

type Runnable interface {
	Start() error
	Stop() error
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
	toRun := []Runnable{}
	p.TLSServer, err = NewTLSServer(p.Log, p.Config.TLS)
	if err != nil {
		p.lock.Unlock()
		return err
	}
	toRun = append(toRun, p.TLSServer)
	if p.Config.Redirect.Enabled {
		p.RedirectServer = NewRedirect(p.Log, p.Config.Redirect)
		toRun = append(toRun, p.RedirectServer)
	}

	p.stopped = make(chan struct{})
	p.running = true
	errs := make(chan error, len(toRun))
	for _, run := range toRun {
		go func(run Runnable) {
			errs <- run.Start()
		}(run)
	}

	p.lock.Unlock()

	errStrings := make([]string, 0)
	err = <-errs
	if err != nil {
		errStrings = append(errStrings, err.Error())
	}

	for _, run := range toRun {
		go run.Stop()
	}

	for i := 1; i < len(toRun); i++ {
		err = <-errs
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	close(errs)

	p.lock.Lock()
	p.running = false
	close(p.stopped)
	p.lock.Unlock()

	if len(errStrings) > 0 {
		return fmt.Errorf(strings.Join(errStrings, "\n"))
	}
	return nil
}

func (p *Proxy) Stop() error {
	if !p.running {
		return nil
	}
	errs := make(chan error, 2)
	tlsServer := p.TLSServer
	redirServer := p.RedirectServer
	var wg sync.WaitGroup
	if tlsServer != nil {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			errs <- tlsServer.Stop()
		}(&wg)
	}
	if redirServer != nil {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			errs <- redirServer.Stop()
		}(&wg)
	}

	<-p.stopped
	wg.Wait()
	close(errs)
	errStrings := make([]string, 0)
	for err := range errs {
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	if len(errStrings) > 0 {
		return fmt.Errorf(strings.Join(errStrings, "\n"))
	}
	return nil
}

func BindAddr(port uint16) string {
	return fmt.Sprintf("0.0.0.0:%d", port)
}
