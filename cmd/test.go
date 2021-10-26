package main

import (
	"fmt"
	mynet "github.com/RomanAvdeenko/utils/net"
	"os/exec"
)

const concurrentMax = 100

type Pong struct {
	Ip    string
	Alive bool
}

func ping(pingChan <-chan string, pongChan chan<- Pong) {
	for ip := range pingChan {
		_, err := exec.Command("ping", "-c1", "-t1", ip).Output()
		var alive bool
		if err != nil {
			alive = false
		} else {
			alive = true
		}
		pongChan <- Pong{Ip: ip, Alive: alive}
	}
}

func receivePong(pongNum int, pongChan <-chan Pong, doneChan chan<- []Pong) {
	var alives []Pong
	for i := 0; i < pongNum; i++ {
		pong := <-pongChan
		//fmt.Println("received:", pong)
		if pong.Alive {
			alives = append(alives, pong)
		}
	}
	doneChan <- alives
}

func main() {
	hosts, _ := mynet.GetHosts("192.168.1.1/24")

	pingChan := make(chan string, concurrentMax)
	pongChan := make(chan Pong, len(hosts))
	doneChan := make(chan []Pong)

	// Start workers
	for i := 0; i < concurrentMax; i++ {
		go ping(pingChan, pongChan)
	}

	// Set job
	for _, ip := range hosts {
		pingChan <- ip
		//fmt.Println("sent: " + ip)
	}
	close(pingChan)

	go receivePong(len(hosts), pongChan, doneChan)

	alives := <-doneChan
	fmt.Println(alives)
}
