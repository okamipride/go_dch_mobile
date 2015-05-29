package main

import (
	"bufio"
	//"crypto/tls"
	"errors"
	"flag"
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
	//did     = "12345678901234567890000000000001"
	domain     = "http://r0402.dch.dlink.com"
	api_getVer = "/ws/api/getVersion?did="
	alive      = "Connection: keep-alive"
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
	conn, err := net.Dial("tcp", "rf_domain")
	log.Println(dev.usr_did, " SendRequest-----------")

}

func main() {

	log.Println("App Start to connect ... ")
	rf_url, num_dev, num_concurrence, delay_int := readArg()
	var my_delay time.Duration = time.Duration(delay_int) * time.Millisecond
	domain = rf_url

	//relay_url, num_dev, num_concurrence, my_delay := readNumDevice()

	total_request := float64(num_dev) // convert to float in order to count percentage
	//go AutoGC()

	errc := make(chan error, num_dev)
	for i := int64(1); i <= num_dev; i++ {
		device := Device{usr_did: genDid(i), usr_hash: genDid(i)}

		//go device.SendRoutine(errc, rf_url)
		go device.SendRequest(rf_url)

		if i%num_concurrence == 0 {
			time.Sleep(my_delay)
		}
	}
	//var total_response float64 = 0
	var total_response_ok float64 = 0
	var total_respones_err float64 = 0

	for v := range errc {
		if v == nil {
			total_response_ok = total_response_ok + 1
		} else {
			total_respones_err = total_respones_err + 1
			log.Println("error occur:", errc)
		}
		total_response := total_response_ok + total_respones_err
		if total_response >= total_request {
			closeCh(errc)
		}

	}
	//closeCh(errc)

	percentage := (total_response_ok / total_request) * 100

	okstr := strconv.FormatFloat(total_response_ok, 'f', 2, 64)
	requeststr := strconv.FormatFloat(total_request, 'f', 2, 64)
	pertstr := strconv.FormatFloat(percentage, 'f', 2, 64)
	log.Println("total_response_ok = ", okstr, "total_request = ", requeststr)
	log.Println("ok/request = ", pertstr, "%")

	for {

		time.Sleep(1000 * time.Millisecond)

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

func readArg() (string, int64, int64, int64) {

	serverPtr := flag.String("serv", "r0401.dch.dlink.com", "relay server address , no port included")
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

func readNumDevice() (string, int64, int64, time.Duration) {
	var relay_addr = ""
	var num_did int64 = 0
	var num_concur int64 = 0
	var delay time.Duration = 100 * time.Millisecond

	for {
		// Number of devices connect to relayd
		consolereader := bufio.NewReader(os.Stdin)

		log.Print("Enter Relay Server Address : ")
		input, err := consolereader.ReadString('\n')

		reg, _ := regexp.Compile("^[a-zA-Z0-9.]*$") // only alphanumeric and dot

		if err != nil {
			log.Println("ReadString error! Retry again = ", err)
			continue
		}

		relay_addr = reg.FindString(input)

		log.Print("Enter Number of Devices: ")
		input, err = consolereader.ReadString('\n')
		if err != nil {
			log.Println("ReadString error! Retry again = ", err)
			continue
		}
		reg, _ = regexp.Compile("^[1-9][0-9]*") // Remove special character only take digits
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
	return relay_addr, num_did, num_concur, delay
}
