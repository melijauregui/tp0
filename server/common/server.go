package common

import (
	"fmt"
	"net"
	"strings"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/communication_protocol/common"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type Server struct {
	listener   net.Listener
	clientConn net.Conn
	running    bool
}

type ServerConfig struct {
	Port int
}

func NewServer(config ServerConfig) (*Server, error) {
	addr := fmt.Sprintf(":%d", config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	server := &Server{
		listener: listener,
		running:  true,
	}
	return server, nil
}

// gracefulShutdown handles shutdown signals by closing the client connection
// (if any) and the server listener, then exiting.
func (s *Server) GracefulShutdown() {
	s.running = false
	if s.clientConn != nil {
		s.clientConn.Close()
		log.Infof("action: graceful_shutdown | result: success | msg: client connection closed")
	}
	if s.listener != nil {
		s.listener.Close()
		log.Infof("action: graceful_shutdown | result: success | msg: server listener closed")
	}
	log.Infof("action: graceful_shutdown | result: success | msg: server closed gracefully")
}

func (s *Server) Run() {
	for s.running {
		conn, err := s.acceptNewConnection()
		if err != nil {
			if !s.running {
				break
			}
			log.Errorf("action: accept_connections | result: fail | error: %v", err)
			continue
		}
		s.clientConn = conn
		s.handleClientConnection()
	}
}

// handleClientConnection processes the client connection by reading the message,
// parsing the bets, storing them, and sending a confirmation back.
func (s *Server) handleClientConnection() {
	// Ensure the connection is closed at the end.
	defer func() {
		if s.clientConn != nil {
			s.clientConn.Close()
			log.Infof("action: close_connection of client| result: success")
			s.clientConn = nil
		}
	}()

	msgStr, err_reading_msg := common.ReadMessage(s.clientConn)
	if err_reading_msg != nil {
		if s.running {
			log.Infof("action: receive_message | result: fail | error: %v", err_reading_msg)
		}
		return
	}

	// Parse bets from the message.
	var betList []Bet
	error_in_bets := false
	betsSplit := strings.Split(msgStr, ";")
	for _, bet := range betsSplit {
		betInfo := strings.Split(bet, ",")
		if len(betInfo) < 6 {
			error_in_bets = true
			continue
		}
		newBet, err_creating_bet := NewBet(betInfo[0], betInfo[1], betInfo[2], betInfo[3], betInfo[4], betInfo[5])
		if err_creating_bet != nil {
			if s.running {
				log.Errorf("action: create_bet | result: fail | error: %v", err_creating_bet)
				error_in_bets = true
				continue
			} else {
				return
			}
		}
		betList = append(betList, newBet)
	}

	err_store_bets := StoreBets(betList)
	if err_store_bets != nil {
		if s.running {
			log.Errorf("action: store_bets | result: fail | error: %v", err_store_bets)
		}
		return
	}
	if error_in_bets {
		log.Errorf("action: apuesta_recibida | result: fail | cantidad: %d", len(betList))
	} else {
		log.Infof("action: apuesta_recibida | result: success | cantidad: %d", len(betList))
	}

	msgServer := fmt.Sprintf("%d apuestas almacenadas", len(betList))

	err_sending_msg := common.SendMessage(s.clientConn, msgServer)
	if err_sending_msg != nil {
		log.Errorf("action: sending server message | result: fail | error: %v", err_sending_msg)
	} else {
		log.Infof("action: sending server message | result: success | msg_server: %s", msgServer)
	}
}

// acceptNewConnection waits for a new client connection.
func (s *Server) acceptNewConnection() (net.Conn, error) {
	log.Infof("action: accept_connections | result: in_progress")
	conn, err := s.listener.Accept()
	if err != nil {
		return nil, err
	}
	remoteAddr := conn.RemoteAddr().String()
	log.Infof("action: accept_connections | result: success | ip: %s", remoteAddr)
	return conn, nil
}
