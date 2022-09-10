package common

import (
	"encoding/csv"
	"net"
	"os"
	"strconv"
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
	ID            int
	ServerAddress string
	QueryWaitTime int
	InitWaitTime  int
	TotalFiles    int
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

func (c *Client) checkForOsSingal() bool {
	select {
	case <-c.finished:
		return true
	default:
		return false
	}
}

func (c *Client) StartClientLoop() {
	time.Sleep(time.Duration(c.config.InitWaitTime) * time.Second) //Wait a few seconds so the server is up and listening for connections

	filepath := "./datasets/dataset-" + strconv.Itoa((c.config.ID%c.config.TotalFiles)+1) + ".csv"
	f, err := os.Open(filepath)
	if err != nil {
		log.Errorf("Unable to read input file %v: %v", filepath, err)
		return
	}
	defer f.Close()

	log.Infof("[CLIENT %v] Reading contestants from %v", c.config.ID, filepath)
	csvReader := csv.NewReader(f)
	contestantsList := []Person{}
	contestant, err := csvReader.Read()
	if err != nil {
		log.Errorf("Unable to parse file as CSV for %v: %v", filepath, err)
		return
	}

	c.createClientSocket()
	defer c.conn.Close()

	totalContestants := 0
	totalWinners := 0
	for contestant != nil {
		if c.checkForOsSingal() {
			c.closeClientResources(f)
			return
		}
		p := Person{FirstName: contestant[0], LastName: contestant[1], Document: contestant[2], Birthdate: contestant[3]}
		totalContestants += 1
		contestantsList = append(contestantsList, p)
		if len(contestantsList) == 100 {
			sendContestantsInfo(contestantsList, &c.conn)
			contestantsList = []Person{}
			winnersInBatch, err := receiveServerResponse(&c.conn)
			if err != nil {
				sendStopedProcessingToServer(c)
				sendFinishedToServer(c)
			}
			totalWinners += winnersInBatch
		}
		contestant, err = csvReader.Read()
		if err != nil {
			log.Infof("Finished reading contestants. Sending last batch of contestants")
			sendContestantsInfo(contestantsList, &c.conn)
			contestantsList = []Person{}
			winnersInBatch, err := receiveServerResponse(&c.conn)
			if err != nil {
				sendStopedProcessingToServer(c)
				sendFinishedToServer(c)
			}
			totalWinners += winnersInBatch
			break
		}
	}

	log.Infof("[CLIENT %v] Total contestants: %v", c.config.ID, totalContestants)
	log.Infof("[CLIENT %v] Total winners: %v", c.config.ID, totalWinners)
	log.Infof("[CLIENT %v] Winners percentage: %v", c.config.ID, 100*float32(totalWinners)/float32(totalContestants))

	if c.checkForOsSingal() {
		c.closeClientResources(f)
		return
	}
	sendStopedProcessingToServer(c)

	err = requestTotalWinners(c)
	if err != nil {
		log.Infof("[CLIENT %v] SIGTERM IN requestotalwinners: %v", c.config.ID, err)
		c.closeClientResources(f)
		return
	}

	if c.checkForOsSingal() {
		c.closeClientResources(f)
		return
	}
	sendFinishedToServer(c)
	c.closeClientResources(f)
}

func (c *Client) closeClientResources(f *os.File) {
	log.Infof("[CLIENT %v] Closing socket %v", c.config.ID, c.conn.LocalAddr().String())
	c.conn.Close()
	log.Infof("[CLIENT %v] Closing file", c.config.ID)
	f.Close()
	log.Infof("[CLIENT %v] Closing channel listening for OS signals", c.config.ID)
	close(c.sigs)
	close(c.finished)
}
