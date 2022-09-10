package common

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"

	log "github.com/sirupsen/logrus"
)

func receiveQueryResponse(conn *net.Conn) ([]byte, []byte) {
	log.Infof("Waiting for server response to query")
	msgType := make([]byte, 1)
	_, err := io.ReadFull(*conn, msgType)
	if err != nil {
		panic("Error receiving response length from server")
	}
	log.Infof("Query response type: %v", msgType)

	msg := make([]byte, 2)
	io.ReadFull(*conn, msg)
	if err != nil {
		panic("Error receiving response from server")
	}
	log.Infof("Query Result. %v", msg)
	// log.Infof("Waiting for server response to query")
	// msg := make([]byte, 3)
	// _, err := io.ReadFull(*conn, msg)
	// if err != nil {
	// 	panic("Error receiving response length from server")
	// }
	// log.Infof("Query response: %v", msg)
	return msgType, msg
}

func receiveServerResponse(conn *net.Conn) int {
	log.Infof("Waiting for server response")
	msgLen := make([]byte, 2)
	_, err := io.ReadFull(*conn, msgLen)
	if err != nil {
		panic("Error receiving response length from server")
	}
	length := binary.LittleEndian.Uint16(msgLen)

	msg := make([]byte, length)
	io.ReadFull(*conn, msg)
	if err != nil {
		panic("Error receiving response from server")
	}
	log.Infof("Received response from server. %v bytes", msgLen)
	winners := processMessage(msg)
	return winners
}

func processMessage(msg []byte) int {
	winners := 0
	bytesRead := 0
	for bytesRead < len(msg) {
		firstNameLenght := int(msg[bytesRead])
		bytesRead += 1
		firstName := msg[bytesRead : bytesRead+firstNameLenght]
		bytesRead += firstNameLenght

		lastNameLenght := int(msg[bytesRead])
		bytesRead += 1
		lastName := msg[bytesRead : bytesRead+lastNameLenght]
		bytesRead += lastNameLenght

		docLenght := int(msg[bytesRead])
		bytesRead += 1
		document := msg[bytesRead : bytesRead+docLenght]
		bytesRead += docLenght

		birthdateLength := int(msg[bytesRead])
		bytesRead += 1
		birthdate := msg[bytesRead : bytesRead+birthdateLength]
		bytesRead += birthdateLength
		winners += 1
		log.Infof("Winner: %v, %v, %v, %v", string(firstName), string(lastName), string(document), string(birthdate))
	}
	return winners
}

func sendContestantsInfo(contestantsList []Person, conn *net.Conn) {
	buf := new(bytes.Buffer)
	for _, contestant := range contestantsList {
		addToBuffer(buf, contestant.FirstName)
		addToBuffer(buf, contestant.LastName)
		addToBuffer(buf, contestant.Document)
		addToBuffer(buf, contestant.Birthdate)
	}

	log.Infof("Sending batch to server")

	msgLen := new(bytes.Buffer)
	err := binary.Write(msgLen, binary.LittleEndian, uint16(len(buf.Bytes()))) //Send 2 bytes with the total length of the msg
	if err != nil {
		panic("Failed to write data length to buffer")
	}
	sendAll(msgLen.Bytes(), conn)
	sendAll(buf.Bytes(), conn)
	log.Infof("Sent batch to server. %v bytes", msgLen)
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
	bytesWritten := 0
	for bytesWritten < len(data) {
		n, err := (*conn).Write(data[bytesWritten:])
		if err != nil {
			panic("Failed to send data to server")
		}
		bytesWritten += n
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
