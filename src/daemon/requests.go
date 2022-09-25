package main

import (
	"encoding/json"
	"log"
	"net/rpc"
	"src/utils"
)

func NodeJoin() (string, error) {
	client, err := rpc.Dial("tcp", "fa22-cs425-2210.cs.illinois.edu:9981")
	if err != nil {
		log.Fatal(err)
	}
	id := JoinRequest(client)
	RetrieveInfo(client)
	// next we show join success message on both introducer and current node
	msg := utils.GetLocalIP() + " successfully joined the ring!"
	if len(memList.Members) > 1 {
		log.Println(msg)
	}
	var buffer []byte
	if err := client.Call("Listener.JoinNotification", msg, &buffer); err != nil {
		log.Fatal("Error in notifying the introducer: ", err)
	}

	// ask the introducer to let other nodes update the membership and monitor lists
	UpdateRequest(client)
	return id, nil
}

func JoinRequest(client *rpc.Client) string {
	// current node sending request to the introducer to join the ring
	ip := utils.GetLocalIP()
	id := utils.GenerateID(ip)
	msg := ip + "\r\n\r\n" + id
	var reply bool
	if err := client.Call("Listener.HandleJoinRequest", []byte(msg), &reply); err != nil {
		log.Fatal("Error in join request: ", err)
	}
	return id
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
}

func UpdateRequest(client *rpc.Client) {
	// The new node use this function to ask the introducer to send tcp messages to other members
	// so that the other members could update their membership and monitor lists accordingly.
	var buffer []byte
	var reply []byte
	reply = []byte("introducer")
	buffer = []byte("introducer")
	if err := client.Call("Listener.UpdateMemList", buffer, &reply); err != nil {
		log.Fatal("Error in other members updating memList: ", err)
	}
	if err := client.Call("Listener.UpdateMonList", buffer, &reply); err != nil {
		log.Fatal("Error in other members updating monList: ", err)
	}
}

//// This part is for the node to leave the ring functionality
//// Similar to the join
func NodeLeave() error {
	client, err := rpc.Dial("tcp", "fa22-cs425-2210.cs.illinois.edu:9981")
	if err != nil {
		log.Fatal(err)
	}
	LeaveRequest(client)
	ClearLocalCache(client)
	// next we show join success message on both introducer and current node
	msg := utils.GetLocalIP() + " successfully left the ring!"
	if len(memList.Members) > 1 {
		log.Println(msg)
	}
	var buffer []byte
	if err := client.Call("Listener.LeftNotification", msg, &buffer); err != nil {
		log.Fatal("Error in notifying the introducer: ", err)
	}

	// ask the introducer to let other nodes update the membership and monitor lists
	UpdateLeaveRequest(client)
	return nil
}

func LeaveRequest(client *rpc.Client) {
	// current node sending request to the introducer to join the ring
	ip := utils.GetLocalIP()
	var reply []byte
	if err := client.Call("Listener.HandleLeaveRequest", ip, &reply); err != nil {
		log.Fatal("Error in leave request: ", err)
	}
}

func ClearLocalCache(client *rpc.Client) {
	// clear the membership list and monitor list in the daemon
	memList.Members = memList.Members[:0]
	monList.Members = monList.Members[:0]
}

func UpdateLeaveRequest(client *rpc.Client) {
	// The new node use this function to ask the introducer to send tcp messages to other members
	// so that the other members could update their membership and monitor lists accordingly.
	var buffer []byte
	var reply []byte
	reply = []byte("introducer")
	buffer = []byte("introducer")
	if err := client.Call("Listener.UpdateMemListLeave", utils.GetLocalIP(), &reply); err != nil {
		log.Fatal("Error in other members updating memList: ", err)
	}
	if err := client.Call("Listener.UpdateMonList", buffer, &reply); err != nil {
		log.Fatal("Error in other members updating monList: ", err)
	}
}
