import socket
import signal
import sys
import os
import time
import threading
import pycurl

serversocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
host = socket.gethostbyname(socket.gethostname())
port = 8005
ra = "/dev/random"


serversocket.bind((host, port))
serversocket.listen(5)

def signal_handler(signal, frame):
    print("Exiting Gracefully!")
    serversocket.close()
    sys.exit(0)

def rasample(SMPL):
    with open("/dev/urandom", 'r') as f:
        data = f.read(SMPL)
    f.close
    return data

def rafsample(SMPL):
    with open("randomfile", 'r') as f:
        data = str(f.read(SMPL))
    f.close
    return data

def test():
    threading.Timer(10.0, test).start()
    print("overwriting random file...")
    try:
        os.remove("randomfile")
    except Exception as rf:
        print(rf)

    f = open("randomfile","a+")
    c = pycurl.Curl()
    c.setopt(pycurl.URL, "https://www.random.org/integers/?num=500&min=1&max=255&col=1&base=10&format=plain&rnd=new")
    ran = str(c.perform())
    print("len:", len(ran))
    f.write(ran)
    f.readline(1)
    """
    for i in range(100):
        raline = rasample(100)
        f.write(raline)
        i = i + 1
    """
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
        rand = rafsample(recd)
        print("sending string:", rand)
        clientsocket.send(rand.encode())
    except Exception as csr:
        print(csr)
