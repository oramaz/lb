package pool

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"sync"
	"time"
)

// conn describes a service connection
type conn struct {
	Url   *url.URL
	Port  string
	Proxy *httputil.ReverseProxy
	alive bool
	load  uint
	mu    *sync.RWMutex
}

// Set connection alive status
func (c *conn) setAlive(alive bool) {
	c.mu.Lock()
	c.alive = alive
	c.mu.Unlock()
}

// Get connection alive status
func (c *conn) getAlive() (alive bool) {
	c.mu.RLock()
	alive = c.alive
	c.mu.RUnlock()

	return
}

// Set connection load
func (c *conn) setLoad(load uint) {
	c.mu.Lock()
	c.load = load
	c.mu.Unlock()
}

// Get connection load
func (c *conn) getLoad() (load uint) {
	c.mu.RLock()
	load = c.load
	c.mu.RUnlock()

	return
}

// Handle service's request
func (c *conn) handleRequest(w http.ResponseWriter, r *http.Request) {
	c.setLoad(c.getLoad() + 1)

	// Generate a random time from 2s to 5s
	min := 2
	max := 5
	t := rand.Intn(max-min) + min

	// Simulate request handling
	time.Sleep(time.Second * time.Duration(t))

	c.setLoad(c.getLoad() - 1)
}

// ConnPool describes a pool of services
type ConnPool struct {
	Conns []*conn
}

// Create a pool of launched services
func Create(hosts []string) (*ConnPool, error) {
	// Check if at least 1 host is passed
	if len(hosts) <= 0 {
		return nil, errors.New("at least 1 host should be passed")
	}

	pool := &ConnPool{}

	// Fill in the pool
	pool.Conns = make([]*conn, len(hosts))
	for i := range pool.Conns {
		h := hosts[i]

		url, err := url.Parse(h)
		if err != nil {
			return nil, err
		}

		pool.Conns[i] = &conn{
			Url:   url,
			Port:  ":" + url.Port(),
			Proxy: getReverseProxy(pool, url),
			alive: true,
			load:  0,
			mu:    new(sync.RWMutex),
		}
	}

	// Launch services
	for i, c := range pool.Conns {
		s := &http.Server{
			Addr:    c.Port,
			Handler: http.HandlerFunc(c.handleRequest),
		}

		go func() {
			if err := s.ListenAndServe(); err != nil {
				return
			}
		}()
		log.Printf("[%s] Service %d started\n", c.Url.Host, i+1)

		// Specific actions for the 1st service
		if i == 0 {
			u := c.Url

			// Shutdown after 100s of running
			time.AfterFunc(time.Second*100, func() {
				if err := s.Shutdown(context.Background()); err != nil {
					log.Fatal(err)
				}
				log.Printf("[%s] Shutdown\n", u.Host)
			})

			s1 := &http.Server{
				Addr:    c.Port,
				Handler: http.HandlerFunc(c.handleRequest),
			}
			// Launch it again on the 200s
			time.AfterFunc(time.Second*200, func() {
				go func() {
					if err := s1.ListenAndServe(); err != nil {
						log.Println(err)
						return
					}
				}()
				log.Printf("[%s] Listen\n", u.Host)
			})
		}
	}

	return pool, nil
}

// Next calculates and returns the connection to transmit the request.
// Uses a Least Connections algorithm.
func (cp *ConnPool) Next() *conn {
	conns := cp.Conns

	sort.SliceStable(conns, func(i, j int) bool {
		return conns[i].load < conns[j].load
	})

	for i := range conns {
		if conns[i].getAlive() {
			return conns[i]
		}
	}

	return nil
}

// HealthCheck is a passive services' health checking function
// running in a background
func (cp *ConnPool) HealthCheck() {
	for _, c := range cp.Conns {
		status := "ok"

		// Ping connection
		alive := pingConn(c.Url)
		c.setAlive(alive)
		if !alive {
			status = "error"
		}

		log.Printf("[%s] Status: %s\n", c.Url.Host, status)
	}
}

// LoadStatistics prints the load of each connection in the pool
func (cp *ConnPool) LoadStatistics() {
	for _, c := range cp.Conns {
		log.Printf("[%s] Load: %d", c.Url.Host, c.getLoad())
	}
}

// GetConn returns the connection with passed URL from the pool
func (cp *ConnPool) GetConn(url *url.URL) *conn {
	for _, c := range cp.Conns {
		if c.Url.String() == url.String() {
			return c
		}
	}

	return nil
}
