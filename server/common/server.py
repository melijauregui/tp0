import socket
import logging
import signal
import sys

from common.communication_protocol import read_message, send_message
from common.utils import Bet, load_bets, store_bets

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_socket = None
        self.running = True  # flag para controlar la ejecución del servidor
        signal.signal(signal.SIGTERM, self.graceful_shutdown)
        signal.signal(signal.SIGINT, self.graceful_shutdown)

        
    def graceful_shutdown(self, signum, frame):
        """
        Signal handler to stop the server
        """
        self.running = False
        if self._client_socket:
            self._client_socket.close()
            self._client_socket = None
            logging.info('action: graceful_shutdown | result: success | msg: client socket closed')
            
        if self._server_socket:
            self._server_socket.close()
            logging.info('action: graceful_shutdown | result: success | msg: server socket closed')
            
        logging.info('action: graceful_shutdown | result: success | msg: server closed gracefully')
        sys.exit(0)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        # TODO: Modify this program to handle signal to graceful shutdown
        # the server
        while self.running:
            self._client_socket = self.__accept_new_connection()
            self.__handle_client_connection()

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            # TODO: Modify the receive to avoid short-reads
            msg_str = read_message(self._client_socket)
            bet_list = []
            bets_split = msg_str.split(";")
            for bet in bets_split:
                bet_info = bet.split(",")
                bet_list.append(Bet(bet_info[5],bet_info[2], bet_info[3], bet_info[0], bet_info[4], bet_info[1]))
            store_bets(bet_list)
            
            bets = load_bets()
            for bet in bets:
                logging.info(f'action: apuesta_guardada | bet: {bet.first_name} {bet.last_name}')
            
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bet_list)}')
            
            msg_server = "Apuesta almacenada"
            msg_server = f"{len(msg_server)}:{msg_server}"
            logging.info(f'action: sending server message | result: success | msg_server: {msg_server} ')
            send_message(self._client_socket, msg_server)
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            if self._client_socket:
                self._client_socket.close()
                logging.info('action: close_connection of client| result: success')
                self.client_socket = None

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
