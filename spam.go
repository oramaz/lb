package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	host := os.Getenv("TARGET_HOST")
	if len(host) == 0 {
		host = "localhost"
	}

	log.Println("Spammer has been started.")

	for {
		time.Sleep(5 * time.Millisecond)
		go func() {
			http.Get(fmt.Sprintf("http://%s:8080", host))
		}()
	}
}
