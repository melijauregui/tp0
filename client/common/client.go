package common

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
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
		return err
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
		err_creating_client := c.createClientSocket()
		if err_creating_client != nil {
			if c.running {
				log.Errorf("action: create_client_socket | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err_creating_client,
				)
			}
			return
		}

		// TODO: Modify the send to avoid short-write
		_, err_sending := fmt.Fprintf(
			c.conn,
			"[CLIENT %v] Message N°%v\n",
			c.config.ID,
			msgID,
		)
		if err_sending != nil {
			if c.running {
				log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err_sending,
				)
			}
			return
		}
		log.Infof("action: send_message | result: success | client_id: %v | msg_id: %v",
			c.config.ID,
			msgID,
		)
		msg, err_reading := bufio.NewReader(c.conn).ReadString('\n')

		if err_reading != nil {
			if c.running {
				log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err_reading,
				)
			}
			return
		}

		err_closing := c.conn.Close()

		if err_closing != nil {
			if c.running {
				log.Errorf("action: connection closed | client_id: %v | signal: %v | result: fail | closed resource: %v", c.config.ID, err_closing)
			}
		}

		c.conn = nil

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		// Wait a time between sending one message and the next one
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
}
