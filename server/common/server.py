import socket
import logging
import signal
import sys

from common.utils import is_winner
from common.comms import receiveContestantsBatch, sendWinnersToClient

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_socket = None
        signal.signal(signal.SIGTERM, self.__sigterm_handler)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while True:
            self._client_socket = self.__accept_new_connection()
            self.__handle_client_connection()

    def __sigterm_handler(self, *args):
            logging.info("SIGTERM received. Gracefully exiting")
            logging.info("Closing server socket {}".format(self._server_socket))
            self._server_socket.shutdown(socket.SHUT_RDWR)
            self._server_socket.close()
            if (self._client_socket.fileno() != -1):
                logging.info("Closing client socket {}".format(self._client_socket))
                self._client_socket.shutdown(socket.SHUT_RDWR)
                self._client_socket.close()

            sys.exit(0)

    def __handle_client_connection(self):
        try:
            while True:
                contestants = receiveContestantsBatch(self._client_socket)
                if contestants == None:
                    break
                winners = filter(is_winner, contestants)
                #Send winners to client
                sendWinnersToClient(winners, self._client_socket)
        except OSError:
            logging.info("Error while reading socket {}".format(self.client_sock))
        finally:
            self._client_socket.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info("Proceed to accept new connections")
        c, addr = self._server_socket.accept()
        logging.info('Got connection from {}'.format(addr))
        return c
