package gptbest

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

func newHTTPClient() *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.Proxy = http.ProxyFromEnvironment
	tr.TLSHandshakeTimeout = 30 * time.Second
	tr.ResponseHeaderTimeout = 90 * time.Second
	tr.ExpectContinueTimeout = 1 * time.Second
	tr.IdleConnTimeout = 90 * time.Second
	return &http.Client{
		Transport: tr,
		Timeout:   10 * time.Minute,
	}
}

func (p *Provider) doWithRetry(req *http.Request, maxAttempts int) (*http.Response, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var lastErr error
	for i := 1; i <= maxAttempts; i++ {
		r := req.Clone(req.Context())
		if req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			r.Body = body
		}
		resp, err := p.httpClient.Do(r)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if i == maxAttempts || !shouldRetryNetErr(err) {
			break
		}
		time.Sleep(time.Duration(i) * 600 * time.Millisecond)
	}
	return nil, lastErr
}

func shouldRetryNetErr(err error) bool {
	if err == nil {
		return false
	}
	if isTLSHandshakeTimeout(err) {
		return true
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection reset") || strings.Contains(msg, "unexpected eof")
}

func isTLSHandshakeTimeout(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "tls handshake timeout")
}
