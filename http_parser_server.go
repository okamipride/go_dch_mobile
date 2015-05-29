package main

import (
	"bufio"
	"fmt"
	"net"
)

var (
	res_ver_ok string = "HTTP/1.1 200 OK\r\nContent-type: text\r\nContent-Length: 2\r\n\r\nOK"
	resp_data  string = "\"status\":\"ok\"," +
		"\"errno\":\"\"," +
		"\"errmsg\":\"\"," +
		"\"version\":\"1.0.1\"," +
		"\"detail\":\"70b3852=2015-04-30 11:45:58 +80800\"\r\n"
	res_ver_chunk = "HTTP/1.1 200 OK\r\n" + "Server: Spaced/0.1\r\n" +
		"Transfer-Encoding: chunked\r\n" +
		"Content-Type: application/javascript; charset=utf-8\r\n" +
		"Connection: close\r\n\r\n" +
		"1\r\n" +
		"{\r\n" +
		"64\r\n" +
		resp_data +
		"1\r\n" +
		"}\r\n" +
		"0\r\n"
)

func main() {
	fmt.Println("Launching http parser server...")
	service := "localhost:8888"
	listener, err := net.Listen("tcp", service)
	if err != nil {
		fmt.Println("Listen Error : ", err)
		return
	}
	conn, _ := listener.Accept()
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("error", err)
		}

		fmt.Print("Message Received:", string(message))
		SendMessage(conn, res_ver_chunk, "12345678901234567890000000000001")
	}

}

func SendMessage(conn net.Conn, msg string, did string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println(did, " Error send request:", err.Error())
	}
}
