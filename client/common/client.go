package common

import (
	"net"
	"os"
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
	for msgID := 1; msgID <= c.config.LoopAmount && c.running; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		// Send the message to the server
		//msg := fmt.Sprintf("%s,%d,%s,%s,%s,%s", c.config.DNI, c.config.Numero, c.config.Nombre, c.config.Apellido, c.config.Nacimiento, c.config.ID)
		msg := "30904465,2201,Santiago Lionel,Lorca,1999-03-17,1"
		receivedMessage, err := SendMessage(c.conn, msg)
		if err != nil {
			log.Errorf("action: send_message | result: fail | id:%s | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: apuesta_enviada | result: success | id:%s",
			c.config.ID,
		)
		log.Infof("action: send_message | result: success | received_message: %v", receivedMessage)

		err_closing := c.conn.Close()
		if err_closing != nil {
			log.Errorf("action: connection closed | client_id: %v | signal: %v | result: fail | closed resource: %v", c.config.ID, err)
		}

		c.conn = nil

		time.Sleep(c.config.LoopPeriod)
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
	// filePath := fmt.Sprintf(".data/agency-%s.csv", c.config.ID)
	log.Infof("action: send_batch_message | result: success | client_id: %v | file_path: ", c.config.ID)
	// readFile, err := os.Open(filePath)
	// if err != nil {
	// 	log.Errorf("action: sending batch message | client_id: %v | result: fail | error : %v", c.config.ID, err)
	// 	return
	// }
	// defer readFile.Close()

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
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
