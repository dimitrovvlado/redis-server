package main

import (
	"flag"
	"log"

	"github.com/dimitrovvlado/redis-server/internal/server"
)

func main() {
	log.SetFlags(0)

	host := flag.String("host", "localhost", "Server hostname")
	port := flag.Int("port", 6379, "Server port")
	flag.Parse()

	err := server.Serve(*host, *port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err.Error())
	}
}
