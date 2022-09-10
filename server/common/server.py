import socket
import logging
import signal
import sys
import os

from common.utils import is_winner, persist_winners
from common.comms import receiveContestantsBatch, sendWinnersToClient
from multiprocessing import Process, Queue, cpu_count, Lock

def handle_client_connection(sockets_queue, file_lock):
        while True:
            logging.info("PID {} Waiting in queue for socket".format(os.getpid()))
            client_socket = sockets_queue.get()
            if client_socket == None:
                logging.info("PID {} received None in sockets queue. Finishing".format(os.getpid())) 
                return
            logging.info("PID {} Read socket from queue {}".format(os.getpid(), client_socket))
            try:
                while True:
                    logging.info("PID {} Waiting for batch".format(os.getpid()))
                    contestants = receiveContestantsBatch(client_socket)
                    if contestants == None:
                        break
                    winners = list(filter(is_winner, contestants))
                    logging.info("PID {} Persisting winners".format(os.getpid()))
                    file_lock.acquire()
                    persist_winners(winners)
                    file_lock.release()
                    #Send winners to client
                    logging.info("PID {} Sending winners to client".format(os.getpid()))
                    sendWinnersToClient(winners, client_socket)
            except Exception as e:
                logging.info("PID {} Error while handling client connection: {}".format(os.getpid(), e))
            finally:
                logging.info("PID {} Closing client socket {}".format(os.getpid(), client_socket))
                client_socket.close()

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._sockets_queue = Queue()
        self._handlers = []
        signal.signal(signal.SIGTERM, self.__sigterm_handler)


    def run(self):
        file_lock = Lock()
        for i in range(cpu_count()):
            p = Process(target=handle_client_connection, args=(self._sockets_queue, file_lock))
            p.start()
            logging.info("Created process with id {}".format(p.pid))
            self._handlers.append(p)

        while True:
            client_socket = self.__accept_new_connection()
            logging.info("Put socket in queue {}".format(client_socket))
            self._sockets_queue.put(client_socket)

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
