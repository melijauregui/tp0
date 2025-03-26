package common

import (
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
	Nombre        string
	Apellido      string
	DNI           string
	Nacimiento    string
	Numero        int
}

// Client Entity that encapsulates how
type Client struct {
	config  ClientConfig
	conn    net.Conn
	Running bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:  config,
		Running: true,
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
	for msgID := 1; msgID <= c.config.LoopAmount && c.Running; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		err_creating_client := c.createClientSocket()
		if err_creating_client != nil {
			if c.Running {
				log.Errorf("action: create_client_socket | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err_creating_client,
				)
			}
			return
		}

		log.Infof("action: create_client_socket | result: success | client_id: %v", c.config.ID)

		// Send the message to the server
		msgSend := fmt.Sprintf("%s,%s,%s,%s,%s,%d", c.config.ID, c.config.Nombre, c.config.Apellido, c.config.DNI, c.config.Nacimiento, c.config.Numero)
		err_sending := SendMessage(c.conn, msgSend)
		if err_sending != nil {
			if c.Running {
				log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err_sending,
				)
			}
			return
		}

		log.Infof("action: apuesta_enviada | result: success | dni: %s | numero: %d",
			c.config.DNI,
			c.config.Numero,
		)

		receivedMessage, err_reading := ReadMessage(c.conn)
		if err_reading != nil {
			if c.Running {
				log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err_reading,
				)
			}
			return
		}

		if receivedMessage == "Apuesta almacenada" {
			log.Infof("action: send_message | result: success | received_message: %v", receivedMessage)
		} else {
			log.Errorf("action: send_message | result: fail | received_message: %v", receivedMessage)
		}

		err_closing := c.conn.Close()

		if err_closing != nil {
			if c.Running {
				log.Errorf("action: connection closed | client_id: %v | signal: %v | result: fail | closed resource: %v", c.config.ID, err_closing)
			}
		} else {
			log.Infof("action: connection closed | result: success | client_id: %v", c.config.ID)
		}

		c.conn = nil
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) StopClient() {
	c.Running = false
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
