import logging


def receiveContestantInfo(client_sock):
    firstName = receiveFieldInfo(client_sock)
    lastName = receiveFieldInfo(client_sock)
    document = receiveFieldInfo(client_sock)
    birthdate = receiveFieldInfo(client_sock)

    return firstName, lastName, document, birthdate

def receiveFieldInfo(client_sock):
    msg = client_sock.recv(1) #byte indicating length of the field we are going to receive
    msgLen = int.from_bytes(msg, "little")
    logging.info("Msg Length: {}".format(msgLen))
    data = client_sock.recv(msgLen).rstrip().decode('utf-8')#TODO: Agregar loop
    logging.info("Data: {}".format(data))
    return data