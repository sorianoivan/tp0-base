package common

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"

	log "github.com/sirupsen/logrus"
)

func receiveServerResponse(conn *net.Conn) byte {
	res, err := bufio.NewReader(*conn).ReadByte()
	if err != nil {
		panic("Error receiving response from server")
	}
	return res
}

func sendContestantsInfo(contestantsList []Person, conn *net.Conn) {
	buf := new(bytes.Buffer)
	for _, contestant := range contestantsList {
		addToBuffer(buf, contestant.FirstName)
		addToBuffer(buf, contestant.LastName)
		addToBuffer(buf, contestant.Document)
		addToBuffer(buf, contestant.Birthdate)
	}

	msgLen := new(bytes.Buffer)
	err := binary.Write(msgLen, binary.LittleEndian, uint16(len(buf.Bytes()))) //Send 2 bytes with the total length of the msg
	if err != nil {
		panic("Failed to write data length to buffer")
	}
	sendAll(msgLen.Bytes(), conn)
	sendAll(buf.Bytes(), conn)
}

func sendPersonInfo(person Person, conn *net.Conn) {
	buf := new(bytes.Buffer)

	addToBuffer(buf, person.FirstName)
	addToBuffer(buf, person.LastName)
	addToBuffer(buf, person.Document)
	addToBuffer(buf, person.Birthdate)

	msgLen := new(bytes.Buffer)
	err := binary.Write(msgLen, binary.LittleEndian, uint16(len(buf.Bytes()))) //Send 2 bytes with the total length of the msg
	if err != nil {
		panic("Failed to write data length to buffer")
	}
	sendAll(msgLen.Bytes(), conn)
	sendAll(buf.Bytes(), conn)
}

func sendAll(data []byte, conn *net.Conn) {
	log.Infof("Going to send %v from socket %v", data, (*conn).LocalAddr().String())
	bytesWritten := 0
	for bytesWritten < len(data) {
		n, err := (*conn).Write(data[bytesWritten:])
		if err != nil {
			panic("Failed to send data to server")
		}
		bytesWritten += n
		log.Infof("Sent %d bytes", n)
	}
}

func addToBuffer(buf *bytes.Buffer, data string) {
	err := binary.Write(buf, binary.LittleEndian, uint8(len(data)))
	if err != nil {
		panic("Failed to write data length to buffer")
	}
	n, _ := buf.WriteString(data)
	if n < len(data) {
		panic("Failed to write data to buffer")
	}
}
