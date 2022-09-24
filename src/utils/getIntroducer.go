package utils

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
)

type Listener int

var portTCP = 9981

func IsIntroducer(indicator string) bool {
	// check the introducer indicator file, if it is located on this machine
	// then this process would be regarded as the introducer, otherwise it would
	// be a common node
	if _, err := os.Stat(indicator); err == nil {
		fmt.Println("----------------I am a noble introducer ^_^----------------")
		return true
	}
	fmt.Println("----------------I am a pariah node :(----------------")
	return false
}

func (l *Listener) GetLine(line []byte, ack *bool) error {
	fmt.Println(string(line))
	return nil
}

func (l *Listener) JoinRequest() error {
	// a new node is joining the ring
}

func (l *Listener) LeaveRequest() error {
	// an existing node is leaving the ring
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

func StartIntroducer(indicator string) {
	// in main function, if the introducer indicator file is not found, this function will be skipped
	// which means it is a common node, otherwise it will start a goroutine for tcp listener.
	if !IsIntroducer(indicator) {
		return
	}
	go IntroducerWorker()
}
