package main

import (
	"fmt"
	"log"
	"time"
)
// ready query's with newQuery
func prepareQuery() {
	var err error
	selectTableId, err = p.NewQuery(selectTable)
	if err != nil {
		log.Fatal(err)
	}
	checkIfUnitFreeId, err = p.NewQuery(checkIfUnitFree)
	if err != nil {
		log.Fatal(err)
	}
	updateItemAmountId, err = p.NewQuery(updateItemAmount)
	if err != nil {
		log.Fatal(err)
	}
	setUnitContainerId, err = p.NewQuery(setUnitContainer)
	if err != nil {
		log.Fatal(err)
	}
	restoreUnitId, err = p.NewQuery(restoreUnit)
	if err != nil {
		log.Fatal(err)
	}
}
// prepare database for new clients, reset all clients imprint from database
func prepareDatabase() {
	p.ExecCmdFast(resetAllUnits)
}

// unit with containerId is nill, in table `unit_info` for client allocation
func CheckIfUnitAvailable() int64 {
	data, err := p.ExecQuery(checkIfUnitFreeId)
	if err != nil {
		return -1
	}
	if len(data) == 0 {
		return -1
	}
	return data[0]["unit_id"].(int64)
}

func checkMaterails() {
	for {
		mutex.Lock()
		fmt.Println(res.Materials)
		mutex.Unlock()
		time.Sleep(time.Second)
	}
}
func updateDB() {
	for ; ; time.Sleep(updateDatabasePer) {
		mutex.Lock()
		for k, v := range res.Materials {
			err := p.ExecCmd(updateItemAmountId, v, k)
			if err != nil {
				log.Printf("[Error in update amount in db query]: %v", err)
				return
			}
			fmt.Printf("+%v: %v\n", v, k)
			res.Materials[k] = 0
		}
		mutex.Unlock()
	}
}
func disconnectClient(i int64) {
	log.Printf("disconnecting %v", i)
	client := srv.ClientConn[i]
	err := p.ExecCmd(restoreUnitId, client.FromServer.UnitId)
	if err != nil {
		log.Printf("[Error in restoring unit id in database]: %v", err)
	}
	client.Conn.Close()
	delete(srv.ClientConn, i)
}

