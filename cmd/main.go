package main

import (
	"fmt"
	"net"
	"sync"

	mynet "github.com/RomanAvdeenko/utils/net"
	"github.com/j-keck/arping"
)

const concurrentMax = 100

type Pong struct {
	Ip    string
	Alive bool
}

func ping(pingChan <-chan string, pongChan chan<- Pong) {
/*	Package arping is a native go library to ping a host per arp datagram, or query a host mac address
	The currently supported platforms are: Linux and BSD.
	The library requires raw socket access. So it must run as root, or with appropriate capabilities under linux: `sudo setcap cap_net_raw+ep <BIN>`. 
 */
	for ip := range pingChan {
		_,_, err :=  arping.Ping(net.ParseIP(ip) )		
		if err != nil {
			pongChan <- Pong{Ip: ip, Alive: false}
		} else {
			pongChan <- Pong{Ip: ip, Alive: true}
		}		
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
