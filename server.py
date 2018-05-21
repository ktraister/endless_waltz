import socket
import signal
import sys
import os
import datetime
import random

serversocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
host = socket.gethostbyname(socket.gethostname())
port = 8000
rahost = "127.0.0.1"
raport = 8005

serversocket.bind((host, port))
serversocket.listen(5)

def signal_handler(signal, frame):
    print("Exiting Gracefully!")
    serversocket.close()
    sys.exit(0)

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

def is_prime(n):
  if n == 2 or n == 3: return True
  if n < 2 or n%2 == 0: return False
  if n < 9: return True
  if n%3 == 0: return False
  r = int(n**0.5)
  f = 5
  while f <= r:
    #print('\t',f)
    if n%f == 0: return False
    if n%(f+2) == 0: return False
    f +=6
  return True

def mkprime():
    n = random.randint(1,100)
    #print(n)
    while not is_prime(n):
        n = random.randint(1,100)
        #print(n)
    return n

def mksec(PRIME, BASE):
    srvsec = random.randint(1,100)
    srvmod = (BASE ** srvsec) % PRIME
    #print("srvsec:", srvsec)
    #print("srvmod:", srvmod)
    return srvmod, srvsec

def dh_est1():
    sharebs = random.randint(1,100)
    sharepm = mkprime()
    servermod, serversec = mksec(sharepm, sharebs)
    response = str(servermod) + "," + str(sharebs) + "," + str(sharepm)
    #print("sharebs:", sharebs)
    #print("sharepm:", sharepm)
    #print("servermod:", servermod)
    #print("serversec:", serversec)
    #print("response:", response)
    return response, servermod, sharepm, serversec

def dh_estf(CLIMOD, SRVSC, SPM):
    CLIMOD = int(CLIMOD)
    SRVSC = int(SRVSC)
    SPM = int(SPM)
    sharsec = (CLIMOD ** SRVSC) % SPM
    #print("sharprime:", SPM)
    #print("srvsc:", SRVSC)
    #print("climod:", CLIMOD)
    #print("FINALLY:", sharsec)
    return sharsec


while 1:
    #signal_handler(signal.SIGINT, signal_handler)
    #signal_handler(signal.SIGTERM, signal_handler)

    #code for clients connecting
    (clientsocket, address) = serversocket.accept()
    print("Client Connected!")

    try:
        #code to receive string from connecting clients
        #recd = clientsocket.recv(1024).decode()
        #recd = int(recd)
        print("Connection from %s" % str(address))
        #print("Received string: %s" % recd)

        #code to perform DH handshake
        r, sm, sp, sc = dh_est1()
        clientsocket.send(r.encode())
        climod = clientsocket.recv(1024).decode()
        #print("climod:", climod)
        ssec = dh_estf(climod, sc, sp)
        print(ssec)
    except Exception as csr:
        print("dh error")
        print(csr)

    #builds a new socket and connects to get pad
    try:
        rasocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        rasocket.connect((rahost, raport))
        msg = "100"
        rasocket.send(msg.encode())
        ppad = rasocket.recv(1024).decode()
        print(ppad)

        #after = operate(ppad, ssec)
        #print(after)

        msg = "Hello client! This is a test string! I am testing my encryption/decryption"
        #msg = "shit"
        print("Message:", msg)
        emsg = encryptstr(msg, ppad)
        print("EMessage:", emsg)
        clientsocket.send(emsg.encode())

        dmsg = decryptstr(emsg, ppad)
        print("Did I decrypt this correctly?:", dmsg)

    except Exception as rsr:
        print("rasocket error")
        print(rsr)


