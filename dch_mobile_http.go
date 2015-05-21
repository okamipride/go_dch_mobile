package main

import (
	"bufio"
	//"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	//"time"
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
	url = "r0401.dch.dlink.com"
	//did     = "12345678901234567890000000000001"
	domain  = "http://r0401.dch.dlink.com"
	api_url = "/ws/api/getVersion?did="
	alive   = "Connection: keep-alive"
	//request = domain + api_url + did
)

func closeHttp(resp *http.Response) {
	log.Println("closeHttp")
	defer resp.Body.Close()
}

//func (dev *Device) SendRoutine(httpclient *http.Client, errc chan error) {
func (dev *Device) SendRoutine(errc chan error) {
	//resp, err := client.Get(domain + api_url + dev.usr_did)
	//resp, err := httpclient.Get(domain + api_url + dev.usr_did)
	resp, err := http.Get(domain + api_url + dev.usr_did)
	defer closeHttp(resp)

	if err != nil {
		//log.Println("get error :", err)
		errc <- err
		close(errc)
		return
	}

	readbytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		errc <- err
		close(errc)
		return
	}

	if resp.StatusCode != 200 {
		log.Println("Response not ok")
		errc <- err
		close(errc)
		return
	}

	log.Println("recieve buffer", string(readbytes))
	errc <- nil
	close(errc)

}

func main() {
	var total_request int64 = 0
	var total_response_ok int64 = 0

	num_dev := readNumDevice()
	total_request = num_dev

	// for keepalive settings
	//tr := &http.Transport{
	//	DisableKeepAlives: false,
	//TLSClientConfig:    &tls.Config{RootCAs: nil},
	//DisableCompression: true,
	//}

	errc := make(chan error)

	for i := int64(1); i <= num_dev; i++ {
		//client := &http.Client{Transport: tr}
		device := Device{usr_did: genDid(i), usr_hash: genDid(i)}
		//go device.SendRoutine(client, errc)
		go device.SendRoutine(errc)
	}

	for v := range errc {
		if v == nil {
			total_response_ok = total_response_ok + 1
			log.Println("get ok", strconv.FormatInt(total_response_ok, 10))
		} else {
			log.Println("error occur:", errc)
		}
	}

	percentage := (total_response_ok / total_request) * 100

	okstr := strconv.FormatInt(total_response_ok, 10)
	requeststr := strconv.FormatInt(total_request, 10)
	pertstr := strconv.FormatInt(percentage, 10)
	log.Println("total_response_ok = ", okstr, "total_request = ", requeststr)
	log.Println("ok/request = ", pertstr, "%")

	//close(errc)

	/*
		for {
			time.Sleep(time.Second)
		}
	*/
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

func readNumDevice() int64 {
	var num_did int64 = 0

	for {
		consolereader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter Number of Devices: ")
		input, err := consolereader.ReadString('\n') // this will prompt the user for input

		if err != nil {
			fmt.Println(err)
			fmt.Println("ReadString error! Retry again")
			continue
		}
		reg, _ := regexp.Compile("^[1-9][0-9]*") // Remove special character only take digits
		num := reg.FindString(input)

		if err != nil {
			fmt.Println(err)
			fmt.Println("ReadString error! Please enter digits")
			continue
		}

		fmt.Println(string(num))

		num_devices, err := strconv.ParseInt(string(num), 0, 64)

		if err != nil {
			fmt.Println(err)
			fmt.Println("Convert Number failed! Please re-enter")
			continue
		}
		num_did = num_devices
		break
		//return num_devices
	}

	return num_did
}
