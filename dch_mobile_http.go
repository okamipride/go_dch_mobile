package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
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
	//urlip       = "52.68.172.23:80"
	url     = "r0401.dch.dlink.com"
	did     = "12345678901234567890000000000001"
	domain  = "http://r0401.dch.dlink.com"
	api_url = "/ws/api/getVersion?did="
	alive   = "Connection: keep-alive"
	//request = domain + api_url + did
)

func closeHttp(resp *http.Response) {
	log.Println("closeHttp")
	defer resp.Body.Close()
}

func (dev *Device) SendRoutine() {
	resp, err := http.Get(domain + api_url + dev.usr_did)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer closeHttp(resp)
	readbytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println("Error while receive response:", err.Error())
	}

	log.Println("recieve buffer", string(readbytes))
}

func main() {
	log.Println("Agent Start to connect ... ")
	num_dev, num_concurrence, my_delay := readNumDevice()

	//go AutoGC()

	for i := int64(1); i <= num_dev; i++ {
		//log.Println("delay time")
		device := Device{usr_did: genDid(i), usr_hash: genDid(i)}
		go device.SendRoutine()
		if num_dev%num_concurrence == 0 {
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

func closeConn(c *http.Response) {
	log.Println("connection close")
	c.Body.Close()
}

func checkError(err error, act string) bool {

	if err != nil {
		log.Println(act + " Error Occur!")
		log.Println("Fatal error: %s", err.Error())
		return true
	}
	return false
}

func readNumDevice() (int64, int64, time.Duration) {
	var num_did int64 = 0
	var num_concur int64 = 0
	var delay time.Duration = 100 * time.Millisecond

	for {
		// Number of devices connect to relayd
		consolereader := bufio.NewReader(os.Stdin)
		log.Print("Enter Number of Devices: ")
		input, err := consolereader.ReadString('\n')
		if err != nil {
			log.Println("ReadString error! Retry again = ", err)
			continue
		}
		reg, _ := regexp.Compile("^[1-9][0-9]*") // Remove special character only take digits
		num := reg.FindString(input)

		if err != nil {
			log.Println("ReadString error! Please enter digits. error = ", err)
			continue
		}

		log.Println(string(num))

		num_devices, err := strconv.ParseInt(string(num), 0, 64)

		if err != nil {
			fmt.Println(err)
			fmt.Println("Convert Number failed! Please re-enter")
			continue
		}
		num_did = num_devices

		// Number of devices connect to relayd continousely without delay
		log.Print("Enter Number of Concurrent Connect: ")
		input, err = consolereader.ReadString('\n')

		if err != nil {
			log.Println("ReadString error! Use number of devices. error = ", err)
			num_concur = num_did
		}

		concur := reg.FindString(input)
		num_concur, err = strconv.ParseInt(string(concur), 0, 64)
		if err != nil {
			log.Println("ReadString error! Use number of devices. error = ", err)
			num_concur = num_did
		}

		// Number of devices connect to relayd continousely without delay
		log.Print("Enter delay ms: ")
		input, err = consolereader.ReadString('\n')

		if err != nil {
			log.Println("ReadString error! Use 100ms. error = ", err)
			delay = 100 * time.Millisecond
		}

		delayStr := reg.FindString(input)
		ms, err := strconv.ParseInt(string(delayStr), 0, 64)

		if err != nil {
			log.Println("ReadString error! Use 100ms,  error = ", err)
			delay = 100 * time.Millisecond
		}

		delay = time.Duration(ms) * time.Millisecond

		break
	}
	return num_did, num_concur, delay
}
