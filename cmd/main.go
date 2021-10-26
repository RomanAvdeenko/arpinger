package main

import (
	"log"
	"net"
	"sync"
	"time"

	mynet "github.com/RomanAvdeenko/utils/net"
	"github.com/j-keck/arping"
)

const concurrentMax = 100

type Pong struct {
	ip    string
	alive bool
	respTime time.Duration
	macAddr net.HardwareAddr
}

func ping(pingChan <-chan string, pongChan chan<- Pong, goRoutineNum int) {
/*	Package arping is a native go library to ping a host per arp datagram, or query a host mac address
	The currently supported platforms are: Linux and BSD.
	The library requires raw socket access. So it must run as root, or with appropriate capabilities under linux: `sudo setcap cap_net_raw+ep <BIN>`. 
 */
 	log.Println("Start goroutine", goRoutineNum)
	for ip := range pingChan {
		var isAlive bool
		macAddr,duration, err :=  arping.Ping(net.ParseIP(ip) )
		if   err != nil {
			isAlive = false
		} else {
			isAlive = true
		}	
		pongChan <- Pong{ip: ip, alive: isAlive, respTime: duration, macAddr: macAddr}	
	}
	log.Println("Stop goroutine", goRoutineNum)
}

func receivePong(pongChan <-chan Pong, done *[]Pong) {	
	log.Println("-->Start recievePong")
	for  pong :=range  pongChan{		
		//fmt.Println("received:", pong)
		if pong.alive {
			*done = append(*done, pong)
			}			
		}
		log.Println("->Stop recievePong")
	}

func main() {
	log.Println("#Start")
	hosts, _ := mynet.GetHosts("192.168.1.1/24")

	pingChan := make(chan string, concurrentMax)
	pongChan := make(chan Pong, len(hosts))
	done := []Pong{}
	wg := &sync.WaitGroup{}

	// Start workers
	for i := 0; i < concurrentMax; i++ {
		i:=i
		go func() {
			defer wg.Done()			
			wg.Add(1)
			ping(pingChan, pongChan, i)
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
	log.Println("Result:", done)	
	log.Println("#Stop")
}
