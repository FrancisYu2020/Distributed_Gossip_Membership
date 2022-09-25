package utils

import (
	"encoding/json"
)

const (
	PING = iota
	ACK
	FAIL
	LEAVE
	JOIN
)

type Message struct {
	SrcIp   string `json:srcip`
	SrcID   string `json:srcid`
	Type    int    `json:type`
	Payload string `json:payload`
}

func CreateMsg(ip string, id string, msgType int, payload string) Message {
	return Message{SrcIp: ip, SrcID: id, Type: msgType, Payload: payload}
}

func Json2Msg(data []byte) Message {
	var msg Message
	json.Unmarshal(data, &msg)
	return msg
}

func Msg2Json(msg Message) []byte {
	jsonData, _ := json.Marshal(msg)
	return jsonData
}
