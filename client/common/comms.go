package common

import (
	"bytes"
	"encoding/binary"
	"net"

	log "github.com/sirupsen/logrus"
)

func sendPersonInfo(person Person, conn *net.Conn) {
	buf := new(bytes.Buffer)

	addToBuffer(buf, *conn, person.FirstName)
	addToBuffer(buf, *conn, person.LastName)
	addToBuffer(buf, *conn, person.Document)
	addToBuffer(buf, *conn, person.Birthdate)
	sendAll(buf.Bytes(), conn)
}

func sendAll(data []byte, conn *net.Conn) {
	log.Infof("Going to send %x from socket %v", data, (*conn).LocalAddr().String())
	bytesWritten := 0
	for bytesWritten < len(data) {
		n, err := (*conn).Write(data[bytesWritten:])
		if err != nil {
			panic("Failed to write data length to buffer")
		}
		bytesWritten += n
		log.Infof("Sent %d bytes", n)
	}
}

func addToBuffer(buf *bytes.Buffer, conn net.Conn, data string) {
	err := binary.Write(buf, binary.LittleEndian, uint8(len(data)))
	if err != nil {
		panic("Failed to write data length to buffer")
	}
	n, _ := buf.WriteString(data)
	if n < len(data) {
		panic("Failed to write data to buffer")
	}
}
