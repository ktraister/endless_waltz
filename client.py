import socket
import sys
import os

clientsocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
#host = socket.gethostbyname(socket.gethostname())
port = 8000
host = sys.argv[1]
msg = sys.argv[2]
dhs = 9

def operate(INPUT, DH):
    result = ''
    rstr = str(INPUT)
    for i in range(0, len(rstr)):
        result = result + chr(ord(rstr[i]) - 5)
    return result


clientsocket.connect((host, port))
clientsocket.send(msg.encode())
data = clientsocket.recv(1024)
print(data)
after = operate(data, dhs)
print(after)
clientsocket.close
