package proxy

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Config struct {
	TLS      TLSConfig      `yaml:"tls"`
	Redirect RedirectConfig `yaml:"redirect"`
}

type TLSConfig struct {
	Port     uint16 `yaml:"port"`
	Upstream string `yaml:"upstream"`
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

type RedirectConfig struct {
	Port         uint16 `yaml:"port"`
	UpstreamPort uint16 `yaml:"upstreamPort"`
}

func ReadConfigFile(file string) (*Config, error) {
	if len(file) == 0 {
		return &Config{}, nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	upstream := os.Getenv("REDIRECT_UPSTREAM_PORT")
	if len(upstream) > 0 {
		uPort64, err := strconv.ParseUint(upstream, 10, 16)
		if err == nil {
			cfg.Redirect.UpstreamPort = uint16(uPort64)
		}
	}

	if cfg.Redirect.UpstreamPort == 0 {
		// default https
		cfg.Redirect.UpstreamPort = uint16(443)
	}

	upstreamTLS := os.Getenv("TLS_UPSTREAM")
	if len(upstreamTLS) > 0 {
		cfg.TLS.Upstream = upstreamTLS
	}

	if cfg.Redirect.UpstreamPort == 0 {
		cfg.Redirect.UpstreamPort = cfg.TLS.Port
	}

	_, err = url.Parse(cfg.TLS.Upstream)
	if err != nil {
		return nil, err
	}

	err = LoadTLS(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadTLS(cfg *Config) error {
	if len(cfg.TLS.CertFile) == 0 {
		return fmt.Errorf("Missing TLS Cert File")
	}
	if len(cfg.TLS.KeyFile) == 0 {
		return fmt.Errorf("Missing TLS Key File")
	}

	certData, err := os.ReadFile(cfg.TLS.CertFile)
	if err != nil {
		return err
	}
	certData = certData
	// _, err = ParsePemEd25519PublicKey(certData)
	// if err != nil {
	// 	return err
	// }

	keyData, err := os.ReadFile(cfg.TLS.KeyFile)
	if err != nil {
		return err
	}
	keyData = keyData
	// _, err = ParsePemEd25519PrivateKey(keyData)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func ConfigOrDefault(cfg Config) Config {
	if cfg.Redirect.Port == 0 && cfg.TLS.Port == 0 {
		cfg.Redirect.Port = 2021
		cfg.TLS.Port = 2022
	}

	if cfg.Redirect.Port == 0 {
		cfg.Redirect.Port = cfg.TLS.Port + 1
	}
	if cfg.TLS.Port == 0 {
		cfg.TLS.Port = cfg.Redirect.Port + 1
	}
	return cfg
}
