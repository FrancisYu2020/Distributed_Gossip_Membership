package main

import (
	"fmt"
	"log"
	"mp2-hangy6-tian23/src/utils"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type list struct {
	members []string
	mu      sync.Mutex
}

/*******  Global Variable ******/

// const introducer = "127.0.0.1"

var localIp = utils.GetLocalIP()
var localName = "test"
var port = 9980
var memList list
var monList list

/*******************************/

// delete failed or leaved node from local list
func delete(targetIp string) {
	// idx is the index of deleted peer
	var idx int
	for i, ip := range memList.members {
		if strings.Compare(ip, targetIp) == 0 {
			idx = i
			break
		}
	}
	memList.members = append(memList.members[:idx], memList.members[idx+1:]...)
}

// based on current member list, update the monitor list
func updateMonList() {
	monList.mu.Lock()
	defer monList.mu.Unlock()
	// idx is the index of current node
	var idx int
	for i, ip := range memList.members {
		if strings.Compare(ip, localIp) == 0 {
			idx = i
			break
		}
	}
	var newList []string
	for i := 0; i < 3; i++ {
		newList = append(newList, memList.members[(idx+i)%len(memList.members)])
	}
	monList.members = newList
}

func handleFailOrLeave(ip string) {
	memList.mu.Lock()
	defer memList.mu.Unlock()
	delete(ip)
	updateMonList()
}

// start to monitor
func startMonitor(ips []string) {
	for _, ip := range ips {
		go monitor(ip)
	}
}

// ping all peers in monitor list every 2.5 seconds
func monitor(ip string) {
	defer func() {
		fmt.Println("monitor done!")
	}()
	var pingMsg utils.Message = utils.CreateMsg(localIp, utils.PING, "")
	msg := utils.Msg2Json(pingMsg)
	var rcvMsg = make([]byte, 1024)
	// set a ticker to send ping message periodically
	ticker := time.NewTicker(2500 * time.Millisecond)
	defer ticker.Stop()
	for {
		flag := true
		select {
		case <-ticker.C:
			dstAddr := &net.UDPAddr{IP: net.ParseIP(ip), Port: port}
			// build connection
			conn, err := net.DialUDP("udp", nil, dstAddr)
			if err != nil {
				log.Fatal("Something wrong when build udp conn with ", ip)
			}
			// send message
			_, err = conn.Write(msg)
			if err != nil {
				log.Fatal("Something wrong when send udp packet to", ip)
			}
			// set read deadline for timeout
			conn.SetReadDeadline(time.Now().Add(time.Duration(2000) * time.Millisecond))

			// try to get ack message from target
			n, err := conn.Read(rcvMsg)
			receive := utils.Json2Msg(rcvMsg[:n])
			fmt.Println(receive)
			if err != nil {
				fmt.Println(err)
				// monitor object failed
				ticker.Stop()
				flag = false
			}

			// monitor object is still alive
			fmt.Println("target is still alive", ip)
			// close the connection
			conn.Close()
		}
		// if monitor peer failed, break the loop
		if !flag {
			// fmt.Println("Break")
			break
		}
	}
}

// listen to specific port: 9980 to reply for ping
func handler() {
	defer func() {
		fmt.Println("handler done!")
	}()
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("Something wrong when resolve local address")
	}
	// build conn
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal("Something wrong when listen port")
	}
	defer conn.Close()
	var rcvMsg = make([]byte, 1024)
	for {
		n, srcAddr, err := conn.ReadFromUDP(rcvMsg)
		if err != nil {
			fmt.Println("Error when read udp packet")
		}
		msg := utils.Json2Msg(rcvMsg[:n])
		fmt.Println(msg)
		// handle different types of messages
		switch msg.Type {
		// reply ack message when receive ping message
		case utils.PING:
			ackMsg := utils.CreateMsg(localIp, utils.ACK, "")
			_, err = conn.WriteToUDP(utils.Msg2Json(ackMsg), srcAddr)
			if err != nil {
				fmt.Println("Error when send back Ack message")
			}
		// delete fail node from local when receive fail message
		case utils.FAIL:
			go handleFailOrLeave(msg.SrcIp)
		// delete leave node from local when receive leave message
		case utils.LEAVE:
			go handleFailOrLeave(msg.SrcIp)
		}
	}
}

func main() {
	test := []string{"127.0.0.1"}
	go startMonitor(test)
	go handler()
	for {
	}
}
