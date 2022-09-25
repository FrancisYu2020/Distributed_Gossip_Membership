package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"src/utils"
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
		if m0.IP == m.IP {
			log.Println("Warning: Received join request from Member who are already in the ring!")
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
	// fmt.Println(*buffer)
	return nil
}

func (l *Listener) JoinNotification(msg string, buffer *[]byte) error {
	// The introducer getting the success message from the node and print the join message
	log.Println(msg)
	return nil
}

func Max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func (l *Listener) UpdateMemList(buffer []byte, msg *[]byte) error {
	// the introducer will be asked by the new node to let other nodes
	// know the existence of the new node and update the monitor list accordingly
	if string(buffer) == "introducer" {
		for _, m := range memList.Members[:Max(1, len(memList.Members)-1)] {
			// log.Println("Dial address is ", m.IP+":"+strconv.Itoa(portTCP))
			client, err := rpc.Dial("tcp", m.IP+":"+strconv.Itoa(portTCP))
			if err != nil {
				log.Fatal(err)
			}
			var buffer1 []byte
			operaChan <- "READ"
			curMem := <-listChan
			if jsonData, err := json.Marshal(curMem); err != nil {
				return err
			} else {
				buffer1 = append(buffer1, jsonData...)
			}
			var reply []byte
			if err := client.Call("Listener.UpdateMemList", buffer1, &reply); err != nil {
				log.Fatal("Introducer: Error in updating membership lists: ", err)
				return err
			}
		}
	} else {
		// log.Println("Hello, I am currently in node ", utils.GetLocalIP())
		if err := json.Unmarshal(buffer, &memList); err != nil {
			log.Fatal("Node: Error in updating membership list: ", err)
		}
	}
	operaChan <- "RESTART"
	return nil
}

func (l *Listener) UpdateMonList(buffer []byte, msg *[]byte) error {
	// the introducer will be asked by the new node to let other nodes
	// know the existence of the new node and update the membership list accordingly
	if string(buffer) == "introducer" {
		if len(memList.Members) == 0 {
			return nil
		}
		if len(memList.Members) < 6 {
			monList.Members = memList.Members[1:]
		} else {
			monList.Members = memList.Members[1:5]
		}
		// log.Println(memList.Members)
		for _, m := range memList.Members[1:] {
			client, err := rpc.Dial("tcp", m.IP+":"+strconv.Itoa(portTCP))
			if err != nil {
				log.Fatal(err)
			}
			var reply []byte
			if err := client.Call("Listener.UpdateMonList", []byte("node"), &reply); err != nil {
				log.Fatal("Introducer: Error in updating monitor lists: ", err)
				return err
			}
		}
	} else {
		operaChan <- "READ"
		curMem := <-listChan
		idx := -1
		for i, m := range curMem.Members {
			if m.IP == utils.GetLocalIP() {
				idx = i
				break
			}
		}
		if idx == -1 {
			log.Fatal("In UpdateMonList: Something went wrong in the membership lists")
		}
		monList.Members = nil
		if len(curMem.Members) < 6 {
			monList.Members = append(monList.Members, curMem.Members[:idx]...)
			monList.Members = append(monList.Members, curMem.Members[idx+1:]...)
		} else {
			for i := 1; i < 5; i++ {
				monList.Members = append(monList.Members, curMem.Members[(idx+i)%len(curMem.Members)])
			}
		}
	}
	return nil
}

//// functions for handling leave request
////
func (l *Listener) HandleLeaveRequest(msg string, ack *[]byte) error {
	// a new node is joining the ring
	// log.Println("Deleting", msg, "...")
	var idx int = -1
	for i, m := range memList.Members {
		if strings.Compare(m.IP, msg) == 0 {
			idx = i
			break
		}
	}
	// log.Println(idx)
	if idx == -1 {
		// do not need to send message now
		return nil
	}
	memList.Members = append(memList.Members[:idx], memList.Members[idx+1:]...)
	// log.Println(memList.Members)
	return nil
}

func (l *Listener) LeftNotification(msg string, buffer *[]byte) error {
	// The introducer getting the success message from the node and print the join message
	log.Println(msg)
	return nil
}

func (l *Listener) UpdateMemListLeave(ipToLeave string, msg *[]byte) error {
	// the introducer will be asked by the new node to let other nodes
	// know the existence of the new node and update the monitor list accordingly
	for _, m := range memList.Members {
		// log.Println("Dial address is ", m.IP+":"+strconv.Itoa(portTCP))
		client, err := rpc.Dial("tcp", m.IP+":"+strconv.Itoa(portTCP))
		if err != nil {
			log.Fatal(err)
		}
		var reply []byte
		if err := client.Call("Listener.HandleLeaveRequest", ipToLeave, &reply); err != nil {
			log.Fatal("Error in updating membership lists when leaving: ", err)
			return err
		}
	}
	return nil
}

// Goroutine worker design
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

func StartTCPServer() {
	// function for starting TCP server, although same as IntroducerWorker
	go IntroducerWorker()
}
