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
        if self._client_socket is not None:
            self._client_socket.close()
            self._client_socket = None
            logging.info('action: graceful_shutdown | result: success | msg: client socket closed')
            
        if self._server_socket:
            self._server_socket.close()
            logging.info('action: graceful_shutdown | result: success | msg: server socket closed')
            
        logging.info('action: graceful_shutdown | result: success | msg: server closed gracefully')

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self.running:
            self._client_socket = self.__accept_new_connection()
            if self._client_socket:
                self.__handle_client_connection()

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            msg_str = read_message(self._client_socket)
            msg = msg_str.split(",")
            bet = Bet(msg[0], msg[1], msg[2], msg[3], msg[4], msg[5])
            store_bets([bet])
            
            logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')
            
            msg_server_success = "Apuesta almacenada"
            logging.info(f'action: sending server message | result: success | msg_server: {msg_server_success} ')
            send_message(self._client_socket, msg_server_success)
            
        except OSError as e:
            if self.running:
                msg_server_error = "Apuesta no almacenada"
                logging.info(f'action: sending server message | result: fail | error: {e} | msg_server: {msg_server_error} ')
                send_message(self._client_socket, msg_server_error)
                logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            if self._client_socket:
                self._client_socket.close()
                self._client_socket = None
                logging.info('action: close_connection of client| result: success')

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """
        try:
            # Connection arrived
            logging.info('action: accept_connections | result: in_progress')
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return c
        except OSError as e:
            if self.running:
                logging.error(f'action: accept_connections | result: fail | error: {e}')
            return None
