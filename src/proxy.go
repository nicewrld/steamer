package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func CreateHTTPClient(config Config) *http.Client {
	var proxyFunc func(*http.Request) (*url.URL, error)

	if !config.DisableProxy {
		proxyURL, _ := url.Parse(fmt.Sprintf(
			"http://%s:%s@%s:%s",
			config.ProxyUser, config.ProxyPassword, config.ProxyHost, config.ProxyPort,
		))
		proxyFunc = http.ProxyURL(proxyURL)
	} else {
		proxyFunc = nil // Do not use a proxy
	}

	transport := &http.Transport{
		Proxy: proxyFunc,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Ignoring SSL errors as specified
		},
		DisableKeepAlives:     true,
		MaxIdleConnsPerHost:   -1,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}

	return client
}
