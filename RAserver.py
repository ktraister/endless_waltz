import socket
import signal
import sys
import os
import time
import threading
import requests

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

"""
def rasample(SMPL):
    with open("/dev/urandom", 'r') as f:
        data = f.read(SMPL)
    f.close
    return data
"""

def rafsample(SMPL):
    with open("randomfile", 'r') as f:
        data = str(f.read(SMPL))
    f.close
    return data

def test():
    threading.Timer(30.0, test).start()
    print("overwriting random file...")
    try:
        os.remove("randomfile")
        f = open("randomfile", "w") 
        f.close
    except Exception as rf:
        print(rf)

    r = requests.get("https://www.random.org/integers/?num=500&min=1&max=255&col=1&base=10&format=plain&rnd=new")
    c = r.content
    ran = str(c)
    ran = ran.replace("n", '')
    ran = ran.replace("b'", '')
    ran = ran.replace("'", '')
    ran = ran.replace('\n', '')
    #print("len:", len(ran))
    #print("randomness: ", ran)
    with open("randomfile","w") as randomfile:
        randomfile.write(ran)


test()

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
        rand = rafsample(recd)
        print("sending string:", rand)
        clientsocket.send(rand.encode())
    except Exception as csr:
        print(csr)
