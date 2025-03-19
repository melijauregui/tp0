package common

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID             string
	ServerAddress  string
	LoopAmount     int
	LoopPeriod     time.Duration
	BatchMaxAmount int
}

// Client Entity that encapsulates how
type Client struct {
	config  ClientConfig
	conn    net.Conn
	running bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:  config,
		running: true,
	}

	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	if c.running {
		c.SendBatchMessage()
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) StopClient() {
	c.running = false
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			log.Errorf("action: connection closed | client_id: %v | signal: %v | result: fail | closed resource: %v", c.config.ID, err)
		} else {
			log.Infof("action: graceful_shutdown client connection | result: success | client_id: %v", c.config.ID)
		}
		c.conn = nil
	}

	log.Infof("action: graceful_shutdown | result: success | client_id: %v", c.config.ID)
	os.Exit(0)

}

func (c *Client) SendBatchMessage() {
	filePath := fmt.Sprintf(".data/agency-%s.csv", c.config.ID)
	readFile, err := os.Open(filePath)
	if err != nil {
		log.Errorf("action: sending batch message | client_id: %v | result: fail | error : %v", c.config.ID, err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	msg := ""
	batchSize := 0
	bet := []string{}

	for fileScanner.Scan() {
		fileLine := fileScanner.Text()
		bet = strings.Split(fileLine, ",")
		msg += fmt.Sprintf("%s,%s,%s,%s,%s,%s;", bet[2], bet[4], bet[0], bet[1], bet[3], c.config.ID)
		if batchSize == 0 {
			c.createClientSocket()
			batchSize++
		} else if batchSize < c.config.BatchMaxAmount-1 {
			batchSize++
		} else {
			c.SendBatchMessage2(bet, msg[0:len(msg)-1])
			batchSize = 0
			msg = ""
			time.Sleep(c.config.LoopPeriod)
		}
	}

	readFile.Close()

	if len(msg) > 0 {
		c.SendBatchMessage2(bet, msg[0:len(msg)-1])
	}
	time.Sleep(5 * time.Second)

}

func (c *Client) SendBatchMessage2(bet []string, msg string) {

	log.Infof("action: send_message_started | result: success | msg: %s", msg)
	receivedMessage, err := SendMessage(c.conn, msg)
	if err != nil {
		log.Errorf("action: send_message | result: fail | id: %s | dni: %v | error: %v",
			c.config.ID,
			bet[2],
			err,
		)
		return
	}

	log.Infof("action: apuesta_enviada | result: success | id: %s | dni: %s",
		c.config.ID,
		bet[2],
	)
	log.Infof("action: apuesta_enviada | result: success | received_message: %v", receivedMessage)

	err_closing := c.conn.Close()
	if err_closing != nil {
		log.Errorf("action: connection closed | client_id: %v | signal: %v | result: fail | closed resource: %v", c.config.ID, err)
	}
	c.conn = nil

}
