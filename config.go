// simplecert
//
// Created by Philipp Mieden
// Contact: dreadl0ck@protonmail.ch
// Copyright © 2018 bestbytes. All rights reserved.
package simplecert

import (
	"errors"
	"log"
	"os"
	"time"
)

type KeyType string

const (
	EC256   = "P256"
	EC384   = "P384"
	RSA2048 = "2048"
	RSA4096 = "4096"
	RSA8192 = "8192"
)

var (
	c *Config

	errNoDirectoryURL     = errors.New("simplecert: no directory url specified in config")
	errNoMail             = errors.New("simplecert: no SSLEmail in config in config")
	errNoDomains          = errors.New("simplecert: no domains specified in config")
	errNoChallenge        = errors.New("simplecert: no challenge method specified in config")
	errNoCacheDir         = errors.New("simplecert: no cache directory specified in config")
	errNoRenewBefore      = errors.New("simplecert: no renew before value set in config")
	errNoCheckInterval    = errors.New("simplecert: no check interval set in config")
	errNoCacheDirPerm     = errors.New("simplecert: no cache directory permission specified in config")
	errUnsupportedKeyType = errors.New("simplecert: unsupported key type specified in config")

	supportedKeyTypes = map[string]bool{
		EC256:   true,
		EC384:   true,
		RSA2048: true,
		RSA4096: true,
		RSA8192: true,
	}
)

// Default contains a default configuration
var Default = &Config{
	// 30 Days before expiration
	RenewBefore: 30 * 24,
	// every two days
	CheckInterval: 2 * 24 * time.Hour,
	SSLEmail:      "",
	DirectoryURL:  "https://acme-v02.api.letsencrypt.org/directory",
	HTTPAddress:   ":80",
	TLSAddress:    ":443",
	CacheDirPerm:  0700,
	Domains:       []string{},
	CacheDir:      "letsencrypt",
	DNSProvider:   "",
	Local:         false,
	UpdateHosts:   true,
	DNSServers:    []string{},
	KeyType:       RSA2048,
}

// Config allows configuration of simplecert
type Config struct {
	// renew the certificate X hours before it expires
	// LetsEncrypt Certs are valid for 90 Days
	RenewBefore int

	// Interval for checking if cert is closer to expiration than RenewBefore
	CheckInterval time.Duration

	// SSLEmail for contact
	SSLEmail string

	// ACME Directory URL. Can be set to https://acme-staging-v02.api.letsencrypt.org/directory for testing
	DirectoryURL string

	// Endpoints for webroot challenge
	// CAUTION: challenge must be received on port 80 and 443
	// if you choose different ports here you must redirect the traffic
	HTTPAddress string

	TLSAddress string

	// UNIX Permission for the CacheDir and all files inside
	CacheDirPerm os.FileMode

	// Domains for which to obtain the certificate
	Domains []string

	// DNSServers overrides the dns resolvers to use for a dns challenge, this is handy if you have a split dns.
	DNSServers []string

	// Path of the CacheDir
	CacheDir string

	// DNSProvider name for DNS challenges (optional)
	// see: https://godoc.org/github.com/go-acme/lego/providers/dns
	DNSProvider string

	// Local runmode
	Local bool

	// UpdateHosts adds the domains to /etc/hosts if running in local mode
	UpdateHosts bool

	// KeyType represents the key algorithm as well as the key size or curve to use.
	KeyType string

	// Handler funcs for graceful service shutdown and restoring
	WillRenewCertificate func()

	DidRenewCertificate      func()
	FailedToRenewCertificate func(error)
}

// CheckConfig checks if config can be used to obtain a cert
func CheckConfig(c *Config) error {
	if c.CacheDir == "" {
		return errNoCacheDir
	}
	if len(c.Domains) == 0 {
		return errNoDomains
	}
	if !c.Local {
		if c.SSLEmail == "" {
			return errNoMail
		}
	}
	if c.DirectoryURL == "" {
		return errNoDirectoryURL
	}

	if c.DNSProvider == "" && c.HTTPAddress == "" && c.TLSAddress == "" {
		return errNoChallenge
	}

	if c.RenewBefore == 0 {
		return errNoCacheDir
	}

	if c.CheckInterval == 0 {
		return errNoCheckInterval
	}

	if c.CacheDirPerm == 0 {
		return errNoCacheDirPerm
	}

	if !supportedKeyTypes[c.KeyType] {
		return errUnsupportedKeyType
	}

	if c.WillRenewCertificate == nil && (c.HTTPAddress != "" || c.TLSAddress != "") {
		log.Println("[WARNING] no WillRenewCertificate handler specified, to handle graceful server shutdown!")
	}
	if c.DidRenewCertificate == nil && (c.HTTPAddress != "" || c.TLSAddress != "") {
		log.Println("[WARNING] no DidRenewCertificate handler specified, to bring the service back up after renewing the certificate!")
	}
	if c.FailedToRenewCertificate == nil {
		log.Println("[WARNING] no FailedToRenewCertificate handler specified! Simplecert will fatal on errors!")
	}

	return nil
}
