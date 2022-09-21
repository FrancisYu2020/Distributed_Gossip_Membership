package main

import (
	"fmt"
	"log"
	"mp2-hangy6-tian23/src/utils"
	"net"
	"strconv"
	"time"
)

/*******  Global Variable ******/

const introducer = "127.0.0.1"

var localIp = utils.GetLocalIP()
var localName = "test"
var pingPort = 9980
var handlePort = 9981
var monList []string

/*******************************/

// start to monitor
func startMonitor(monList []string) {
	for _, ip := range monList {
		go monitor(ip)
	}
}

// ping all peers in monitor list every 2.5 seconds
func monitor(ip string) {
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
			dstAddr := &net.UDPAddr{IP: net.ParseIP(ip), Port: pingPort}
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
			for {
				_, err = conn.Read(rcvMsg)
				data := utils.Json2Msg(rcvMsg)
				if err != nil {
					// monitor object failed
					ticker.Stop()
					flag = false
					break
				}
				if data.Type == utils.ACK {
					break
				}
			}
			// monitor object is still alive
			fmt.Println("%s is still alive", ip)
			// close the connection
			conn.Close()
		}
		// if monitor peer failed, break the loop
		if !flag {
			break
		}
	}
}

func delete(e string)

// listen to specific port: 9980 to reply for ping
func handler() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(handlePort))
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
			go delete(msg.Payload)
		// delete leave node from local when receive leave message
		case utils.LEAVE:
			go delete(msg.Payload)
		}
	}
}
