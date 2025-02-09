package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/dimitrovvlado/redis-server/internal/protocol"
)

func main() {
	log.SetFlags(0)

	host := flag.String("host", "localhost", "Server hostname, defaults to 'localhost'")
	port := flag.Int("port", 6379, "Server port, defaults to 6379")
	flag.Parse()
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatalf("Could not connect to Redis at %s:%d: %v", *host, *port, err.Error())
	}
	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 4096)
	rbuf := make([]byte, 1024)
	for {
		fmt.Printf("%s:%d> ", *host, *port)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		if line == "exit" || line == "quit" {
			return
		}

		encoded := protocol.Encode(encodeCmd(line))

		_, err := conn.Write([]byte(encoded))
		if err != nil {
			log.Fatalf("Write error: %v", err.Error())
			return
		}

		for {
			n, err := conn.Read(rbuf)
			if err != nil {
				log.Fatalf("Read error: %v", err.Error())
				return
			}
			if n > 0 {
				buf = append(buf, rbuf[:n]...)
				frame, size := protocol.ExtractFrameFromBuffer(buf)
				if frame != nil {
					fmt.Printf("%s\n", frame.String())
					buf = buf[size:] //trim the extracted frame
					break
				}
			}
		}
	}
}

func encodeCmd(cmd string) protocol.Resp {
	fields := strings.Fields(cmd)
	var resp protocol.Array
	for _, f := range fields {
		resp.Items = append(resp.Items, protocol.BulkString{Data: protocol.Ptr(f)})
	}
	return resp
}
