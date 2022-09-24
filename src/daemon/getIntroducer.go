package main

import (
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

func (l *Listener) GetLine(line []byte, ack *bool) error {
	fmt.Println(string(line))
	return nil
}

func (l *Listener) HandleJoinRequest(msg []byte, ack *bool) error {
	// a new node is joining the ring
	buffer := strings.Split(string(msg), "\r\n\r\n")
	ip := buffer[0]
	id := buffer[1]
	fmt.Println("Node", ip, "is joining...")
	fmt.Println(ip, "----", id, len(ip), len(id))
	fmt.Println(string(msg), "  join request received!")
	// log.Println("Join request received!!!----------------------------")
	return nil
}

func (l *Listener) HandleLeaveRequest() error {
	// an existing node is leaving the ring
	return nil
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
