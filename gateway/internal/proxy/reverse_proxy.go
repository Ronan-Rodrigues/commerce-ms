package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewReverseProxy cria um proxy reverso para um serviço interno
func NewReverseProxy(targetURL string) (http.Handler, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
		if clientIP := req.Header.Get("X-Real-IP"); clientIP == "" {
			req.Header.Set("X-Real-IP", req.RemoteAddr)
		}
	}

	return proxy, nil
}
