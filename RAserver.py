import socket
import signal
import sys
import os
import time
import threading

serversocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
host = socket.gethostbyname(socket.gethostname())
port = 8000
ra = "/dev/random"


serversocket.bind((host, port))
serversocket.listen(5)

def signal_handler(signal, frame):
    print("Exiting Gracefully!")
    serversocket.close()
    sys.exit(0)

def rasample(SMPL):
    with open("/dev/random", 'rb') as f:
        #print(repr(f.read(SMPL)))
        data = repr(f.read(SMPL))
    f.close
    return data

def test():
    threading.Timer(10.0, test).start()
    print("overwriting random file...")
    os.remove("randomfile")
    f = open("randomfile","a+")
    for i in range(100):
        raline = rasample(100)
        f.write(raline)
        i = i + 1
    f.close




while 1:
    #signal_handler(signal.SIGINT, signal_handler)
    #signal_handler(signal.SIGTERM, signal_handler)

    test()

    (clientsocket, address) = serversocket.accept()
    print("Client Connected!")

    try:
        recd = clientsocket.recv(1024).decode()
        recd = int(recd)
        print("Connection from %s" % str(address))
        print("Received string: %s" % recd)
        rand = rasample(recd)
        clientsocket.send(rand.encode())
    except Exception as csr:
        print(csr)
