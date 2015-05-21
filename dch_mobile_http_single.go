package main

import (
	"io/ioutil"
	"log"
	"net/http"
	//"os"
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
	//urlip       = "52.68.172.23:80"
	url = "r0401.dch.dlink.com"
	did = "12345678901234567890000000000003"
	//did     = "e0b26bcf62326764b9a2dc22142fe727"
	domain  = "http://r0401.dch.dlink.com"
	api_url = "/ws/api/getVersion?did="
	alive   = "Connection: keep-alive"
	request = domain + api_url + did
)

func closeHttp(resp *http.Response) {
	log.Println("closeHttp")
	defer resp.Body.Close()
}

func SendRoutine() {
	resp, err := http.Get(request)

	if err != nil {
		log.Println(err)
		//os.Exit(1)
		return
	}

	defer closeHttp(resp)
	readbytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println("Error while receive response:", err.Error())
	}

	log.Println("recieve buffer", string(readbytes))

}

func main() {
	go SendRoutine()

	for {
		time.Sleep(time.Second)
	}
	log.Println("Exit Program")
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
