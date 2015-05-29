package main

import (
	"bufio"
	//"crypto/tls"
	//"errors"
	"flag"
	"fmt"
	//"io/ioutil"
	"log"
	"net/http"
	//"os"
	//"regexp"
	//"strconv"
	"net"
	"time"
)

const (
	RECV_BUF_LEN     = 65535
	CONFIG_LINE_SIZE = 1024
)

var did_prefix = "1234567890" + "1234567890"

type Device struct {
	usr_did  string
	usr_hash string
}

var (
	did         = "12345678901234567890000000000001"
	domain      = "http://r0402.dch.dlink.com"
	api_getVer  = "/ws/api/getVersion?did="
	alive       = "Connection: keep-alive"
	get_request = "GET /ws/api/getVersion?did=" + did + " HTTP/1.1\r\n\r\n"
)

func closeHttp(resp *http.Response) {
	log.Println("closeHttp")
	defer resp.Body.Close()
}

func closeCh(ch chan error) {
	log.Println("closeChannel")
	close(ch)
}

func (dev *Device) SendRequest(rf_domain string) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", rf_domain)
	if err != nil {
		log.Println("error", err, " url=", rf_domain)
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpaddr)
	defer closeConn(conn)
	if err != nil {
		fmt.Println("error=", err)
	}
	SendMessage(conn, get_request)

	status, err := bufio.NewReader(conn).ReadString('\n')
	fmt.Println("resp = ", status)

}

func main() {

	log.Println("App Start to connect ... ")
	rf_url := readArg()
	domain = rf_url + ":8888"
	//go AutoGC()
	device := Device{usr_did: genDid(10), usr_hash: genDid(10)}
	device.SendRequest(domain)

	for {
		time.Sleep(1000 * time.Millisecond)
	}
	log.Println("Exit Program")
}

func checkError(err error, act string) bool {

	if err != nil {
		log.Println(act + " Error Occur!")
		log.Println("Fatal error: %s", err.Error())
		return true
	}
	return false
}
func closeConn(c *net.TCPConn) {
	log.Println("connection close")
	c.Close()
}

func SendMessage(conn *net.TCPConn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println(" Error send request:", err.Error())
	}
}

func readArg() string {

	serverPtr := flag.String("serv", "r0401.dch.dlink.com", "relay server address , no port included")

	var svar string
	flag.StringVar(&svar, "svar", "bar", "command line arguments")
	flag.Parse()
	fmt.Println("server:", *serverPtr)

	fmt.Println("tail:", flag.Args())

	return *serverPtr
}

func genDid(num int64) string {
	return did_prefix + fmt.Sprintf("%012x", num)
}
