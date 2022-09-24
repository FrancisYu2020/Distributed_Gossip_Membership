package main

import (
	"fmt"
	"log"
	"net/rpc"
	"src/utils"
	"strconv"
)

func JoinRequest(portTCP int) error {
	client, err := rpc.Dial("tcp", "fa22-cs425-2210.cs.illinois.edu:"+strconv.Itoa(portTCP))
	fmt.Println(client)
	if err != nil {
		log.Fatal(err)
	}
	ip := utils.GetLocalIP()
	id := utils.GenerateID(ip)
	msg := ip + "\r\n\r\n" + id
	var reply bool
	err = client.Call("Listener.HandleJoinRequest", []byte(msg), &reply)
	if err != nil {
		log.Fatal("Error in join request: ", err)
	}
	return nil
}

func LeaveRequest() error {
	return nil
}
