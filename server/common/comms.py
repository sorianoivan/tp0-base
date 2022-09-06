import logging

def recvAll(clientSock, n):
    bytesRead = 0
    msg = b''
    while bytesRead < n:
        received = clientSock.recv(n - bytesRead)
        msg += received
        bytesRead += len(received)
    return msg


def receiveContestantInfo(clientSock):
    msg = recvAll(clientSock, 2) #Receive 2 bytes with the length of the message
    msgLen = int.from_bytes(msg, "little")
    logging.info("Total Msg Length: {}".format(msgLen))
    data = recvAll(clientSock, msgLen)
    logging.info("Data Received: {}".format(data))
    bytesRead = 0
    firstName, bytesRead = readFieldInfo(data, bytesRead)
    lastName, bytesRead = readFieldInfo(data, bytesRead)
    document, bytesRead = readFieldInfo(data, bytesRead)
    birthdate, bytesRead = readFieldInfo(data, bytesRead)

    return firstName, lastName, document, birthdate

def readFieldInfo(data, bytesRead):
    fieldLen = data[bytesRead]
    bytesRead += 1
    fieldData = data[bytesRead:bytesRead + fieldLen].decode("utf-8")
    bytesRead += fieldLen
    logging.info("fieldData: {}".format(fieldData))
    return fieldData, bytesRead