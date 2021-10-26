package main

import (
	"fmt"
	"os/exec"
	"sync"

	mynet "github.com/RomanAvdeenko/utils/net"
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

func receivePong(pongChan <-chan Pong, done *[]Pong) {	
	for  pong :=range  pongChan{		
		//fmt.Println("received:", pong)
		if pong.Alive {
			*done = append(*done, pong)
			}			
		}
	}

func main() {
	hosts, _ := mynet.GetHosts("192.168.1.1/24")

	pingChan := make(chan string, concurrentMax)
	pongChan := make(chan Pong, len(hosts))
	done := []Pong{}
	wg := &sync.WaitGroup{}

	// Start workers
	for i := 0; i < concurrentMax; i++ {
		go func() {
			defer wg.Done()
			wg.Add(1)
			ping(pingChan, pongChan)
		}()
	}

	// Set job
	for _, ip := range hosts {
		pingChan <- ip
		//fmt.Println("sent: " + ip)
	}
	close(pingChan)

	// Start Receiver
	go receivePong(pongChan, &done)

	// Wait for all workers done and close pongChan
	wg.Wait()
	close(pongChan)

	// Get results	
	fmt.Println(done)
}
