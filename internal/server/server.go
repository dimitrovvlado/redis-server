package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/dimitrovvlado/redis-server/internal/commands"
	"github.com/dimitrovvlado/redis-server/internal/datastore"
	"github.com/dimitrovvlado/redis-server/internal/protocol"
)

func Serve(host string, port int, ds *datastore.Datastore) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Failed to establish a connection with the client: %v", err.Error())
		}
		defer conn.Close()
		go handleConnection(conn, ds)
	}
}

func handleConnection(conn net.Conn, ds *datastore.Datastore) {
	buf := make([]byte, 0, 4096)
	rbuf := make([]byte, 1024)
	defer conn.Close()
	for {
		n, err := conn.Read(rbuf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				addr := conn.RemoteAddr()
				if addr != nil {
					log.Printf("Connection closed by client: %s", addr.String())
				} else {
					log.Printf("Connection closed by client.")
				}
				return
			}
		}
		if n > 0 {
			buf = append(buf, rbuf[:n]...)
			frame, size := protocol.ExtractFrameFromBuffer(buf)
			if frame != nil {
				result, err := commands.HandleCommand(frame, ds)
				if err != nil {
					log.Println("Error handling command: ", err)
				} else {
					_, err := conn.Write(protocol.Encode(result))
					if err != nil {
						log.Println("Error writing to connection: ", err)
					}
				}
				//trim to remove frame
				buf = buf[size:]
			}
		}
	}
}
