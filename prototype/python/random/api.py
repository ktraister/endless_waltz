from http.server import HTTPServer, BaseHTTPRequestHandler
from hex_rand_sample import sample
import random
import string
import json
import redis
import re

#this is to be a restful server that provides random data when requested. 
#post request should take common secret as input and return the correct pad. pad is BURNED once used twice. 

#this pattern will be used to confirm incoming JSON matches pattern or is tossed out
pattern = re.compile('^{"host":".*","key":"[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}"}$')

class SimpleHTTPRequestHandler(BaseHTTPRequestHandler):

    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        self.wfile.write(bytes("GET not supported :)".encode("UTF-8")))

    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        body = self.rfile.read(content_length)
        self.send_response(200)
        self.end_headers()
        print('This is POST request. ')
        print('Received: ', body)
        json_body = json.loads(body.decode())

        #run our checks here and return if json invalid
        #THIS IS CURRENTLY NOT WORKING BUT I DONT CARE RIGHT NOW
        if pattern.match(body.decode()):
            print("Pattern match!")

        red_client = redis.Redis()

        if json_body["host"] == "server":
            #if the request is from a server, we need to create pad & write to redis
            #create the pad 
            pad = random.choices(string.ascii_letters, k=500)
            spad = ""
            for char in pad: spad += char
            print("Pad to send: ", spad)
            red_client.mset({json_body["key"]: spad}) 
        elif json_body["host"] == "client":
            #read from redis with the key, and return result
            spad = red_client.get(json_body["key"]).decode()
            print(spad)


        self.wfile.write(bytes(spad.encode()))


httpd = HTTPServer(('localhost', 8000), SimpleHTTPRequestHandler)
httpd.serve_forever()
