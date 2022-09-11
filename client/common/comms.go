package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

func sendStopedProcessingToServer(c *Client) error {
	//Send message of finished processing
	log.Infof("Sending finished processing message to server")
	msg := new(bytes.Buffer)
	err := msg.WriteByte('\f')
	if err != nil {
		return err
	}
	err = msg.WriteByte('\f')
	if err != nil {
		return err
	}
	err = sendAll(msg.Bytes(), &c.conn)
	if err != nil {
		return err
	}
	return nil
}

func sendFinishedToServer(c *Client) error {
	log.Infof("Sending finish message to server")
	msgLen := new(bytes.Buffer)
	err := msgLen.WriteByte('\n')
	if err != nil {
		return err
	}
	err = msgLen.WriteByte('\n')
	if err != nil {
		return err
	}
	err = sendAll(msgLen.Bytes(), &c.conn)
	if err != nil {
		return err
	}
	return nil
}

func requestTotalWinners(c *Client) error {
	for {
		if c.checkForOsSingal() {
			return errors.New("Received SIGTERM while waiting for total winners")
		}
		log.Infof("Sending query message to server")
		msg := new(bytes.Buffer)
		err := msg.WriteByte('?')
		if err != nil {
			return err
		}
		err = msg.WriteByte('?')
		if err != nil {
			return err
		}
		err = sendAll(msg.Bytes(), &c.conn)
		if err != nil {
			return err
		}
		msgType, msgValue, err := receiveQueryResponse(&c.conn)
		if err != nil {
			return err
		}
		if msgType[0] == 'P' {
			activeAgencies := binary.LittleEndian.Uint16(msgValue)
			log.Infof("There are %v agencies still processing", activeAgencies)
			log.Infof("Waiting before making the query again")
			time.Sleep(time.Duration(time.Duration(c.config.QueryWaitTime).Seconds()))
		} else if msgType[0] == 'W' {
			totalWinners := binary.LittleEndian.Uint16(msgValue)
			log.Infof("There are %v total winners", totalWinners)
			break
		}
	}
	return nil
}

func receiveQueryResponse(conn *net.Conn) ([]byte, []byte, error) {
	log.Infof("Waiting for server response to query")
	msgType := make([]byte, 1)
	_, err := io.ReadFull(*conn, msgType)
	if err != nil {
		return nil, nil, err
	}

	msg := make([]byte, 2)
	io.ReadFull(*conn, msg)
	if err != nil {
		return nil, nil, err
	}
	return msgType, msg, nil
}

func receiveServerResponse(conn *net.Conn) (int, error) {
	log.Infof("Waiting for server response")
	msgLen := make([]byte, 2)
	_, err := io.ReadFull(*conn, msgLen)
	if err != nil {
		return -1, nil
	}
	length := binary.LittleEndian.Uint16(msgLen)

	msg := make([]byte, length)
	io.ReadFull(*conn, msg)
	if err != nil {
		return -1, nil
	}
	log.Infof("Received response from server. %v bytes", msgLen)
	winners := processMessage(msg)
	return winners, nil
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

func sendContestantsInfo(contestantsList []Person, conn *net.Conn) error {
	buf := new(bytes.Buffer)
	for _, contestant := range contestantsList {
		err := addToBuffer(buf, contestant.FirstName)
		if err != nil {
			return err
		}
		err = addToBuffer(buf, contestant.LastName)
		if err != nil {
			return err
		}
		err = addToBuffer(buf, contestant.Document)
		if err != nil {
			return err
		}
		err = addToBuffer(buf, contestant.Birthdate)
		if err != nil {
			return err
		}
	}

	log.Infof("Sending batch to server")

	msgLen := new(bytes.Buffer)
	err := binary.Write(msgLen, binary.LittleEndian, uint16(len(buf.Bytes()))) //Send 2 bytes with the total length of the msg
	if err != nil {
		return err
	}
	err = sendAll(msgLen.Bytes(), conn)
	if err != nil {
		return err
	}
	err = sendAll(buf.Bytes(), conn)
	if err != nil {
		return err
	}
	log.Infof("Sent batch to server. %v bytes", msgLen)
	return nil
}

func sendAll(data []byte, conn *net.Conn) error {
	bytesWritten := 0
	for bytesWritten < len(data) {
		n, err := (*conn).Write(data[bytesWritten:])
		if err != nil {
			return err
		}
		bytesWritten += n
	}
	return nil
}

func addToBuffer(buf *bytes.Buffer, data string) error {
	err := binary.Write(buf, binary.LittleEndian, uint8(len(data)))
	if err != nil {
		return err
	}
	n, _ := buf.WriteString(data)
	if n < len(data) {
		return err
	}
	return nil
}
