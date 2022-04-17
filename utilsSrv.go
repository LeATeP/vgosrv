package main

import (
	"fmt"
	"log"
	"time"
)

// ready query's with newQuery
func prepareQuery() {
	for i, k := range queryMap {
		var err error
		k.id, err = p.NewQuery(k.query)
		if err != nil {
			log.Fatal(err)
			return
		}
		queryMap[i] = k
	}
	selectItemsId 	   = queryMap[1].id
	selectUnitInfoId   = queryMap[2].id
	checkIfUnitFreeId  = queryMap[3].id
	updateItemAmountId = queryMap[4].id
	setUnitContainerId = queryMap[5].id
	restoreUnitId 	   = queryMap[6].id

}

// prepare database for new clients, reset all clients imprint from database
func prepareDatabase() {
	p.ExecCmdFast(resetAllUnits)
}

// unit with containerId is nill, in table `unit_info` for client allocation
func CheckIfUnitAvailable() (int64, error) {
	data, err := p.ExecQuery(checkIfUnitFreeId)
	if err != nil {
		return -1, err
	}
	if len(data) == 0 {
		return -1, fmt.Errorf("no free units")
	}
	return data[0]["id"].(int64), nil
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
