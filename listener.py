import socket
import botocore
import boto3
import logging
import sys
import signal
import datetime
import os

#if you want to change the logging level, change 'level' to logging.{DEBUG,INFO,WARNING,ERROR,CRITICAL}
logging.basicConfig(filename='/var/log/listener.log', level=logging.INFO)
serversocket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
#listening address/port for the listener
#host = "192.168.1.101"
host = socket.gethostbyname(socket.gethostname())
print(host)
port = 8000
s3 = boto3.resource('s3')
ext = "/AutoUpdate/"
mytoken = "f5b52f865335f6aa4a50814ad2bbbd53"

serversocket.bind((host, port))
serversocket.listen(5)
print ("Started server successfully @ %s" % datetime.datetime.now())
logging.info("Started server successfully @ %s" % datetime.datetime.now())

#this piece of code to close socket and exit gracefully on sigint or sigterm
def signal_handler(signal, frame):
    print("Exiting Gracefully!")
    logging.info("Exiting Gracefully @ %s " % datetime.datetime.now())
    serversocket.close()
    sys.exit(0)

#piece of code to delete files on delete events
def delfile(FILENAME):
    try:
        os.remove(FILENAME)
        logging.info("Removing file %s" % FILENAME)
        print("Removing %s" % FILENAME)
    except Exception as d:
        print("Could not delete %s" % FILENAME)
        print("Exception: %s" % d)

#need a piece of code to test for directories in the path of a key
def ldirck(key):
    try:
        pathindex = key.rfind("/")
        path = ext + key[:pathindex]
        print("ldirck -> path:", path)
        mkdir(path)
    except Exception as ldc:
        print(ldc)

#piece of code to make directories if they don't exist
def mkdir(FILENAME):
    try:
        if not os.path.exists(FILENAME):
            os.makedirs(FILENAME)
    except Exception as md:
        print("Could not determine dir status or mkdir!")
        print(md)

#this piece of code will try to download the object it's called with
def getfile(BUCKET_NAME, KEY, FILENAME):
    try:
        #print("key: %s" % KEY)
        #print("bucket: %s" % BUCKET_NAME)
        print(BUCKET_NAME, KEY, FILENAME)
        s3.Bucket(BUCKET_NAME).download_file(KEY, FILENAME)
        print("Downloaded object successfully!")
        logging.info("Downloaded object %s successfully!" % KEY)
    except botocore.exceptions.ClientError as e:
        if e.response['Error']['Code'] == "404":
            print("Object does not exist!")
            logging.warning("Object %s does not exist! (404)" % KEY)
        elif e.response['Error']['Code'] == "403":
            print("Object exists, but you don't have the correct permissions!")
            logging.warning("Unable to access object %s, incorrect permissions! (403)" % KEY)
        else:
            logging.error("ERROR WHILE GETTING FILE: %s" % e)

while 1:
    #catch sigint and sigterm and exit gracefully
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    #this code gets lambda messages
    (clientsocket, address) = serversocket.accept()
    print ("connection found!")

    #adding some error handling for the data recv
    try:
        data = clientsocket.recv(1024).decode()
        logging.info("Connection accepted from: %s " % str(address))
        logging.debug("Received string: %s" % data)
    except Exception as csr:
        logging.error("Unable to handle string received from %s" % str(address))
        logging.error("CliSock Exception: %s" % csr)
        continue

    #print (data)

    #extract remote token and compare with mytoken
    rtoken = data.split(',', 1)[0]
    if rtoken != mytoken:
        print("Token did not match!")
        logging.warning("Unauthorized Token from %s" % str(address))
        continue

    #need to check for improper inputs
    if data.count(',') < 3:
        print("Malformed input string")
        logging.warning("Malformed string from %s" % str(address))
        continue

    #moved this down here so that we only respond to vaidated messages
    #response to confirm recipt of message
    r = str(socket.gethostname()) + " recieved message @ " + str(datetime.datetime.now())
    try:
        clientsocket.send(r.encode())
    except Exception as snd:
        logging.warning("Could not respond to to %s"% str(address))
        continue

    #split up key and bucket strings from data
    #need to add event type to the string sent, specify pull or delete of file
    bucket = data.split(',', 2)[1]
    key = data.split(',', 3)[2]
    action = data.split(',', 4)[3]
    lfile = ext + key
    print("Received Bucket: %s" % bucket)
    print("Received key: %s" % key)
    print("Local File Name: %s" % lfile)
    logging.info("Bucket: %s Key: %s LFile: %s" % (bucket, key, lfile))

    #decide if we're getting or deleting file
    if action == "Put":
        ldirck(key)
        getfile(bucket, key, lfile)
    if action == "Delete":
        delfile(lfile)
    if action == "DirPut":
        mkdir(lfile)

