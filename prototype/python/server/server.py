import socket
import ssl
import sys
import requests
import uuid

sys.path.append("../common")
import xor

HOST = "127.0.0.1"
PORT = 60000

server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

server = ssl.wrap_socket(
    server, server_side=True, keyfile="../common/key.pem", certfile="../common/certificate.pem"
)

if __name__ == "__main__":
    server.bind((HOST, PORT))
    server.listen(0)

    while True:
        connection, client_address = server.accept()
        while True:
            try:
                data = connection.recv(1024)
                print("Troubleshooting: ", data)
                if not data:
                    break


                if data.decode("UTF-8") == "HELO":
                    print("Recieved HELO from client, generating key and getting pad...")
                    conn_uuid = str(uuid.uuid4())
                    connection.send(bytes(conn_uuid.encode("UTF-8")))
                    random_req_data = str({"host":"server", "key":conn_uuid}).replace("'", '"')
                    pad = requests.post("http://localhost:8000", data=random_req_data).content
                else: 
                    print("Receiving data from client...")
                    decrypted_data = xor.pad_decrypt(data.decode('utf-8').strip('][').split(', '), pad.decode())
                    print(decrypted_data)
                    msg = ""
                    for char in decrypted_data: msg += char
                    print(f"Received: ", msg)
                    connection.send(bytes("received".encode("utf-8")))
            except Exception as e:
                print(e)
