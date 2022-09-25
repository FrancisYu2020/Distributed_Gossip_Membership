package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"src/utils"
	"strconv"
	"strings"
	"time"
)

type Member struct {
	IP string
	ID string
}

type list struct {
	Members []Member
}

/*******  Global Variable ******/

// const introducer = "127.0.0.1"

var localIp = utils.GetLocalIP()
var localID = utils.GenerateID(localIp)
var port = 9980
var portTCP = 9981
var memList list
var monList list
var operaChan = make(chan string, 1024)
var stopChan = make(chan struct{})
var responChan = make(chan list, 1024)
var listChan = make(chan list, 1024)
var bufferChan = make(chan Member, 1024) // buffer for goroutines to transfer Members

/*******************************/

// delete failed or leaved node from local list
func del(target string) {
	// fmt.Println("------------------")
	// fmt.Println("Cur Members:", memList.Members)
	// fmt.Println("kill Member:", target)
	// kill cur monitors
	if strings.Compare(target, localID) == 0 {
		// do not delete self
		return
	}
	var idx int = -1
	for i, m := range memList.Members {
		if strings.Compare(m.ID, target) == 0 {
			idx = i
			break
		}
	}
	if idx == -1 {
		// do not need to send message now
		// fmt.Println("After Members:", memList.Members)
		// fmt.Println("------------------")
		return
	}
	memList.Members = append(memList.Members[:idx], memList.Members[idx+1:]...)

	idx = -1
	for i, m := range memList.Members {
		if strings.Compare(m.ID, localID) == 0 {
			idx = i
			break
		}
	}
	monList.Members = []Member{}
	// if we do not have at least 3 other Members
	if len(memList.Members) <= 3 {
		monList.Members = append(monList.Members, memList.Members[:idx]...)
		monList.Members = append(monList.Members, memList.Members[idx+1:]...)
	} else {
		var newList []Member
		for i := 1; i <= 2; i++ {
			newList = append(newList, memList.Members[(idx+i)%len(memList.Members)])
		}
		newList = append(newList, memList.Members[(idx-1)%len(memList.Members)])
		monList.Members = newList
	}
	// fmt.Println("After Members:", memList.Members)
	// fmt.Println("------------------")
	return
}

func checkExit(target string) bool {
	operaChan <- "READ"
	curMem := <-listChan
	for _, m := range curMem.Members {
		if strings.Compare(m.ID, target) == 0 {
			return true
		}
	}
	return false
}

// handle operations
func operationsBank() {
	for {
		operation := <-operaChan
		// fmt.Println(operation)
		id := operation[3:]
		if strings.Compare(operation[:3], "DEL") == 0 {
			del(id)
		} else if strings.Compare(operation[:3], "ADD") == 0 {
			memList.Members = append(memList.Members, <-bufferChan)
			// fmt.Println(memList.Members)
		} else if strings.Compare(operation[:3], "MON") == 0 {
			responChan <- monList
		} else if strings.Compare(operation[:4], "READ") == 0 {
			listChan <- memList
			// } else if strings.Compare(operation[:4], "KILL") == 0 {
			// 	for i := 0; i < len(monList.Members)-1; i++ {
			// 		killChan <- "KILL"
			// 	}
		} else if strings.Compare(operation[:7], "RESTART") == 0 {
			// fmt.Println("OK, CLOSE ALL")
			close(stopChan)
			time.Sleep(10 * time.Millisecond)
			// fmt.Println("OK NOW RESTART")
			stopChan = make(chan struct{})
			startMonitor(stopChan)
		}
	}
}

func handleFailOrLeaveMsg(m utils.Message) {
	// delete failed node
	operaChan <- "DEL" + m.Payload
	failMsg := utils.Msg2Json(utils.CreateMsg(localIp, localID, utils.FAIL, m.Payload))
	operaChan <- "MON"
	curMon := <-responChan
	// send fail message to others
	for _, mem := range curMon.Members {
		dstAddr := &net.UDPAddr{IP: net.ParseIP(mem.IP), Port: port}
		// build connection
		conn, err := net.DialUDP("udp", nil, dstAddr)
		if err != nil {
			log.Fatal("Something wrong when build udp conn with ", mem.ID)
		}
		// send message
		_, err = conn.Write(failMsg)
		if err != nil {
			log.Fatal("Something wrong when send udp packet to", mem.ID)
		}
	}
}

// start to monitor
func startMonitor(stopChan <-chan struct{}) {
	// fmt.Println("Try to start all monitors!")
	for _, mon := range monList.Members {
		go func(mon Member) {
			// fmt.Println("Start Monitor", mon.id)
			var pingMsg utils.Message = utils.CreateMsg(localIp, localID, utils.PING, "")
			msg := utils.Msg2Json(pingMsg)
			var rcvMsg = make([]byte, 1024)
			// set a ticker to send ping message periodically
			ticker := time.NewTicker(2500 * time.Millisecond)
			for {
				select {
				case <-ticker.C:
					// dstAddr := &net.UDPAddr{IP: net.ParseIP(mon.ip), Port: port}
					// // build connection
					// fmt.Println(dstAddr)
					conn, err := net.Dial("udp", mon.IP+":"+strconv.Itoa(port))
					if err != nil {
						log.Fatal("Something wrong when build udp conn with ", mon.ID)
					}
					// send message
					_, err = conn.Write(msg)
					// fmt.Println("Ping:", mon.ID, pingMsg)
					if err != nil {
						log.Fatal("Something wrong when send udp packet to", mon.ID)
					}
					// set read deadline for timeout
					conn.SetReadDeadline(time.Now().Add(time.Duration(2000) * time.Millisecond))

					// try to get ack message from target
					_, err = conn.Read(rcvMsg)
					// _ = utils.Json2Msg(rcvMsg[:n])
					if err != nil {
						// fmt.Println("Dead!", mon.ID)
						// monitor object failed
						ticker.Stop()
						// delete the failed node
						operaChan <- "MON"
						failMsg := utils.Msg2Json(utils.CreateMsg(localIp, localID, utils.FAIL, mon.ID))
						// get the monitor list
						curMon := <-responChan
						// send fail message to others
						for _, m := range curMon.Members {
							if strings.Compare(m.ID, mon.ID) == 0 {
								continue
							}
							dstAddr := &net.UDPAddr{IP: net.ParseIP(m.IP), Port: port}
							// build connection
							conn, err := net.DialUDP("udp", nil, dstAddr)
							if err != nil {
								log.Fatal("Something wrong when build udp conn with " + m.ID)
							}
							// send message
							_, err = conn.Write(failMsg)
							if err != nil {
								log.Fatal("Something wrong when send udp packet to" + m.ID)
							}
						}
						operaChan <- "DEL" + mon.ID
						operaChan <- "RESTART"
					}
					// monitor object is still alive
					// fmt.Println("target is still alive", mon.ID)
					// close the connection
					conn.Close()
				case <-stopChan:
					// fmt.Println("Break ", mon.ID)
					return
				}
			}
		}(mon)
	}
}

// ping all peers in monitor list every 2.5 seconds

// listen to specific port: 9980 to reply for ping
func handler() {
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
		// fmt.Println("Received:", msg)
		// handle different types of messages
		switch msg.Type {
		// reply ack message when receive ping message
		case utils.PING:
			ackMsg := utils.CreateMsg(localIp, localID, utils.ACK, "")
			_, err = conn.WriteToUDP(utils.Msg2Json(ackMsg), srcAddr)
			if err != nil {
				fmt.Println("Error when send back Ack message")
			}
		// delete fail node from local when receive fail message
		case utils.FAIL:
			if checkExit(msg.Payload) {
				go handleFailOrLeaveMsg(msg)
			}
		// delete leave node from local when receive leave message
		case utils.LEAVE:
			go handleFailOrLeaveMsg(msg)
		}
	}
}

func commandServer() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Plesase input the command.")
	for {
		text, _, _ := reader.ReadLine()
		command := string(text)
		if strings.Compare(command, "list_mem") == 0 {
			operaChan <- "READ"
			curMem := <-listChan
			for _, mem := range curMem.Members {
				fmt.Println(mem.ID)
			}
		} else if strings.Compare(command, "list_self") == 0 {
			fmt.Println(localID)
		} else if strings.Compare(command, "join") == 0 {
			if err := NodeJoin(); err != nil {
				log.Fatal("Error in joining new node: ", err)
			}
		} else if strings.Compare(command, "leave") == 0 {
			if err := LeaveRequest(); err != nil {
				log.Fatal("Error in joining new node: ", err)
			}
		} else if strings.Compare(command, "list_mon") == 0 {
			operaChan <- "MON"
			curMon := <-responChan
			for _, mem := range curMon.Members {
				fmt.Println(mem.ID)
			}
		}
	}
}

func main() {
	// start the introducer if the indicator file is found
	if IsIntroducer("/home/hangy6/introducer") || IsIntroducer("/home/tian23/introducer") {
		fmt.Println("----------------I am a noble introducer ^_^----------------")
		StartIntroducer()
	} else {
		fmt.Println("----------------I am a pariah node :(----------------")
	}

	// fmt.Println(localID)
	// fmt.Println(localIp)
	// memList.Members = []Member{{"fa22-cs425-2201.cs.illinois.edu", "test5"}, {"fa22-cs425-2202.cs.illinois.edu", "test6"}, {"fa22-cs425-2203.cs.illinois.edu", "test7"}, {"fa22-cs425-2204.cs.illinois.edu", "test4"}, {localIp, localID}}
	// memList.Members = []Member{{localIp, localID}}
	// monList.Members = []Member{{"fa22-cs425-2201.cs.illinois.edu", "test5"}, {"fa22-cs425-2202.cs.illinois.edu", "test6"}, {"fa22-cs425-2204.cs.illinois.edu", "test4"}}
	// monList.Members = []Member{{localIp, localID}}
	go operationsBank()
	operaChan <- "ADD"
	localIP := utils.GetLocalIP()
	bufferChan <- Member{localIP, utils.GenerateID(localIP)}
	go startMonitor(stopChan)
	go handler()
	go commandServer()
	for {
	}
}
