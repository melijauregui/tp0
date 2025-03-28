package common

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/communication_protocol/common"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type Server struct {
	listener           net.Listener
	numberOfAgencies   int
	running            bool
	runningLock        sync.Mutex
	clientsConn        map[string]net.Conn
	lockClientsConn    sync.Mutex
	agenciesWaiting    map[int]string
	winnerRevealed     bool
	lockWinnerRevealed sync.Mutex
	betsLock           sync.Mutex
	wg                 sync.WaitGroup
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
		clientsConn:      map[string]net.Conn{},
	}
	return server, nil
}

func (s *Server) Run() {
	for s.IsRunning() {
		conn, ip, err := s.acceptNewConnection()
		if err != nil {
			if !s.IsRunning() {
				log.Infof("action: accepted connection fail for quitting | result: success")
				return
			}
			log.Errorf("action: accept_connections | result: fail | error: %v", err)
			continue
		}
		s.wg.Add(1)
		go s.handleClientConnection(conn, ip)

		s.canRevealWinners()
	}
	s.wg.Wait()
}

// handleClientConnection processes the client connection by reading the message,
// parsing the bets, storing them, and sending a confirmation back.
func (s *Server) handleClientConnection(clientConn net.Conn, ip string) {
	defer s.wg.Done()
	s.lockClientsConn.Lock()
	s.clientsConn[ip] = clientConn
	s.lockClientsConn.Unlock()
	// Ensure the connection is closed at the end.
	defer func() {
		s.lockClientsConn.Lock()
		if s.clientsConn[ip] != nil {
			s.clientsConn[ip].Close()
			delete(s.clientsConn, ip)
			log.Infof("action: close_connection of client| result: success")
		}
		s.lockClientsConn.Unlock()
	}()

	msgStr, err_reading_msg := common.ReadMessage(clientConn)
	if err_reading_msg != nil {
		if s.IsRunning() {
			log.Infof("action: receive_message | result: fail | error: %v", err_reading_msg)
		}
		return
	}

	if strings.Contains(msgStr, "Winners, please?") {
		s.handleAgencyWaitingMessage(clientConn, msgStr)
	} else {
		s.handleStoreBetsMessage(clientConn, msgStr)
	}
}

func (s *Server) handleStoreBetsMessage(clientConn net.Conn, msgStr string) {
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
			if s.IsRunning() {
				log.Errorf("action: create_bet | result: fail | error: %v", err_creating_bet)
				error_in_bets = true
				continue
			} else {
				return
			}
		}
		betList = append(betList, newBet)
	}
	s.betsLock.Lock()
	err_store_bets := StoreBets(betList)
	s.betsLock.Unlock()
	if err_store_bets != nil {
		if s.IsRunning() {
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
	err_sending_msg := common.SendMessage(clientConn, msgServer)
	if err_sending_msg != nil {
		if s.IsRunning() {
			log.Errorf("action: sending server message | result: fail | error: %v", err_sending_msg)
		}
	} else {
		log.Infof("action: sending server message | result: success | msg_server: %s", msgServer)
	}
}

func (s *Server) handleAgencyWaitingMessage(clientConn net.Conn, msgStr string) {
	msgSplit := strings.Split(msgStr, ",")
	agency, err_convert := strconv.Atoi(msgSplit[0])

	if err_convert != nil {
		log.Errorf("action: convert_agency | result: fail | error: %v", err_convert)
		return
	}
	msg := ""
	s.lockWinnerRevealed.Lock()
	if s.winnerRevealed {
		msg = s.agenciesWaiting[agency]
		if len(msg) > 0 {
			msg = msg[0 : len(msg)-1]
		}
		log.Infof("action: send winners agency | result: success | agency: %d", agency)
		delete(s.agenciesWaiting, agency)
	} else {
		msg = "No winners yet"
		s.agenciesWaiting[agency] = ""
		log.Infof("action: waiting agency | result: success | agency: %d", agency)
	}
	s.lockWinnerRevealed.Unlock()

	err_sending_msg := common.SendMessage(clientConn, msg)
	if err_sending_msg != nil {
		if s.IsRunning() {
			log.Errorf("action: send client message | result: fail | error: %v | msg_server:", err_sending_msg, msg)
		}
	} else {
		log.Infof("action: send client message | result: success | msg_server: %s", msg)
	}

	s.lockWinnerRevealed.Lock()
	agenciesWaiting := len(s.agenciesWaiting)
	s.lockWinnerRevealed.Unlock()
	if agenciesWaiting == 0 && s.IsRunning() {
		s.GracefulShutdown()
	}

}

func (s *Server) canRevealWinners() {
	s.lockWinnerRevealed.Lock()
	if !s.winnerRevealed && len(s.agenciesWaiting) == s.numberOfAgencies {
		log.Infof("action: sorteo | result: success")
		s.betsLock.Lock()
		bets, err_loading_bets := LoadBets()
		s.betsLock.Unlock()
		if err_loading_bets != nil {
			if s.IsRunning() {
				log.Errorf("action: load_bets | result: fail | error: %v", err_loading_bets)
			}
			return
		}

		for _, bet := range bets {
			if !s.IsRunning() {
				break
			}
			if HasWon(bet) {
				s.agenciesWaiting[bet.Agency] += fmt.Sprintf("%s;", bet.Document)
			}
		}
		s.winnerRevealed = true
	}
	s.lockWinnerRevealed.Unlock()
}

// acceptNewConnection waits for a new client connection.
func (s *Server) acceptNewConnection() (net.Conn, string, error) {
	log.Infof("action: accept_connections | result: in_progress")
	conn, err := s.listener.Accept()
	if err != nil {
		return nil, "", err
	}
	remoteAddr := conn.RemoteAddr().String()
	log.Infof("action: accept_connections | result: success | ip: %s", remoteAddr)
	return conn, remoteAddr, nil
}

// gracefulShutdown handles shutdown signals by closing the client connection
// (if any) and the server listener, then exiting.
func (s *Server) GracefulShutdown() {
	s.runningLock.Lock()
	s.running = false
	s.runningLock.Unlock()
	log.Infof("action: graceful_shutdown | result: in_progress")

	s.lockClientsConn.Lock()
	for ip, conn := range s.clientsConn {
		conn.Close()
		s.clientsConn[ip] = nil
		log.Infof("action: close_connection of client | result: success | ip: %v", ip)
	}
	s.lockClientsConn.Unlock()

	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			log.Infof("action: listener.Close() finished | result: success| error: %v", err)
		} else {
			log.Infof("action: listener.Close() finished | result: success")
		}
	}
	log.Infof("action: graceful_shutdown | result: success | msg: server closed gracefully")
}

func (s *Server) IsRunning() bool {
	s.runningLock.Lock()
	running := s.running
	defer s.runningLock.Unlock()
	return running
}
