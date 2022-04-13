package main

import (
	"fmt"
	"log"
	"time"
)

func checkMaterails() {
	for {
		fmt.Println(res.Materials)
		time.Sleep(time.Second)
	}
}
func updateDB() {
	prep, err := p.NewQuery(updateItemAmount)
	if err != nil {
		log.Printf("can't Run updateDB, %v\n", err)
		return
	}
	for ; ; time.Sleep(waitToUpdateDB) {
		for k, v := range res.Materials {
			err = p.ExecCmd(prep, res.Materials[k], v)
			if err != nil {
				log.Printf("[Error in executing query]: %v", err)
				return
			}
			fmt.Printf("+%v: %v\n", v, k)
			res.Materials[k] = 0
		}
	}
}

// query database to get necessary info
func GetInfo() ([]map[string]any, error) {
	id, err := p.NewQuery(unitsInfo)
	defer p.CloseQuery(id)

	if err != nil {
		return nil, err
	}
	data, err := p.ExecQuery(id)
	if err != nil {
		return nil, err
	}
	return data, nil
}