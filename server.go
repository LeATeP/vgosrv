package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"psql"
	"server"
	"time"
)

var srv *server.Server
var p psql.PsqlInterface

var (
	res          server.Resources // type of data received from client
	waitToUpdateDB = time.Second * 10000
)

const (
	selectTable      = `select * from items order by id;`
	checkIfUnitFree  = `select * from unit_info where container_id is null;`
	updateItemAmount = `update items set amount = amount + $1 where id = $2;`
	unitsInfo        = `select * from unit_info where container_id is null;`
)

func main() {
	var err error
	res.Materials = map[string]int64{}
	srv, err = server.NewServer()
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("server started")

	p, err = psql.PsqlConnect()
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("psql conn ready")

	go updateDB(); go checkMaterails()
	for i := int64(0); ; i++ {
		AcceptConn(i)
	}
}


func AcceptConn(i int64) {
	conn, err := srv.Listener.Accept() // listen for clients
	if err != nil {
		log.Printf("[failed to connect]: %v\n", err)
	}
	fmt.Printf("connected [%v]: %v\n", i, conn)

	srv.ClientConn[i] = &server.Client{
		Conn:    conn,
		Receive: gob.NewDecoder(conn),
		Send:    gob.NewEncoder(conn),
	}
	if !SendInfoToClient(i) {
		disconnectClient(i)
		return
	}
	go ManageConnection(i)
}

// check if unit is available in table `unit_info`
func CheckIfUnitAvailable() int64 {
	id, err := p.NewQuery(checkIfUnitFree)
	defer p.CloseQuery(id)
	if err != nil {
		return -1
	}
	data, err := p.ExecQuery(id)
	if err != nil {
		return -1
	}
	if len(data) == 0 {
		return -1
	}
	return data[0]["unit_id"].(int64)
}

func SendInfoToClient(i int64) bool {
	unitId := CheckIfUnitAvailable()

	if unitId == -1 {
		log.Printf("[Error in getting info from db(%v)]", "SendInfoToClient")
		return false
	}
	client := srv.ClientConn[i]
	client.FromServer = server.FromServer{
		ClientId:  i,
		UnitId:    unitId,
		TickSpeed: time.Second,
		Running:   true,
	}
	client.Send.Encode(&server.Message{MsgCode: 2, FromServer: client.FromServer})
	return true
}

// send necessary info to client about server

func ManageConnection(i int64) {
	var msg server.Message
	client := srv.ClientConn[i]
	defer disconnectClient(i)
	for {
		msg = server.Message{}
		if err := client.Receive.Decode(&msg); err != nil {
			log.Printf("%v [err]: %v\n", i, err) // well would be to put client identifiers like containerId and stuff
			return
		}
		switch msg.MsgCode {
		case 1: // get ping that client is active
		case 2: // get info about client
		case 3: // something changed in client
		case 4: // client shutting down
			fmt.Println(msg)
			fmt.Println("client shutting down ", i)
			return
		case 5: // client reloading
		case 6: // update resources
			for i, k := range msg.Resources.Materials {
				res.Materials[i] += k
			}
		default:
			fmt.Println("0, something wrong")
		}
	}
}
func disconnectClient(i int64) {
	log.Printf("disconnecting %v", i)
	client := srv.ClientConn[i]
	client.Conn.Close()
	delete(srv.ClientConn, i)
}
