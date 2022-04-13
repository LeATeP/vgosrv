package server

import (
	"net"
	"os"
	"time"
)

func NewServer() (*Server, error) {
	ls, err := net.Listen("tcp", listenToLocal)
	if err != nil {
		return nil, err
	}
	return &Server{
		Start: time.Now().UTC(),
		Info: Info{
			SrvId:       "0",
			SrvName:     "main",
			ContainerId: os.Getenv("HOSTNAME"),
			MaxLoad:     100,
		},
		Running:    true,
		Stats:      SrvStats{},
		ClientConn: map[int64]*Client{},
		Listener:   ls,
	}, nil
}
