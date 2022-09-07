package common

import (
	"encoding/csv"
	"net"
	"os"
	"os/signal"
	"syscall"
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
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	person Person
	conn   net.Conn
	sigc   chan os.Signal //Channel to listen for OS signals like SIGTERM
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, person Person) *Client {
	client := &Client{
		config: config,
		person: person,
		sigc:   make(chan os.Signal, 1),
	}
	signal.Notify(client.sigc, syscall.SIGTERM)
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
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	c.createClientSocket()
	defer c.conn.Close()
	go func() {
		<-c.sigc
		log.Infof("[CLIENT %v] SIGTERM received. Exiting gracefully", c.config.ID)
		log.Infof("[CLIENT %v] Closing socket %v", c.config.ID, c.conn.LocalAddr().String())
		c.conn.Close()
		os.Exit(0)
	}()

	filepath := "./datasets/dataset-" + c.config.ID + ".csv"
	log.Infof("Filepath: %v", filepath)
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
	for contestant != nil {
		p := Person{FirstName: contestant[0], LastName: contestant[1], Document: contestant[2], Birthdate: contestant[3]}
		contestantsList = append(contestantsList, p)
		if len(contestantsList) == 10 {
			sendContestantsInfo(contestantsList, &c.conn)
			time.Sleep(60 * time.Second)
			contestantsList = []Person{}
			c.conn.Close()
			c.createClientSocket()
		}
		contestant, err = csvReader.Read()
		if err != nil {
			log.Errorf("Finished reading contestants")
			break
		}
	}

	//sendPersonInfo(c.person, &c.conn)
	//res := receiveServerResponse(&c.conn)
	// if res == 'W' {
	// 	log.Infof("[CLIENT %v] %v %v is a lottery winner", c.config.ID, c.person.FirstName, c.person.LastName)
	// } else {
	// 	log.Infof("[CLIENT %v] %v %v is not a lottery winner", c.config.ID, c.person.FirstName, c.person.LastName)
	// }
	log.Infof("[CLIENT %v] Closing connection", c.config.ID)
}
