package server

import (
	"encoding/gob"
	"net"
	"time"
)

const (
	TickSpeed        = time.Second
	listenToLocal    = "localhost:9000"
	connectToLocal   = "localhost:9000"
	connectToNetwork = "postgres:9000"
)

// Server Specific Types "Below"
// type Info about the server that passes to client
type Info struct {
	SrvId       string
	SrvName     string
	ContainerId string
	MaxLoad     int64
}
type Server struct {
	Start      time.Time
	Info       Info
	Stats      SrvStats
	Running    bool
	Listener   net.Listener
	ClientConn map[int64]*Client // info the server have about client
}
type SrvStats struct {
	InMsgs   int64
	OutMsgs  int64
	InBytes  int64
	OutBytes int64
}
type Client struct {
	Conn       net.Conn
	Send       *gob.Encoder
	Receive    *gob.Decoder
	FromServer FromServer
	FromClient FromClient
}
type FromClient struct {
	StartTime   time.Time // when client started working
	ContainerId string    // hostname
	Running     bool      // `false` the client is stopping working
}
type FromServer struct {
	ClientId  int64         // assigned Client Id by server
	UnitId    int64         // free unit assigned to client
	TickSpeed time.Duration // time between server ticks
	Running   bool          //  singal `false` to stop working for client
}
type Message struct {
	MsgCode    int64      // msg code system, to understand what msg come up is
	Resources  Resources  // output of the client work
	FromServer FromServer // info about server
	FromClient FromClient // info about client
}
type Resources struct {
	Materials map[string]int64
}

// ---------------------------------------------
// Client Specific types
