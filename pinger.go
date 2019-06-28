package sonar

import (
	"net"
	"fmt"
	"math/rand"
	"time"
	"encoding/json"
	"errors"
)

type Pinger struct {
	domain        string
	lastPingID    int
	querySender   *net.UDPConn
	echoReceiver  *net.UDPConn
	generator     *rand.Rand
	remoteAddress *net.UDPAddr
	echoChan      chan Echo
}

type Echo struct {
	LocalAddress string
	Services []Service
}

func CreatePinger(groupAddress string, groupPort int, domain string) (*Pinger, error){
	listenAddress, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", groupAddress, groupPort - 1))
	if err != nil{
		return nil, err
	}
	remoteAddress, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", groupAddress, groupPort))
	if err != nil{
		return nil, err
	}
	receiver, err := net.ListenMulticastUDP("udp", nil, listenAddress)
	if err != nil{
		return nil, err
	}
	sender, err := net.ListenUDP("udp", nil)
	if err != nil{
		return nil, err
	}

	return &Pinger{domain:domain, querySender: sender, echoReceiver:receiver,
		generator:rand.New(rand.NewSource(time.Now().UnixNano())), remoteAddress:remoteAddress}, nil
}

func (pinger *Pinger)Query(timeout time.Duration) (echo Echo, err error){
	pinger.echoChan = make(chan Echo)
	go func() {
		var buffer = make([]byte, DefaultBufferSize)
		for {
			if err = pinger.sendPingMessage();err != nil{
				return
			}
			count, _, err := pinger.echoReceiver.ReadFromUDP(buffer)
			if err != nil{
				close(pinger.echoChan)
				return
			}
			var msg Message
			if err = json.Unmarshal(buffer[:count], &msg);err != nil{
				close(pinger.echoChan)
				return
			}
			if msg.ID == pinger.lastPingID{
				echo.Services = msg.Services
				echo.LocalAddress = msg.Requestor
				pinger.echoChan <- echo
				return
			}
		}
	}()
	timer := time.NewTimer(timeout)
	select {
	case <- timer.C:
		//timeout
		pinger.echoReceiver.Close()
		return echo, errors.New("query timeout")
	case echo, opened := <- pinger.echoChan:
		if !opened {
			//closed
			return echo, errors.New("get echo fail")
		}
		return echo, nil
	}
}

//async call
func (pinger *Pinger)TryQuery() error{
	pinger.echoChan = make(chan Echo)
	if err := pinger.sendPingMessage(); err != nil{
		return err
	}
	go pinger.asyncRoutine()
	return nil
}

func (pinger *Pinger)asyncRoutine(){
	var buffer = make([]byte, DefaultBufferSize)
	for {
		if err := pinger.sendPingMessage();err != nil{
			close(pinger.echoChan)
			return
		}
		count, _, err := pinger.echoReceiver.ReadFromUDP(buffer)
		if err != nil{
			close(pinger.echoChan)
			return
		}
		var msg Message
		if err = json.Unmarshal(buffer[:count], &msg);err != nil{
			continue
		}
		if msg.ID == pinger.lastPingID{
			pinger.echoChan <- Echo{msg.Requestor, msg.Services}
			close(pinger.echoChan)
			break
		}
	}
}

//async call
func (pinger *Pinger)GetResult(timeout time.Duration) (echo Echo, err error){

	var timer = time.NewTimer(timeout)
	select{
	case echo, opened := <- pinger.echoChan:
		if !opened {
			//closed
			return echo, errors.New("get echo fail")
		}
		return echo, nil
	case <- timer.C:
		//timeout
		return echo, errors.New("get echo timeout")
	}
}

func (pinger *Pinger)sendPingMessage() (error){
	const (
		MaxID = 0xffff
	)
	var pingID = pinger.generator.Intn(MaxID)
	var ping = Message{Type:SonarPing, Domain: pinger.domain, ID:pingID}
	data, err := json.Marshal(ping)
	if err != nil{
		return err
	}
	_, err = pinger.querySender.WriteToUDP(data, pinger.remoteAddress)
	pinger.lastPingID = pingID
	return err
}