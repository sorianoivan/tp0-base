import socket
import logging
import signal
import sys

from multiprocessing import Process, Queue, cpu_count, Lock
from common.client_handler import handle_client_connection

from common.tracker import track_winners

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._sockets_queue = Queue()
        self._tracker_handler = None
        self._tracking_input = Queue()
        self._handlers = []
        signal.signal(signal.SIGTERM, self.__sigterm_handler)


    def run(self):
        tracking_output = Queue()
        self._tracker_handler = Process(target=track_winners, args=(self._tracking_input, tracking_output))
        self._tracker_handler.start()

        file_lock = Lock()
        for i in range(cpu_count()):
            p = Process(target=handle_client_connection, args=(self._sockets_queue, file_lock, self._tracking_input, tracking_output))
            p.start()
            logging.info("Created process with id {}".format(p.pid))
            self._handlers.append(p)

        while True:
            client_socket = self.__accept_new_connection()
            logging.info("Put socket in queue {}".format(client_socket))
            self._sockets_queue.put(client_socket)
            self._tracking_input.put('S')#Add 1 active agency

    def __sigterm_handler(self, *args):
            logging.info("SIGTERM received. Gracefully exiting")
            logging.info("Closing server socket {}".format(self._server_socket))
            self._server_socket.shutdown(socket.SHUT_RDWR)
            self._server_socket.close()

            logging.info("Sending None to the sockets_queue")
            for i in range(len(self._handlers)):
                self._sockets_queue.put(None)

            for p in self._handlers:
                logging.info("Joining process with pid {}".format(p.pid))
                p.join()

            logging.info("Putting None in inut queue so the tracker process ends")
            self._tracking_input.put(None)
            logging.info("Joining tracker process")
            self._tracker_handler.join()

            sys.exit(0)

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
