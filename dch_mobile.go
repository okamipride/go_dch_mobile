package main

import (
	"fmt"
	"net"
	"os"
	"time"
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
	did         = "12345678901234567890123456789012"
	get_request = "GET /ws/api/getVersion?did=" + did + " HTTP/1.1\r\n"
	host        = "Host: 0401.dch.dlink.com\r\n"
	alive       = "Connection: keep-alive\r\n\r\n"
	get_msg     = get_request + host + alive
)

func SendRoutine() {

	tcpaddr, err := net.ResolveTCPAddr("tcp", urlip)
	if checkError(err, "App ResolveTCPAddr") {
		os.Exit(0)
	}

	conn, err := net.DialTCP("tcp", nil, tcpaddr)
	defer closeConn(conn)

	for i := 0; i < 3; i++ {

		//defer conn.Close() // close when leave the loop
		if checkError(err, "App DialTCP") {
			continue
		}

		fmt.Println("App TCP Connect ... ")

		SendMessage(conn, get_msg)

		fmt.Println("Agent Write = %s", get_request)

		echo := GetMessage(conn)
		if echo == "" {
			fmt.Println("echo empty")
			break
		}

		time.Sleep(time.Second * 2)

	}

	fmt.Println("SendRoutine Exist")
}

func main() {
	go SendRoutine()

	for {
		time.Sleep(time.Second)
	}
}

func closeConn(c *net.TCPConn) {
	fmt.Println("connection close")
	c.Close()
}

func checkError(err error, act string) bool {

	if err != nil {
		fmt.Println(act + " Error Occur!")
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		return true
	}
	//fmt.Println(act + " no Error")
	return false
}

func SendMessage(conn *net.TCPConn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		println("Error send request:", err.Error())
	} else {
		println("Request sent")
	}
}

func GetMessage(conn *net.TCPConn) string {
	buf_recever := make([]byte, RECV_BUF_LEN)
	n, err := conn.Read(buf_recever)
	if err != nil {
		println("Error while receive response:", err.Error())
		return ""
	}

	echodata := make([]byte, n)
	copy(echodata, buf_recever)

	return string(echodata)
}
