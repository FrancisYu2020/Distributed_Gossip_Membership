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
	fmt.Println(*buffer)
	return nil
}

func (l *Listener) JoinNotification(msg string, buffer *[]byte) error {
	// The introducer getting the success message from the node and print the join message
	log.Println(msg)
	return nil
}

func (l *Listener) UpdateMemList(msg string, buffer *[]byte) error {
	// the introducer will be asked by the new node to let other nodes
	// know the existence of the new node and update the monitor list accordingly
	//TODO: finish this function
	if msg == "introducer" {
		for _, m := range memList.Members[1:] {
			client, err := rpc.Dial("tcp", m.IP+":"+strconv.Itoa(portTCP))
			if err != nil {
				log.Fatal(err)
			}
			*buffer = nil
			operaChan <- "READ"
			curMem := <-listChan
			if jsonData, err := json.Marshal(curMem); err != nil {
				return err
			} else {
				*buffer = append(*buffer, jsonData...)
			}
			fmt.Println(string(*buffer), "----------------------------------------")
			if err := client.Call("Listener.UpdateMemList", "node", &buffer); err != nil {
				fmt.Println("Hello")
				log.Fatal("Introducer: Error in updating membership lists: ", err)
				return err
			}
		}
	} else {
		log.Println("Hello, I am currently in node ", utils.GetLocalIP())
		if err := json.Unmarshal(*buffer, &memList); err != nil {
			log.Fatal("Node: Error in updating membership list: ", err)
		}
	}

	return nil
}

func (l *Listener) UpdateMonList(msg string, buffer *[]byte) error {
	// the introducer will be asked by the new node to let other nodes
	// know the existence of the new node and update the membership list accordingly
	// TODO: finish this function
	if msg == "introducer" {
		for _, m := range memList.Members[1:] {
			client, err := rpc.Dial("tcp", m.IP+":"+strconv.Itoa(portTCP))
			if err != nil {
				log.Fatal(err)
			}
			if err := client.Call("Listener.UpdateMonList", "node", &buffer); err != nil {
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
			for i := 1; i < 4; i++ {
				monList.Members = append(monList.Members, curMem.Members[(idx+i)%len(curMem.Members)])
			}
		}
	}
	return nil
}

func (l *Listener) HandleLeaveRequest() error {
	// an existing node is leaving the ring
	// TODO: finish this function
	return nil
}

func Leave(target string) {
	// TODO: finish this function
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

func StartTCPServer() {
	// function for starting TCP server, although same as IntroducerWorker
	go IntroducerWorker()
}
