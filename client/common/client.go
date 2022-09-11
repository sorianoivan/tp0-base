package common

import (
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
	InitWaitTime  int
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
	time.Sleep(time.Duration(c.config.InitWaitTime) * time.Second) //Wait a few seconds so the server is up and listening for connections
	c.createClientSocket()
	defer c.conn.Close()
	go func() {
		<-c.sigc
		log.Infof("[CLIENT %v] SIGTERM received. Exiting gracefully", c.config.ID)
		log.Infof("[CLIENT %v] Closing socket %v", c.config.ID, c.conn.LocalAddr().String())
		c.conn.Close()
		os.Exit(0)
	}()

	sendPersonInfo(c.person, &c.conn)
	res := receiveServerResponse(&c.conn)
	if res == 'W' {
		log.Infof("[CLIENT %v] %v %v is a lottery winner", c.config.ID, c.person.FirstName, c.person.LastName)
	} else {
		log.Infof("[CLIENT %v] %v %v is not a lottery winner", c.config.ID, c.person.FirstName, c.person.LastName)
	}
	log.Infof("[CLIENT %v] Closing connection", c.config.ID)
}
