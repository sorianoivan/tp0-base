import os
import logging
from common.utils import Contestant

def recvAll(clientSock, n):
    bytesRead = 0
    msg = b''
    while bytesRead < n:
        received = clientSock.recv(n - bytesRead)
        if len(received) == 0:
            raise Exception("Received 0 bytes from server") 
        msg += received
        bytesRead += len(received)

    if msg == b'\n\n':
        logging.info("PID {} Received finish message from {}".format(os.getpid(), clientSock.getpeername()))
        return None

    return msg


def receiveContestantsBatch(clientSock):
    logging.info("PID {} Waiting for batch from {}".format(os.getpid(), clientSock.getpeername()))
    msg = recvAll(clientSock, 2) #Receive 2 bytes with the length of the message
    if msg == None:
        return None
    msgLen = int.from_bytes(msg, "little")
    data = recvAll(clientSock, msgLen)
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

def readFieldInfo(data, bytesRead):
    fieldLen = data[bytesRead]
    bytesRead += 1
    fieldData = data[bytesRead:bytesRead + fieldLen].decode("utf-8")
    bytesRead += fieldLen
    return fieldData, bytesRead

def sendWinnersToClient(winners, clientSock):
    logging.info("PID {} Sending winners to client {}".format(os.getpid(), clientSock.getpeername()))
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

    clientSock.sendall(len(msg).to_bytes(2, 'little'))
    clientSock.sendall(msg)  
    logging.info("PID {} Sent winners to client. {} bytes".format(os.getpid(), len(msg)))

