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

type queryPool struct {
	id    int64
	name  string
	query string
}

var queryMap = map[int64]queryPool{
	1: {name: `selectItems`, query: `SELECT * FROM items order by id;`},
	2: {name: `selectUnitInfo`, query: `SELECT * FROM unit_info WHERE id = $1;`},
	3: {name: `checkUnitFree`, query: `SELECT * FROM unit_info WHERE container_id IS NULL;`},
	4: {name: `updateItemAmount`, query: `UPDATE items SET amount = amount + $1 WHERE name = $2;`},
	5: {name: `setUnitContainer`, query: `UPDATE unit_info SET container_id = $1 WHERE id = $2;`},
	6: {name: `restoreUnit`, query: `UPDATE unit_info SET container_id = NULL WHERE id = $1;`},
}

var (
	selectItemsId      int64
	selectUnitInfoId   int64
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
	unitId, err := CheckIfUnitAvailable()
	if err != nil {
		log.Printf("[Error in CheckIfUnitAvailable]: %v", err)
		return false
	}

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
	err = client.Send.Encode(&server.Message{MsgCode: 2, FromServer: client.FromServer})
	if err != nil {
		log.Printf("[Error in sending info to client]: %v", err)
		return false
	}
	return true
}

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
			fmt.Println(msg)
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
