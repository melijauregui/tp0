package common

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/communication_protocol/common"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type Server struct {
	listener         net.Listener
	clientConn       net.Conn
	running          bool
	agenciesWaiting  map[int]string
	winnerRevealed   bool
	numberOfAgencies int
}

type ServerConfig struct {
	Port             int
	NumberOfAgencies int
}

func NewServer(config ServerConfig) (*Server, error) {
	addr := fmt.Sprintf(":%d", config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	server := &Server{
		listener:         listener,
		running:          true,
		winnerRevealed:   false,
		agenciesWaiting:  map[int]string{},
		numberOfAgencies: config.NumberOfAgencies,
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

		if !s.winnerRevealed && len(s.agenciesWaiting) == s.numberOfAgencies {
			log.Infof("action: sorteo | result: success")
			s.winnerRevealed = true
			s.revealWinners()
		}
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

	if strings.Contains(msgStr, "Winners, please?") {
		s.handleAgencyWaitingMessage(msgStr)
	} else {
		s.handleStoreBetsMessage(msgStr)
	}
}

func (s *Server) handleStoreBetsMessage(msgStr string) {
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

func (s *Server) handleAgencyWaitingMessage(msgStr string) {
	msgSplit := strings.Split(msgStr, ",")
	agency, err_convert := strconv.Atoi(msgSplit[0])

	if err_convert != nil {
		log.Errorf("action: convert_agency | result: fail | error: %v", err_convert)
		return
	}
	msg := ""
	if s.winnerRevealed {
		msg = s.agenciesWaiting[agency]
		if len(msg) > 0 {
			msg = msg[0 : len(msg)-1]
		}
		log.Infof("action: send winners agency | result: success | agency: %d", agency)
	} else {
		msg = "No winners yet"
		s.agenciesWaiting[agency] = ""
		log.Infof("action: waiting agency | result: success | agency: %d", agency)
	}

	log.Infof("action: send msg to waiting agency | result: success | msg: %s", msg)

	err_sending_msg := common.SendMessage(s.clientConn, msg)
	if err_sending_msg != nil {
		log.Errorf("action: send client message | result: fail | error: %v", err_sending_msg)
	} else {
		log.Infof("action: send client message | result: success | msg_server: %s", msg)
	}

}

func (s *Server) revealWinners() {
	bets, err_loading_bets := LoadBets()
	if err_loading_bets != nil {
		log.Errorf("action: load_bets | result: fail | error: %v", err_loading_bets)
		return
	}
	for _, bet := range bets {
		if !s.running {
			break
		}
		if HasWon(bet) {
			s.agenciesWaiting[bet.Agency] += fmt.Sprintf("%s;", bet.Document)
		}
	}
	log.Infof("action: reveal_winners | result: success")
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
