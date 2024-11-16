package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func CreateHTTPClient(config Config) *http.Client {
	proxyURL, _ := url.Parse(fmt.Sprintf(
		"http://%s:%s@%s:%s",
		config.ProxyUser, config.ProxyPassword, config.ProxyHost, config.ProxyPort,
	))

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
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
