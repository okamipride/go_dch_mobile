package main

import (
	//"bufio"
	"flag"
	"fmt"
	//"io/ioutil"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	RECV_HEADER_CAP        = 1024 * 2
	RECV_BUF_CAP           = 1024 * 10
	RECV_BODY_CAP          = RECV_BUF_CAP
	RFE_REQUEST_TIMEOUT    = 5000
	RELAY_RESPONSE_TIMEOUT = 15000
)

var did_prefix = "1234567890" + "1234567890"

type Device struct {
	usr_did  string
	usr_hash string
}

var (
	did            = "12345678901234567890000000000001"
	get_ver_prefix = "GET /ws/api/getVersion?did="
	http_1_1       = " HTTP/1.1\r\n"
	host           = "Host: 0301.dch.dlink.com\r\n"
	alive          = "Connection: keep-alive\r\n\r\n"
)

func closeHttp(resp *http.Response) {
	//log.Print(" closeHttp ")
	defer resp.Body.Close()
}

func (dev *Device) SendRoutine(url string, loginfo bool) {
	get_ver_msg := get_ver_prefix + dev.usr_hash + http_1_1 + host + alive
	tcpaddr, err := net.ResolveTCPAddr("tcp", url)
	if err != nil {
		log.Println("error", err, " url=", url)
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpaddr)
	defer closeConn(conn)

	if err != nil {
		log.Println("connect error", err, "url = ", url)
		return // retry escape
	}

	if loginfo {
		fmt.Printf(" %s %s ", dev.usr_did[27:32], "Connected")
	}
	conn.Write([]byte(get_ver_msg))

	/*
		rep, er_2 := GetReqRepEx(conn, RELAY_RESPONSE_TIMEOUT, RELAY_RESPONSE_TIMEOUT, nil)
		if er_2 != nil {
			log.Println("response error", er_2)
		}

		fmt.Println("response=", rep)
	*/
}

func main() {
	log.Println("App Start to connect ... ")
	rf_url, num_dev, num_concurrence, delay_int := readArg()
	var my_delay time.Duration = time.Duration(delay_int) * time.Millisecond
	rf_url = rf_url + ":80"
	debuglog := true
	//go AutoGC()

	for i := int64(1); i <= num_dev; i++ {
		//log.Println("delay time")
		device := Device{usr_did: genDid(i), usr_hash: genDid(i)}
		go device.SendRoutine(rf_url, debuglog)
		if i%num_concurrence == 0 {
			time.Sleep(my_delay)
		}
	}

	for {
		time.Sleep(time.Second)
	}

	log.Println("Exit Program")
}

func genDid(num int64) string {
	return did_prefix + fmt.Sprintf("%012x", num)
}

func closeConn(c *net.TCPConn) {
	log.Println("connection close")
	c.Close()
}

func checkError(err error, act string) bool {

	if err != nil {
		log.Println(act + " Error Occur!")
		log.Println("Fatal error: %s", err.Error())
		return true
	}
	return false
}

func readArg() (string, int64, int64, int64) {

	serverPtr := flag.String("serv", "r0402.dch.dlink.com", "relay server address , no port included")
	//serverPtr := flag.String("serv", "172.31.4.183:80", "relay server address , no port included")
	devPtr := flag.Int64("dev", 1, "number of devices want to connect to relay server")
	concurPtr := flag.Int64("concur", 1, "Concurrent send without delay")
	delayPtr := flag.Int64("delay", 10, "Delay between concurrent send")

	var svar string
	flag.StringVar(&svar, "svar", "bar", "command line arguments")
	flag.Parse()
	fmt.Println("server:", *serverPtr)
	fmt.Println("dev:", *devPtr)
	fmt.Println("concurrent:", *concurPtr)
	fmt.Println("delay:", *delayPtr)
	fmt.Println("tail:", flag.Args())

	return *serverPtr, *devPtr, *concurPtr, *delayPtr
}

/****
	Parser Portion
*****/

func SetTImeout(c *net.TCPConn, timeout int) {
	if timeout == 0 {
		c.SetReadDeadline(time.Time{})
	} else {
		c.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(timeout)))
	}
}

func (dev *Device) GetResponsec(c *net.TCPConn, wait_timeout, read_timeout int, tsheader string) string {
	header := make([]byte, 0, RECV_HEADER_CAP)
	headbuf := make([]byte, 1, RECV_BUF_CAP)
	//body := make([]byte, 0, RECV_BODY_CAP)

	for {
		n, err := c.Read(headbuf)
		SetTImeout(c, read_timeout)
		if err != nil { // Read Error
			if err == io.EOF {
				fmt.Println("EOF rechea")
			}
			break
		}

		header = append(header, headbuf[:n]...)
		if strings.Index(string(header), "\r\n\r\n") > 0 {
			break
		}
	}

	return string(header)
}

func GetReqRepEx(c *net.TCPConn, wait_timeout, read_timeout int, tsheader string) (reqrep string, er error) {
	var chunked bool
	var contentlength int
	var newheader string
	header := make([]byte, 0)
	headbuf := make([]byte, 1)
	body := make([]byte, 0)

	defer func() {
		SetTImeout(c, wait_timeout)
		if er == nil {
			reqrep = newheader + string(body)
		} else {
			reqrep = ""
		}
		return
	}()

	//===================================  設定SOCKET連上後等待第一個\n的時間  ==============================
	SetTImeout(c, wait_timeout)
	//===================================  此部分讀完 HEADER ============================================

	for er == nil {
		n, err := c.Read(headbuf)
		SetTImeout(c, read_timeout)
		er = err
		header = append(header, headbuf[:n]...)
		if strings.Index(string(header), "\r\n\r\n") > 0 {
			break
		}
	}
	if er != nil {
		return
	}
	//===================================  塞入TimeStamp到Header尾端 ====================================
	newheader = string(header[0:len(header)-2]) + fmt.Sprintf("%s: %0.5f\r\n\r\n", tsheader, (float64(time.Now().UnixNano())/1000000000))

	//===================================  此部份決定Body該如何收  ======================================
	have_content_length := strings.Index(newheader, "Content-Length:")
	have_chunked := strings.Index(newheader, "chunked")
	if have_content_length > 0 {
		h := []byte(header)[have_content_length+16:]
		stopindex := strings.Index(string(h), "\r\n")
		contentlength, er = strconv.Atoi(string([]byte(h)[0:stopindex]))
		if er != nil {
			return
		}
		chunked = false
		log.Println("[Header]Content-Length: ", contentlength)
	} else if have_chunked > 0 {
		chunked = true
		log.Println("[Header]Transfer-Encoding: chunked")
	}
	//===================================  此部分開始讀 BODY ============================================
	var n int
	if chunked {
		tmp := make([]byte, 1024*32)
		c.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(50)))
		n, er = c.Read(tmp)
		body = append(body, tmp[:n]...)
		if er != nil {
			return
		}
	} else if contentlength > 0 {
		content := make([]byte, contentlength)
		n, er = io.ReadFull(c, content)
		body = append(body, content[:contentlength]...)
		log.Printf("Content:%d,Data:%d,Err:%s", contentlength, n, er)
		if er != nil {
			return
		}
	}
	//===================================  此部分回傳Response ==========================================
	return
}
