package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/rpc"
	"src/utils"
)

func NodeJoin() error {
	client, err := rpc.Dial("tcp", "fa22-cs425-2210.cs.illinois.edu:9981")
	if err != nil {
		log.Fatal(err)
	}
	JoinRequest(client)
	RetrieveInfo(client)
	// next we show join success message on both introducer and current node
	msg := utils.GetLocalIP() + " successfully joined the ring!"
	if len(memList.Members) > 1 {
		log.Println(msg)
	}
	var buffer []byte
	log.Println(memList, "---------------first join node (introducer)")
	if jsonData, err := json.Marshal(memList); err != nil {
		return err
	} else {
		buffer = append(buffer, jsonData...)
	}
	if err := client.Call("Listener.JoinNotification", msg, &buffer); err != nil {
		log.Fatal("Error in notifying the introducer: ", err)
	}

	// ask the introducer to let other nodes update the membership and monitor lists
	UpdateRequest(client)
	log.Println("Other nodes updated successfully!")
	return nil
}

func JoinRequest(client *rpc.Client) {
	// current node sending request to the introducer to join the ring
	ip := utils.GetLocalIP()
	id := utils.GenerateID(ip)
	msg := ip + "\r\n\r\n" + id
	var reply bool
	if err := client.Call("Listener.HandleJoinRequest", []byte(msg), &reply); err != nil {
		log.Fatal("Error in join request: ", err)
	}
}

func RetrieveInfo(client *rpc.Client) {
	// current node sending request to the introducer to retrieve the membership list and
	// the monitor list for this node
	var reply []byte
	clientIP := utils.GetLocalIP()
	// "Please send me the membership and monitor lists!"
	if err := client.Call("Listener.HandleRetrieveInfo", clientIP, &reply); err != nil {
		log.Fatal("Error in retrieving membership list and monitor list: ", err)
	}
	// fmt.Println(reply)
	// lists := bytes.Split(reply, []byte("\r\n\r\n"))
	if err := json.Unmarshal(reply, &memList); err != nil {
		log.Fatal("Error in retrieving membership list and monitor list: ", err)
	}

	// From the current membershiplist, infer the monitor list
	// handle the monitor list for the new node
	if len(memList.Members) < 6 {
		for _, m := range memList.Members {
			if m.IP == clientIP {
				continue
			}
			monList.Members = append(monList.Members, m)
		}
	} else {
		// each node will monitor on (curr + i) % len(memList)-th node in the memList
		// to ensure the consistency
		monList.Members = memList.Members[:4]
	}
	fmt.Println(memList.Members, monList.Members)
}

func UpdateRequest(client *rpc.Client) {
	// The new node use this function to ask the introducer to send tcp messages to other members
	// so that the other members could update their membership and monitor lists accordingly.
	var buffer []byte
	if err := client.Call("Listener.UpdateMemList", "Please let other members know me!", &buffer); err != nil {
		log.Fatal("Error in other members updating memList: ", err)
	}
	if err := client.Call("Listener.UpdateMonList", "Please let other members update their monitor lists", &buffer); err != nil {
		log.Fatal("Error in other members updating monList: ", err)
	}
}

func LeaveRequest() error {
	return nil
}
