package main

import (
	"fmt"
	"net"
	"os"
)

const (
	RECV_BUF_LEN     = 65535
	CONFIG_LINE_SIZE = 1024
)

type Device struct {
	procid   int
	usr_did  string
	usr_hash string
}

var (
	urlip       = "52.68.172.23:80"
	did         = "72dfc969ff9530f735f19691c655d07f"
	get_request = "GET /ws/api/getVersion?did=" + did + " HTTP/1.1\r\n"
	host        = "Host: 0301.dch.dlink.com\r\n"
	alive       = "Connection: keep-alive\r\n\r\n"
	get_msg     = get_request + host + alive
)

func main() {
	//establish connection
	fmt.Println("Agent Start to connect ... ")
	tcpaddr, err := net.ResolveTCPAddr("tcp",
		urlip)
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpaddr)
	checkError(err)
	fmt.Println("Agent TCP Connect ... ")

	go SendData(conn, get_msg)
	c := make(chan bool)
	go RecieveData(conn, c)
	<-c
	fmt.Println("Program is going to exit")
	//time.Sleep(5000)
	//os.Exit(0)
}

/*
	Sending data to Relay server
*/

func SendData(conn *net.TCPConn, msg string) {
	_, err := conn.Write([]byte(msg))
	fmt.Println("sendData write")
	checkError(err)
	fmt.Println("data send: %s", msg)
}

/*
	Recieve data from Relay server
*/

func RecieveData(conn *net.TCPConn, cs chan bool) {
	for {
		buf_recever := make([]byte, RECV_BUF_LEN)
		_, err := conn.Read(buf_recever)

		if err != nil {
			fmt.Println("Error while receive response:", err.Error())
			cs <- false
			break
		}

		fmt.Println("recieve data:%s", string(buf_recever))
	}
	cs <- true
}

/**
	Methods
**/

/**
	Utilities
**/

func checkError(err error) {

	if err != nil {
		fmt.Println("Error Occur!")
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
