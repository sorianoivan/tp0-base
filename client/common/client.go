package common

import (
	"bufio"
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

	sendPersonInfo(c.person, &c.conn)
	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	if err != nil {
		log.Errorf(
			"[CLIENT %v] Error reading from socket. %v.",
			c.config.ID,
			err,
		)
		c.conn.Close()
		return
	}
	log.Infof("[CLIENT %v] Message from server: %v", c.config.ID, msg)
	log.Infof("[CLIENT %v] Closing connection", c.config.ID)
	defer c.conn.Close()
}
