package common

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type Person struct {
	FirstName string
	LastName  string
	Document  string
	Birthdate string
}

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
}

// Client Entity that encapsulates how
type Client struct {
	config   ClientConfig
	conn     net.Conn
	sigs     chan os.Signal //Channel to listen for OS signals like SIGTERM
	finished chan bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, sigs chan os.Signal, finished chan bool) *Client {
	client := &Client{
		config:   config,
		sigs:     sigs,
		finished: finished,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Fatalf(
			"[CLIENT %v] Could not connect to server. Error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	log.Infof("[CLIENT %v] Connected to server with socket %v", c.config.ID, c.conn.LocalAddr().String())
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	time.Sleep(5 * time.Second) //Wait a few seconds so the server is up and listening for connections
	c.createClientSocket()
	defer c.conn.Close()

	filepath := "./datasets/dataset-" + c.config.ID + ".csv"
	f, err := os.Open(filepath)
	if err != nil {
		log.Errorf("Unable to read input file %v: %v", filepath, err)
		return
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	contestantsList := []Person{}
	contestant, err := csvReader.Read()
	if err != nil {
		log.Errorf("Unable to parse file as CSV for %v: %v", filepath, err)
		return
	}
	totalContestants := 0
	totalWinners := 0
	for contestant != nil {
		select {
		case <-c.finished:
			log.Infof("[CLIENT %v] Closing socket %v", c.config.ID, c.conn.LocalAddr().String())
			c.conn.Close()
			log.Infof("[CLIENT %v] Closing file %v", c.config.ID, filepath)
			f.Close()
			return
		default:
		}
		p := Person{FirstName: contestant[0], LastName: contestant[1], Document: contestant[2], Birthdate: contestant[3]}
		totalContestants += 1
		contestantsList = append(contestantsList, p)
		if len(contestantsList) == 100 {
			sendContestantsInfo(contestantsList, &c.conn)
			contestantsList = []Person{}
			totalWinners += receiveServerResponse(&c.conn)
		}
		contestant, err = csvReader.Read()
		if err != nil {
			log.Infof("Finished reading contestants. Sending last batch of contestants")
			sendContestantsInfo(contestantsList, &c.conn)
			contestantsList = []Person{}
			totalWinners += receiveServerResponse(&c.conn)
			break
		}
	}

	log.Infof("[CLIENT %v] Total contestants: %v", c.config.ID, totalContestants)
	log.Infof("[CLIENT %v] Total winners: %v", c.config.ID, totalWinners)
	log.Infof("[CLIENT %v] Winners percentage: %v", c.config.ID, 100*float32(totalWinners)/float32(totalContestants))

	//Send message of finished processing
	log.Infof("Sending finished processing message to server")
	msg := new(bytes.Buffer)
	err = msg.WriteByte('\f')
	if err != nil {
		panic("Failed to write byte to buffer")
	}
	err = msg.WriteByte('\f')
	if err != nil {
		panic("Failed to write byte to buffer")
	}
	sendAll(msg.Bytes(), &c.conn)

	//Send winners request to server. TODO: meter esto en funcion
	for {
		log.Infof("Sending query message to server")
		msg := new(bytes.Buffer)
		err = msg.WriteByte('?')
		if err != nil {
			panic("Failed to write byte to buffer")
		}
		err = msg.WriteByte('?')
		if err != nil {
			panic("Failed to write byte to buffer")
		}
		sendAll(msg.Bytes(), &c.conn)
		msgType, msgValue := receiveQueryResponse(&c.conn)
		if msgType[0] == 'P' {
			activeAgencies := binary.LittleEndian.Uint16(msgValue)
			log.Infof("There are %v agencies still processing", activeAgencies)
			log.Infof("Going to sleep")
			time.Sleep(15 * time.Second) //TODO: Meter el tiempo en una env variable o config
		} else if msgType[0] == 'W' {
			totalWinners := binary.LittleEndian.Uint16(msgValue)
			log.Infof("There are %v total winners", totalWinners)
			break //Ver si esto sale del for
		}
	}

	//TODO: Meter esto en una funcion tipo finish() y meter el send de los bytes en una funcion de comms
	log.Infof("Sending finish message to server")
	msgLen := new(bytes.Buffer)
	err = msgLen.WriteByte('\n')
	if err != nil {
		panic("Failed to write byte to buffer")
	}
	err = msgLen.WriteByte('\n')
	if err != nil {
		panic("Failed to write byte to buffer")
	}
	sendAll(msgLen.Bytes(), &c.conn)

	log.Infof("[CLIENT %v] Closing socket %v", c.config.ID, c.conn.LocalAddr().String())
	c.conn.Close()
	log.Infof("[CLIENT %v] Closing channel listening for OS signals", c.config.ID)
	close(c.sigs)
}
