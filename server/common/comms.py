import datetime
import logging
from common.utils import Contestant

def recvAll(clientSock, n):
    logging.info("n: {}".format(n))
    bytesRead = 0
    msg = b''
    while bytesRead < n:
        received = clientSock.recv(n - bytesRead)
        msg += received
        bytesRead += len(received)

    if msg == b'\n\n':
        return None
    return msg


def receiveContestantsBatch(clientSock):
    logging.info("Waiting for batch from {}".format(clientSock))
    msg = recvAll(clientSock, 2) #Receive 2 bytes with the length of the message
    if msg == None:
        return None
    msgLen = int.from_bytes(msg, "little")
    data = recvAll(clientSock, msgLen)
    logging.info("Batch Received From client. {} bytes".format(msgLen))
    
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
    logging.info("Sending winners to client {}".format(clientSock))
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
    logging.info("WINNERS VEING SENT. {}".format(msg))
    clientSock.sendall(msg)  
    logging.info("Sent winners to client. {} bytes".format(len(msg)))

