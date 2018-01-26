import socket
import signal
import sys
import os

serversocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
host = socket.gethostbyname(socket.gethostname())
port = 8000

serversocket.bind((host, port))

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
        data = clientsocket.recv(1024).decode()
        print("Connection from %s" % str(address))
        print("Received string: %s" % data)
    except Exception as csr:
        print(csr)
