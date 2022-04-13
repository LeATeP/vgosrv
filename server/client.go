package server

import (
	"encoding/gob"
	"log"
	"net"
	"os"
	"time"
)

func NewClient() (*Client, error) {
	var conn net.Conn
	var err error
	for i:=0; i<10; i++ { // 5 second to try to connect to a server
		conn, err = net.Dial("tcp", connectToLocal) // connect to a server
		if err != nil {
			log.Printf("[failed to connect after %v sec]: %v\n", float32(i)/2, err)
			time.Sleep(time.Second / 2)
			continue
		}
		break
	}
	if conn == nil {
		return nil, err
	}
	return &Client{
		Conn:    conn,
		Receive: gob.NewDecoder(conn),
		Send:    gob.NewEncoder(conn),
		FromClient: FromClient{
			StartTime:   time.Now().UTC(),
			ContainerId: os.Getenv("HOSTNAME"),
			Running:     true,
		},
	}, nil
}
