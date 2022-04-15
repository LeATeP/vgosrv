package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"psql"   // https://github.com/LeATeP/Vaava_go/psql
	"server" // https://github.com/LeATeP/Vaava_go/server
	"sync"
	"time"
)

var srv *server.Server
var p psql.PsqlInterface

var (
	res   server.Resources // type of data received from client
	mutex = sync.RWMutex{}

	updateDatabasePer = time.Second * 6
)

const ( // exec at the start of the server
	resetAllUnits = `UPDATE unit_info SET container_id = NULL;`
)

const (
	selectTable      = `select * from items order by id;`
	checkIfUnitFree  = `select * from unit_info where container_id is null order by unit_id;`
	updateItemAmount = `update items set amount = amount + $1 where name = $2;`
	setUnitContainer = `update unit_info set container_id = $1 where unit_id = $2;`
	restoreUnit      = `update unit_info set container_id = null where unit_id = $1;`
)

var (
	selectTableId      int64
	checkIfUnitFreeId  int64
	updateItemAmountId int64
	setUnitContainerId int64
	restoreUnitId      int64
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
	prepareDatabase()
	prepareQuery()

	go updateDB()
	go checkMaterails()
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

	mutex.Lock()
	srv.ClientConn[i] = &server.Client{
		Conn:    conn,
		Receive: gob.NewDecoder(conn),
		Send:    gob.NewEncoder(conn),
	}
	mutex.Unlock()
	if !SendInfoToClient(i) {
		disconnectClient(i)
		return
	}
	go ManageConnection(i)
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
	err := client.Send.Encode(&server.Message{MsgCode: 2, FromServer: client.FromServer})
	if err != nil {
		log.Printf("[Error in sending info to client]: %v", err)
		return false
	}
	return true
}

// send necessary info to client about server

func ManageConnection(i int64) {
	var msg server.Message
	mutex.Lock()
	client := srv.ClientConn[i]
	mutex.Unlock()
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
			mutex.Lock()
			client.FromClient = msg.FromClient
			container := msg.FromClient.ContainerId
			id := client.FromServer.UnitId
			mutex.Unlock()

			err := p.ExecCmd(setUnitContainerId, container, id)
			if err != nil {
				log.Printf("[Error in setting container id into db]: %v", err)
				return
			}
		case 3: // something changed in client
		case 4: // client shutting down
			fmt.Println("client shutting down ", i)
			return
		case 5: // client reloading
		case 6: // update resources
			mutex.Lock()
			for i, k := range msg.Resources.Materials {
				res.Materials[i] += k
			}
			mutex.Unlock()
		default:
			fmt.Println("0, something wrong")
		}
	}
}
