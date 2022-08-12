package main

import (
	"log"

	"github.com/oramaz/lb/internal/pool"
	"github.com/oramaz/lb/internal/server"
	"github.com/oramaz/lb/internal/util"
)

func main() {
	// Get configs from config.json file
	c := util.New()

	// Create pool of launched services
	pool, err := pool.Create(c.Hosts)
	if err != nil {
		log.Fatal("error while spawning the servers: ", err)
	}

	// Run load-balancer server
	s := server.New(c.Port, pool)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
