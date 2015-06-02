package main

import (
	//"bufio"
	"flag"
	"fmt"
	//"io/ioutil"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	//"strings"
	//"encoding/hex"
	"time"
)

var did_base int64 = 0
var state_log = false
var msg_delay_msg = 1000

const (
	RECV_HEADER_CAP        = 1024 * 10
	RECV_BUF_CAP           = 1024 * 10
	RECV_BODY_CAP          = RECV_BUF_CAP
	RFE_REQUEST_TIMEOUT    = 5 * 1000
	RELAY_RESPONSE_TIMEOUT = 15 * 1000
)

var did_prefix = "1234567890" + "1234567890"

type Device struct {
	usr_did  string
	usr_hash string
}

var (
	did            = "6b326abfca91fe5c9c7fa28612910a6c"
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
	get_ver_msg := get_ver_prefix + dev.usr_did + http_1_1 + host + alive
	//get_ver_msg := get_ver_prefix + did + http_1_1 + host + alive
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
	conn.SetKeepAlive(true)
	if loginfo {
		fmt.Printf(" %s %s ", dev.usr_did[27:32], "Connected")
	}

	for {
		conn.Write([]byte(get_ver_msg))

		//head_str, body_str, exit_state := dev.GetRespByReg(conn, RELAY_RESPONSE_TIMEOUT, RELAY_RESPONSE_TIMEOUT)
		_, _, exit_state := dev.GetRespByReg(conn, RELAY_RESPONSE_TIMEOUT, RELAY_RESPONSE_TIMEOUT)
		//fmt.Println("response hearder =", head_str)
		//fmt.Println("response hearder =", body_str)
		if exit_state == READ_ERR {
			fmt.Println("READ ERROR")
		} else {
			fmt.Println("READ OK")
		}
		time.Sleep(time.Duration(msg_delay_msg) * time.Millisecond)
	}

}

func main() {
	log.Println("App Start to connect ... ")
	rf_url, num_dev, num_concurrence, delay_int := readArg()
	var my_delay time.Duration = time.Duration(delay_int) * time.Millisecond
	rf_url = rf_url + ":80"
	debuglog := true
	//go AutoGC()
	for i := int64(1); i <= num_dev; i++ {
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

const (
	HEADER_READING = iota
	CHUNK_READING
	TEXT_READING
	READ_END
	READ_ERR
)

const (
	CHUNKED_READ_STOP = iota
	CHUNKED_READ_COUNT
	CHUNKED_READ_COUNT_R
	CHUNKED_READ_COUNT_END
	CHUNKED_READ_TEXT
	CHUNKED_READ_TEXT_R
	CHUNKED_READ_TEXT_END
	CHUNKED_READ_END
)

func (dev *Device) GetRespByReg(c *net.TCPConn, wait_timeout int, read_timeout int) (string, string, int) {
	header := make([]byte, 0, RECV_HEADER_CAP)
	readbuf := make([]byte, 1, RECV_BUF_CAP)
	body := make([]byte, 0, RECV_BODY_CAP)
	state := HEADER_READING            // Reading state
	chunked_state := CHUNKED_READ_STOP // State used in Reading Chunked data
	var content_length int64           // use in Content-Length reading (Text format)
	digitTemp := make([]byte, 0, 1024) // use to store chunked length
	var nextLength int64 = 0           // use to store  chunked length

	for state != READ_END && state != READ_ERR {

		switch state {
		case HEADER_READING:
			n, err := c.Read(readbuf)
			SetTImeout(c, read_timeout)
			if err != nil { // Read Error
				if err == io.EOF {
					fmt.Println("EOF reach")
					state = READ_ERR
					break
				}
				fmt.Println("Read Error ", err)
				state = READ_ERR
				break
			}

			header = append(header, readbuf[:n]...)
			find := findheader(string(header))

			if find != "" { //find header end , determine is text or chunched
				err, content_length = isText(find)
				if err != nil {
					chuncked := isChunked(string(header))
					if chuncked {
						state = CHUNK_READING
						chunked_state = CHUNKED_READ_COUNT
						if state_log {
							fmt.Println("****CHUNKED_READ_COUNT")
						}
						break
					} else {
						state = READ_ERR
						fmt.Println("error = ", err)
					}

				}
				if state_log {
					fmt.Println("cont_len:", content_length)
				}
				state = TEXT_READING
				if state_log {
					fmt.Println("****TEXT_READING")
				}
			}
			/*
			   READ Text DATA
			*/
		case TEXT_READING:
			//fmt.Println("TEXT_READING")
			content := make([]byte, content_length)
			n, er := io.ReadFull(c, content)
			body = append(body, content[:content_length]...)
			if state_log {
				log.Printf("Content:%d,Data:%d,Err:%s", content_length, n, er)
				fmt.Println("Data =", string(body))
			}
			state = READ_END
			if state_log {
				fmt.Println("****READ_END")
			}
			/*
			   READ CHUNKED DATA
			*/
		case CHUNK_READING:
			//fmt.Printf(" CHUNK_READING ")
			n, err := c.Read(readbuf)
			SetTImeout(c, read_timeout)
			if err != nil { // Read Error
				if err == io.EOF {
					fmt.Println("EOF reach")
					state = READ_END
				}
				fmt.Println("Read Error ", err)
				state = READ_ERR
			}

			body = append(body, readbuf[:n]...)

			switch chunked_state {
			case CHUNKED_READ_COUNT:
				//fmt.Println("CHUNKED_READ_COUNT", readbuf, string(readbuf))
				isR := isCartByte(readbuf)
				if isR {
					chunked_state = CHUNKED_READ_COUNT_R
					if state_log {
						fmt.Println("==== Next  State = CHUNKED_READ_COUNT_R")
					}
					break
				}
				isDigit, _ := isDigitBytes(readbuf)
				if !isDigit {
					state = READ_ERR
					break
				}
				digitTemp = append(digitTemp, readbuf[:n]...) // save read-digit for length
				if state_log {
					fmt.Println("CHUNKED_READ_COUNT digitTemp", string(digitTemp), "length=", len(string(digitTemp)))
				}
			case CHUNKED_READ_COUNT_R:
				isN := isNewlineByte(readbuf)
				if !isN {
					state = READ_ERR
					break
				}
				chunked_state = CHUNKED_READ_COUNT_END
				if state_log {
					fmt.Println("==== Next  State = CHUNKED_READ_COUNT_END")
				}
			case CHUNKED_READ_COUNT_END:
				err, nextLength = lineHexGetInt(string(digitTemp))
				if state_log {
					fmt.Println("CHUNKED_READ_COUNT_END nextLength=", nextLength, "buf=", readbuf, string(readbuf), "string length", string(digitTemp))
				}
				if err != nil {
					fmt.Println("CHUNKED_READ_COUNT_END error", err)
					state = READ_ERR
					break
				}
				// No error and read length = 0
				if nextLength == 0 { // End of Read Chunk
					if isCartByte(readbuf) {
						chunked_state = CHUNKED_READ_END // Prepare to READ END
						if state_log {
							fmt.Println("==== Next  State = CHUNKED_READ_END_R")
						}
					}
					break
				}
				digitTemp = digitTemp[:0] // reset temp
				chunked_state = CHUNKED_READ_TEXT
				if state_log {
					fmt.Println("==== Next  State = CHUNKED_READ_TEXT")
				}

			case CHUNKED_READ_TEXT:
				nextLength = nextLength - 1
				if nextLength <= 0 {
					if isCartByte(readbuf) {
						chunked_state = CHUNKED_READ_TEXT_R
						if state_log {
							fmt.Println("==== Next  State = CHUNKED_READ_TEXT_R")
						}
					} else {
						state = READ_ERR
						if state_log {
							fmt.Println("==== Next  State = READ_ERR")
						}
					}
				}

			case CHUNKED_READ_TEXT_R:
				//fmt.Println("CHUNKED_READ_COUNT_R readbuf", readbuf)
				isNewline := isNewlineByte(readbuf)
				if !isNewline {
					state = READ_ERR
					fmt.Println("==== Next  State = READ_ERR")
					break
				}
				// Becasue we had already read a count byte from buffer
				chunked_state = CHUNKED_READ_TEXT_END
				if state_log {
					fmt.Println("==== Next  State = CHUNKED_READ_TEXT_END")
				}
			case CHUNKED_READ_TEXT_END:
				digitTemp = append(digitTemp, readbuf[:n]...)
				chunked_state = CHUNKED_READ_COUNT
				if state_log {
					fmt.Println("==== Next  State = CHUNKED_READ_COUNT")
				}

			case CHUNKED_READ_END:
				if isNewlineByte(readbuf) { // new line , the end the chunck
					chunked_state = CHUNKED_READ_STOP
					state = READ_END
					if state_log {
						fmt.Println("**** Next  State = READ_END")
					}
					break
				} else {
					chunked_state = CHUNKED_READ_STOP
					state = READ_ERR
					fmt.Println("error : no /n in the chunck end")
					break
				}

			case CHUNKED_READ_STOP:
				if state_log {
					fmt.Println("CHUNKED_READ_STOP")
				}
				state = READ_END
				if state_log {
					fmt.Println("==== Next  State = READ_END")
				}
				break
			}
			//n, err := c.Read(body)
		} // switch 1 end

	} // for end
	if state_log {
		fmt.Println("Exist READING state =", state)
	}
	return string(header), string(body), state
}

func findheader(str string) string {
	reg, _ := regexp.Compile("((?s).*?)\\r\\n\\r\\n")
	find := reg.FindString(str)
	//fmt.Println("findLen = ", len(find))
	return find
}

func isText(str string) (error, int64) {
	reg, err := regexp.Compile("[C|c][O|o][N|n][T|t][E|e][N|n][T|t]-[L|l][E|e][N|n][G|g][T|t][H|h]:.*?\\r\\n") // Content-Length
	if err != nil {
		return err, 0
	}
	find := reg.FindString(str)

	if find != "" {
		err, num := lineGetInt(find)
		if err != nil {
			return err, 0
		}
		//fmt.Println("isText = ", num)
		return nil, num
	}
	return errors.New("not found"), 0
}

func isChunked(header string) bool {
	reg, err := regexp.Compile("[C|c][H|h][U|u][N|n][K|k][E|e][D|d]") //Chuncked
	if err != nil {
		return false
	}
	find := reg.FindString(header)

	if find == "" {
		return false
	}

	return true
}

func isEndLine(str string) bool {
	reg, _ := regexp.Compile("((?s).*?)\\r\\n") // any thing has and end
	find := reg.FindString(str)
	if find != "" {
		return true
	}
	return false
}

func isCartByte(rbyte []byte) bool {
	reg, _ := regexp.Compile("\\r")
	find := reg.FindString(string(rbyte))
	if find != "" {
		//fmt.Println("isCartByte", rbyte)
		return true
	}
	return false
}

func isNewlineByte(nbyte []byte) bool {
	reg, _ := regexp.Compile("\\n")
	find := reg.FindString(string(nbyte))
	if find != "" {
		//fmt.Println("isNewlineByte", nbyte)
		return true
	}
	return false
}

func isDigitBytes(dgbyte []byte) (bool, string) {
	num_reg, _ := regexp.Compile("[0-9]+")
	find := num_reg.FindString(string(dgbyte))
	if find == "" { // not found
		return false, ""
	}
	return true, find
}

func lineGetInt(str string) (error, int64) {
	if str != "" {
		num_reg, err := regexp.Compile("[0-9]+")
		if err != nil {
			return err, 0
		}
		findnum := num_reg.FindString(str)
		if findnum == "" {
			return errors.New("digit not found"), 0
		}
		//num, err := strconv.Atoi(findnum)
		num, err := strconv.ParseInt(findnum, 0, 64)
		if err != nil {
			return err, 0
		}
		return nil, num
	}
	return errors.New("digit not found"), 0
}

func lineHexGetInt(str string) (error, int64) {
	//fmt.Println("lineHexGetInt str=", str)
	if str != "" {
		num_reg, err := regexp.Compile("[0-9a-fA-F]+")
		if err != nil {
			fmt.Println("lineHexGetInt reg error")
			return err, 0
		}
		findnum := num_reg.FindString(str)
		if findnum == "" {
			return errors.New("Hex String not found"), 0
		}
		//fmt.Println("lineHexGetInt find=", findnum)
		return nil, Conv16To10base(findnum)
	}
	return errors.New("digit not found"), 0
}

func Conv16To10base(num string) int64 {
	length := len(num)
	var single string
	var base int64 = 1
	var sum int64 = 0
	for i := int64(length); i > 0; i-- {
		single = num[i-1 : i]
		//fmt.Println(" char = ", single)
		switch single {
		case "a", "A":
			sum = sum + 10*base
		case "b", "B":
			sum = sum + 11*base
		case "c", "C":
			sum = sum + 12*base
		case "d", "D":
			sum = sum + 13*base
		case "e", "E":
			sum = sum + 14*base
		case "f", "F":
			sum = sum + 15*base
		default:
			digit, _ := strconv.ParseInt(single, 0, 64)
			sum = sum + digit*base
		}
		//fmt.Println("Conv16To10base sum ", sum)
		base = base * 16
	}
	return sum
}
