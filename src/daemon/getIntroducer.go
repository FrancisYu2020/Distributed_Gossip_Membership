package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
)

type Listener int

func IsIntroducer(indicator string) bool {
	// check the introducer indicator file, if it is located on this machine
	// then this process would be regarded as the introducer, otherwise it would
	// be a common node
	// return true if it is an introducer, otherwise false
	if _, err := os.Stat(indicator); err == nil {
		return true
	}
	return false
}

func MemberIdentical(m1 Member, m2 Member) {
	//
}

func (l *Listener) GetLine(line []byte, ack *bool) error {
	fmt.Println(string(line))
	return nil
}

func (l *Listener) HandleJoinRequest(msg []byte, ack *bool) error {
	// a new node is joining the ring
	buffer := strings.Split(string(msg), "\r\n\r\n")
	m := Member{buffer[0], buffer[1]}

	// prevent node to join if it is already in the ring
	operaChan <- "READ"
	curMem := <-listChan
	for _, m0 := range curMem.Members {
		if m0.ip == m.ip {
			log.Println("Received join request from Member who are already in the ring!")
			return nil
		}
	}
	bufferChan <- m
	operaChan <- "ADD"
	return nil
}

func (l *Listener) HandleRetrieveInfo(msg string, buffer *[]byte) error {
	// sending the Membership list and monitor list to the new node
	operaChan <- "READ"
	curMem := <-listChan
	if jsonData, err := json.Marshal(curMem); err != nil {
		return err
	} else {
		*buffer = append(*buffer, jsonData...)
	}
	return nil
}

func (l *Listener) HandleLeaveRequest() error {
	// an existing node is leaving the ring
	return nil
}

func Leave(target string) {
	// kill cur monitors
	// if strings.Compare(target, localID) == 0 {
	// 	// do not delete self
	// 	return
	// }
	// var idx int = -1
	// for i, m := range memList.Members {
	// 	if strings.Compare(m.id, target) == 0 {
	// 		idx = i
	// 		break
	// 	}
	// }
	// if idx == -1 {
	// 	// do not need to send message now
	// 	// fmt.Println("After Members:", memList.Members)
	// 	// fmt.Println("------------------")
	// 	return
	// }
	// memList.Members = append(memList.Members[:idx], memList.Members[idx+1:]...)

	// idx = -1
	// for i, m := range memList.Members {
	// 	if strings.Compare(m.id, localID) == 0 {
	// 		idx = i
	// 		break
	// 	}
	// }
	// monList.Members = []Member{}
	// // if we do not have at least 3 other Members
	// if len(memList.Members) <= 3 {
	// 	monList.Members = append(monList.Members, memList.Members[:idx]...)
	// 	monList.Members = append(monList.Members, memList.Members[idx+1:]...)
	// } else {
	// 	var newList []Member
	// 	for i := 1; i <= 2; i++ {
	// 		newList = append(newList, memList.Members[(idx+i)%len(memList.Members)])
	// 	}
	// 	newList = append(newList, memList.Members[(idx-1)%len(memList.Members)])
	// 	monList.Members = newList
	// }
	// // fmt.Println("After Members:", memList.Members)
	// // fmt.Println("------------------")
	// return
}

func IntroducerWorker() {
	// the introducer start the tcp and listen locally
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(portTCP))
	fmt.Println("running at port: ", portTCP)
	if err != nil {
		log.Fatal("Error in resolving local address: ", err)
	}
	inbound, err := net.ListenTCP("tcp", (*net.TCPAddr)(tcpAddr))
	if err != nil {
		log.Fatal("Error in introducer listening port: ", err)
	}

	listener := new(Listener)
	rpc.Register(listener)
	rpc.Accept(inbound)
}

func StartIntroducer() {
	// in main function, if the introducer indicator file is not found, this function will be skipped
	// which means it is a common node, otherwise it will start a goroutine for tcp listener.
	go IntroducerWorker()
}
