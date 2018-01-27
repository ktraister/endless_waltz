import socket
import signal
import sys
import os
import datetime

serversocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
host = socket.gethostbyname(socket.gethostname())
port = 8000


serversocket.bind((host, port))
serversocket.listen(5)

def signal_handler(signal, frame):
    print("Exiting Gracefully!")
    serversocket.close()
    sys.exit(0)


while 1:
    #signal_handler(signal.SIGINT, signal_handler)
    #signal_handler(signal.SIGTERM, signal_handler)

    (clientsocket, address) = serversocket.accept()
    print("Client Connected!")

    try:
        recd = clientsocket.recv(1024).decode()
        recd = int(recd)
        print("Connection from %s" % str(address))
        print("Received string: %s" % recd)
        r = "Recieved connection from" + str(address) + "@" + str(datetime.datetime.now())
        clientsocket.send(r.encode())
    except Exception as csr:
        print(csr)
