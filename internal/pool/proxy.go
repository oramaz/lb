package pool

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/oramaz/lb/internal/util"
)

// getReverseProxy returns httputil.ReverseProxy configured object
func getReverseProxy(pool *ConnPool, url *url.URL) (proxy *httputil.ReverseProxy) {
	proxy = httputil.NewSingleHostReverseProxy(url)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		// log.Printf("[%s] Error: %s\n", url.Host, e.Error())

		/*
			Retry to send a service request 3 times
		*/
		retries := util.GetRetryFromContext(r)
		if retries < 3 {
			time.Sleep(time.Millisecond * 10)
			// log.Printf("[%s] Retrying the request...\n", url.Host)
			ctx := context.WithValue(r.Context(), util.Retry, retries+1)
			proxy.ServeHTTP(w, r.WithContext(ctx))

			return
		}

		// If service didn't ask after 3 retries - it's not alive
		conn := pool.GetConn(url)
		conn.setAlive(false)

		/*
			Attempt to send the request to another alive service
		*/
		attempts := util.GetAttemptsFromContext(r)
		// log.Printf("[%s] Attempt %d\n", url.Host, attempts)
		if attempts > len(pool.Conns) {
			log.Printf("[%s] Max attempts reached\n", r.URL.Host)
			http.Error(w, "Service not available", http.StatusServiceUnavailable)
			return
		}

		next := pool.Next()
		if next == nil {
			http.Error(w, "Service not available", http.StatusServiceUnavailable)
			return
		}
		ctx := context.WithValue(r.Context(), util.Attempts, attempts+1)

		pool.Next().Proxy.ServeHTTP(w, r.WithContext(ctx))
	}

	return
}

// pingConn sends a request to ping the passed URL
func pingConn(u *url.URL) bool {
	timeout := 2 * time.Second

	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Printf("[%s] Connection failed: %s\n", u.Host, err)
		return false
	}
	defer conn.Close()

	return true
}
