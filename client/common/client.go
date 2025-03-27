package common

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/communication_protocol/common"
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
	Running bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:  config,
		Running: true,
		// fileReader: nil,
	}

	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	if c.Running {
		conn, err := net.Dial("tcp", c.config.ServerAddress)
		if err != nil {
			return err
		}
		c.conn = conn
	}
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	if c.Running {
		c.SendBatchMessages()

		intentos := 0
		for intentos < c.config.LoopAmount && c.Running {
			err_wating_winners := c.WaitForWinners()
			if err_wating_winners != nil && c.Running {
				log.Errorf("action: waiting_winners | result: fail | client_id: %v | error: %v", c.config.ID, err_wating_winners)
				intentos++
			} else {
				break
			}
		}
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) StopClient() {
	c.Running = false
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			log.Errorf("action: connection closed | result: fail | client_id: %v | signal: %v", c.config.ID, err)
		} else {
			log.Infof("action: graceful_shutdown client connection | result: success | client_id: %v", c.config.ID)
		}
		c.conn = nil
	}

	log.Infof("action: graceful_shutdown | result: success | client_id: %v", c.config.ID)
}

func (c *Client) SendBatchMessages() {
	filePath := fmt.Sprintf(".data/agency-%s.csv", c.config.ID)
	readFile, err_opening_file := os.Open(filePath)
	if err_opening_file != nil {
		log.Errorf("action: sending batch message | client_id: %v | result: fail | error : %v", c.config.ID, err_opening_file)
	}

	defer func() {
		if error_closing_file := readFile.Close(); error_closing_file != nil {
			log.Errorf("action: closing file | result: fail | client_id: %v | error: %v", c.config.ID, error_closing_file)
		} else {
			log.Infof("action: closing file | result: success | client_id: %v ", c.config.ID)
		}
	}()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	msg := ""
	batchSize := 0

	for fileScanner.Scan() && c.Running {
		fileLine := fileScanner.Text()
		bet := strings.Split(fileLine, ",")
		if len(bet) != 5 {
			log.Errorf("action: sending batch message | result: fail | client_id: %v | error: invalid bet format", c.config.ID)
			continue
		}
		msg += fmt.Sprintf("%s,%s,%s,%s,%s,%s;", c.config.ID, bet[0], bet[1], bet[2], bet[3], bet[4])
		if batchSize == 0 {
			err_creating_client := c.createClientSocket()
			if err_creating_client != nil {
				if c.Running {
					log.Errorf("action: sending batch message | result: fail | client_id: %v | error: %v", c.config.ID, err_creating_client)
				}
				return
			}
			batchSize++
		} else if batchSize < c.config.BatchMaxAmount-1 {
			batchSize++
		} else {
			c.SendBatchMessage(msg[0 : len(msg)-1])
			batchSize = 0
			msg = ""
		}
	}

	if len(msg) > 0 && c.Running {
		c.SendBatchMessage(msg[0 : len(msg)-1])
	}

}

func (c *Client) SendBatchMessage(msg string) {

	log.Infof("action: send_message_started | result: success | msg: %s", msg)
	err_sending_msg := common.SendMessage(c.conn, msg)
	if err_sending_msg != nil {
		if c.Running {
			log.Errorf("action: send_message | result: fail | id: %s | error: %v",
				c.config.ID,
				err_sending_msg,
			)
		}
		return
	}
	log.Infof("action: apuesta_enviada | result: success | id: %s",
		c.config.ID,
	)

	receivedMessage, err_reading_msg := common.ReadMessage(c.conn)
	if err_reading_msg != nil {
		if c.Running {
			log.Errorf("action: read_message | result: fail | id: %s | error: %v",
				c.config.ID,
				err_reading_msg,
			)
		}
		return
	}

	if receivedMessage != fmt.Sprintf("%d apuestas almacenadas", len(strings.Split(msg, ";"))) {
		log.Errorf("action: apuesta_enviada | result: fail | id: %s | received_message: %v",
			c.config.ID,
			receivedMessage,
		)
	} else {
		log.Infof("action: apuesta_enviada | result: success | id: %s | received_message: %v",
			c.config.ID,
			receivedMessage,
		)
	}

	if c.conn != nil {
		err_closing := c.conn.Close()
		if err_closing != nil {
			if c.Running {
				log.Errorf("action: connection closed | result: fail | client_id: %v | signal: %v | closed resource: %v", c.config.ID, err_closing)
			}
		}
		log.Infof("action: connection closed | result: success | client_id: %v ", c.config.ID)
		c.conn = nil
	}
}

func (c *Client) WaitForWinners() error {
	log.Infof("action: waiting_winners | result: in_progress | client_id: %v", c.config.ID)
	knowsWinners := false
	i := 0
	for !knowsWinners && c.Running {
		err_creating_socket := c.createClientSocket()
		if err_creating_socket != nil {
			return err_creating_socket
		}
		log.Infof("action: create_socket | result: success | client_id: %v", c.config.ID)
		msg := fmt.Sprintf("%s,%s;", c.config.ID, "Winners, please?")
		err_sending_msg := common.SendMessage(c.conn, msg)
		if err_sending_msg != nil {
			return err_sending_msg
		}

		log.Infof("action: send_message | result: success | id: %s | message: %s",
			c.config.ID,
			msg,
		)

		receivedMessage, err_reading_msg := common.ReadMessage(c.conn)
		if err_reading_msg != nil {
			return err_reading_msg
		}

		if receivedMessage == "No winners yet" {
			log.Infof("action: winners_received | result: success | id: %s | received_message: '%v'",
				c.config.ID,
				receivedMessage,
			)
		} else {
			knowsWinners = true
			log.Infof("action: winners_received | result: success | id: %s | received_message: '%v'",
				c.config.ID,
				receivedMessage,
			)
			var number_winners int
			if receivedMessage != "" {
				number_winners = len(strings.Split(receivedMessage, ";"))
			} else {
				number_winners = 0
			}

			log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d | id: %s ",
				number_winners,
				c.config.ID,
			)
		}
		if c.conn != nil {
			err_closing := c.conn.Close()
			if err_closing != nil && c.Running {
				log.Errorf("action: connection closed | result: fail | client_id: %v | signal: %v | closed resource: %v", c.config.ID, err_closing)
			}
			log.Infof("action: connection closed | result: success | client_id: %v", c.config.ID)
			c.conn = nil
		}
		if !knowsWinners && c.Running {
			log.Infof("action: waiting_winners sleep | result: in_progress | client_id: %v", c.config.ID)
			i++
			time.Sleep(time.Duration(i) * time.Second)
		}
	}
	return nil
}
