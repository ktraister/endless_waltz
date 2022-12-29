package main

import (
    "bytes"
    "net/http"
    "encoding/json"
    "fmt"
    "log"
    "crypto/tls"
    "flag"
)

type Random_Req struct {
    Host string
    UUID string
}

func main() {
    log.SetFlags(log.Lshortfile)
    //lets setup our flags here
    msgPtr := flag.String("message", "", "a message to encrypt and send")
    hostPtr := flag.String("host", "localhost", "the server to send the message to")
    flag.Parse()

    if len(*msgPtr) > 4096 { panic("We dont support this yet!") }

    conf := &tls.Config{
	 // FIx tHis ItS BADDDD
         InsecureSkipVerify: true,
    }

    conn, err := tls.Dial("tcp", fmt.Sprintf("%s:6000", *hostPtr), conf)
    if err != nil {
        log.Println(err)
        return
    }

    n, err := conn.Write([]byte("HELO\n"))
    if err != nil {
        log.Println(n, err)
        return
    }

    //read in response from server
    buf := make([]byte, 100)
    n, err = conn.Read(buf)
    if err != nil {
        log.Println(n, err)
        return
    }
    println(string(buf[:n]))

    //reach out to server and request Pad
    data := Random_Req {
	Host: "client",
	UUID: fmt.Sprintf("%v", string(buf[:n])),
    }
    rapi_data, _ := json.Marshal(data)
    if err != nil {
	fmt.Println(err)
    }
    req, err := http.NewRequest("POST", "http://localhost:8090/api/otp", bytes.NewBuffer(rapi_data))
    req.Header.Set("Content-Type", "application/json; charset=UTF-8")
    client := &http.Client{}
    resp, error := client.Do(req)
    if error != nil {
	    panic(error)
    }
    var res map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&res)
    raw_pad := fmt.Sprintf("%v", res["Pad"])
    cipherText := pad_encrypt(*msgPtr, raw_pad)
    println(fmt.Sprintf("Ciphertext: %v\n", cipherText))

    n, err = conn.Write([]byte(fmt.Sprintf("%v\n", cipherText)))
    if err != nil {
        log.Println(n, err)
        return
    }

    conn.Close()

}
