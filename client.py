import socket
import sys
import os
import random

clientsocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
#host = socket.gethostbyname(socket.gethostname())
port = 8000
host = sys.argv[1]
msg = sys.argv[2]
#clisc var for holding DH client value, must be generated new every time
#clisc = random.randint(1,100)
#dhs is final result, again, new every time
#dhs = 9

def operate(INPUT, DH):
    result = ''
    rstr = str(INPUT)
    for i in range(0, len(rstr)):
        result = result + chr(ord(rstr[i]) - DH)
    return result

def dh_est1(DATA, SECRET):
    base = DATA.split(',', 2)[1]
    base = int(base)
    print("sharedbase:", base)
    prime = DATA.split(',', 3)[2]
    prime = int(prime)
    print("sharedprime:", prime)
    mod1 = (base ** SECRET) % prime
    print("climod:", mod1)
    return mod1

def dh_est2(CLISEC, DATA):
    srvmod = DATA.split(',', 1)[0]
    srvmod = int(srvmod)
    print("srvmod:", srvmod)
    prime = DATA.split(',', 3)[2]
    prime = int(prime)
    CLISEC = int(CLISEC)
    print("srvmod:", srvmod)
    print("prime:", prime)
    print("clisec:", CLISEC)
    mod2 = ( srvmod ** CLISEC) % prime
    print("FINALLY:", mod2)

clientsocket.connect((host, port))
clientsocket.send(msg.encode())

data = clientsocket.recv(1024).decode()
print(data)

clisec = random.randint(1, 100)
climod = dh_est1(data, clisec)
climod = str(climod)
clientsocket.send(climod.encode())
print("climod", climod)
dh_est2(clisec, data)

#after = operate(data, dhs)
#print(after)

clientsocket.close
