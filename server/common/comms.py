import os
import logging
from common.utils import Contestant

special_connection_msgs = [b'\n\n', b'\f\f', b'??']

def recvAll(client_sock, n):
    bytesRead = 0
    msg = b''
    while bytesRead < n:
        received = client_sock.recv(n - bytesRead)
        if len(received) == 0:
            raise Exception("Received 0 bytes from server") 
        msg += received
        bytesRead += len(received)

    return msg


def receiveContestantsBatch(client_sock):
    logging.info("PID {} Waiting for batch from {}".format(os.getpid(), client_sock.getpeername()))
    msg = recvAll(client_sock, 2)
    if msg in special_connection_msgs:
        return msg
    
    msgLen = int.from_bytes(msg, "little")
    data = recvAll(client_sock, msgLen)
    logging.info("PID {} Batch Received From client. {} bytes".format(os.getpid(), msgLen))
    
    bytesRead = 0
    contestants = []
    while bytesRead < msgLen:
        firstName, bytesRead = readFieldInfo(data, bytesRead)
        lastName, bytesRead = readFieldInfo(data, bytesRead)
        document, bytesRead = readFieldInfo(data, bytesRead)
        birthdate, bytesRead = readFieldInfo(data, bytesRead)
        contestants.append(Contestant(firstName, lastName, document, birthdate))

    return contestants

def readFieldInfo(data, bytes_read):
    field_len = data[bytes_read]
    bytes_read += 1
    field_data = data[bytes_read:bytes_read + field_len].decode("utf-8")
    bytes_read += field_len
    return field_data, bytes_read

def sendWinnersToClient(winners, client_sock):
    logging.info("PID {} Sending winners to client {}".format(os.getpid(), client_sock.getpeername()))
    msg = bytearray()
    for winner in winners:
        msg.append(len(winner.first_name.encode('utf-8')))
        msg.extend(winner.first_name.encode('utf-8'))
        msg.append(len(winner.last_name.encode('utf-8')))
        msg.extend(winner.last_name.encode('utf-8'))
        msg.append(len(winner.document))
        msg.extend(winner.document.encode('utf-8'))
        birthdate = winner.birthdate.strftime('%Y-%m-%d')
        msg.append(len(birthdate))
        msg.extend(birthdate.encode('utf-8'))

    client_sock.sendall(len(msg).to_bytes(2, 'little'))
    client_sock.sendall(msg)  
    logging.info("PID {} Sent winners to client. {} bytes".format(os.getpid(), len(msg)))

def sendTrackerInfo(client_sock, info):
    client_sock.sendall(info['type'].encode('utf-8'))
    client_sock.sendall(info['value'].to_bytes(2, 'little'))
    logging.info("PID {} Sent Tracker info to client.".format(os.getpid()))
