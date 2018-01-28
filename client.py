import socket
import sys
import os
import random

clientsocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
#host = socket.gethostbyname(socket.gethostname())
port = 8000
host = sys.argv[1]
rahost = "192.168.1.100"
raport = 8005
#msg = sys.argv[2]

def encryptstr(MSG, KEY):
    finstr = ''
    for i in range(0, len(MSG)):
        charnum = ord(MSG[i])
        #print("\nmessage[i]:", message[i])
        #print("charnum:", charnum)
        keynum  = ord(KEY[i])
        #print("\nkeychar[i]", otp[i])
        #print("keynum[i]:", keynum)
        #the 128 bit key is working for now, may be a problem in the future
        resnum = (charnum + keynum) % 128
        #print("\nresnum:", resnum)
        reschar = chr(resnum)
        #print("reschar:", reschar)
        finstr = finstr + reschar
    print("Final String:", finstr)
    return finstr

def decryptstr(MSG, KEY):
    finstr = ''
    for i in range(0, len(MSG)):
        charnum = ord(MSG[i])
        #print("\nmessage[i]:", message[i])
        #print("charnum:", charnum)
        keynum  = ord(KEY[i])
        #print("\nkeychar[i]", otp[i])
        #print("keynum[i]:", keynum)
        #reference above comment for 128 modulus
        resnum = (charnum - keynum) % 128
        #print("\nresnum:", resnum)
        reschar = chr(resnum)
        #print("reschar:", reschar)
        finstr = finstr + reschar
    print("Final String:", finstr)
    return finstr

def operate(INPUT, DH):
    result = ''
    rstr = str(INPUT)
    for i in range(0, len(rstr)):
        result = result + chr(ord(rstr[i]) - DH)
    return result

def dh_est1(DATA, SECRET):
    base = DATA.split(',', 2)[1]
    base = int(base)
    #print("sharedbase:", base)
    prime = DATA.split(',', 3)[2]
    prime = int(prime)
    #print("sharedprime:", prime)
    mod1 = (base ** SECRET) % prime
    #print("climod:", mod1)
    return mod1

def dh_est2(CLISEC, DATA):
    srvmod = DATA.split(',', 1)[0]
    srvmod = int(srvmod)
    #print("srvmod:", srvmod)
    prime = DATA.split(',', 3)[2]
    prime = int(prime)
    CLISEC = int(CLISEC)
    #print("srvmod:", srvmod)
    #print("prime:", prime)
    #print("clisec:", CLISEC)
    mod2 = ( srvmod ** CLISEC) % prime
    #print("FINALLY:", mod2)
    return mod2

#code to connect to defined host and port
clientsocket.connect((host, port))
#clientsocket.send(msg.encode())

#code to receive DH handshake details
data = clientsocket.recv(1024).decode()
#print(data)
clisec = random.randint(1, 100)
climod = dh_est1(data, clisec)
climod = str(climod)
clientsocket.send(climod.encode())
#print("climod", climod)
ssec = dh_est2(clisec, data)
print(ssec)

#need to build socket to connect to RAserver and get pre-pad
rasocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
rasocket.connect((rahost, raport))
msg = "100"
rasocket.send(msg.encode())
ppad = rasocket.recv(1024).decode()
print(ppad)

#after = operate(ppad, ssec)
#print(after)

emsg = clientsocket.recv(1024).decode()
print("EMessage:", emsg)
msg = decryptstr(msg, ppad)
print("Message:", msg)

clientsocket.close
