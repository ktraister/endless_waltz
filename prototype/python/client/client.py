import socket
import ssl
import requests
import sys

sys.path.append("../common")
import xor

SERVER_HOST = "127.0.0.1"
SERVER_PORT = 6000

HOST = "127.0.0.1"
PORT = 0

client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
client.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

client = ssl.wrap_socket(client, keyfile="../common/key.pem", certfile="../common/certificate.pem")

if __name__ == "__main__":
    client.bind((HOST, PORT))
    client.connect((SERVER_HOST, SERVER_PORT))

    print("sending message " + sys.argv[1])
    #the client needs to connect here. we send HELO and get back a key
    client.send(bytes("HELO".encode("UTF-8")))
    random_key = client.read().decode()
    print("Random key from server: ", random_key)
    random_req_data = str({"host":"client", "key":random_key}).replace("'", '"')

    pad = requests.post("http://localhost:8000", data=random_req_data).content
    print("Pad: ", pad.decode())
    cipher_text = xor.pad_encrypt(sys.argv[1], pad.decode())

    client.send(str(cipher_text).encode("utf-8"))
    print(client.read())
