import logging
import os

from common.comms import receiveContestantsBatch, sendTrackerInfo, sendWinnersToClient
from common.utils import is_winner, persist_winners


def request_total_winners(client_socket, tracking_input, tracking_output):
    logging.info("PID {} Sending ? to Tracker process".format(os.getpid()))
    tracking_input.put('?')#Mando consulta
    res = tracking_output.get()
    logging.info("PID {} Received from Tracker process {}".format(os.getpid(), res))
    return res

def handle_client_connection(sockets_queue, file_lock, tracking_input, tracking_output):
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
                    if contestants == b'\n\n':
                        logging.info("PID {} Received finish message from {}".format(os.getpid(), client_socket.getpeername()))
                        break
                    elif contestants == b'\f\f':
                        logging.info("PID {} Sending F to Tracker process".format(os.getpid()))
                        tracking_input.put('F')#Digo que termine de procesar
                        continue
                    elif contestants == b'??':
                        logging.info("PID {} Received query request message from {}".format(os.getpid(), client_socket.getpeername()))
                        info = request_total_winners(client_socket, tracking_input, tracking_output)
                        sendTrackerInfo(client_socket, info)
                        continue

                    winners = list(filter(is_winner, contestants))
                    logging.info("PID {} Persisting winners".format(os.getpid()))
                    file_lock.acquire()
                    persist_winners(winners)
                    file_lock.release()
                    #Send winners to client
                    logging.info("PID {} Sending winners to client".format(os.getpid()))
                    sendWinnersToClient(winners, client_socket)
                    logging.info("PID {} Sending amount of winners ({}) to Tracker process".format(os.getpid(), len(winners)))
                    tracking_input.put(len(winners))
            except Exception as e:
                logging.info("PID {} Error while handling client connection: {}".format(os.getpid(), e))
            finally:
                logging.info("PID {} Closing client socket {}".format(os.getpid(), client_socket))
                client_socket.close()
                # logging.info("PID {} Sending F to Tracker process".format(os.getpid()))
                # tracking_input.put('F')#Digo que termine de procesar