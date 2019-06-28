package sonar

import (
	"net"
	"fmt"
	"log"
	"encoding/json"
)

type Listener struct {
	OutputLog     bool
	services      []Service
	domain        string
	queryReceiver *net.UDPConn
	echoSender    *net.UDPConn
	pingerAddress *net.UDPAddr
	exitChan      chan bool
}

func CreateListener(groupAddress string, groupPort int, domain string, listenInterface *net.Interface) (*Listener, error){
	listenAddress, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", groupAddress, groupPort))
	if err != nil{
		return nil, err
	}
	receiveAddress, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", groupAddress, groupPort - 1))
	if err != nil{
		return nil, err
	}
	receiver, err := net.ListenMulticastUDP("udp", listenInterface, listenAddress)
	if err != nil{
		return nil, err
	}
	if err = receiver.SetReadBuffer(DefaultBufferSize);err != nil{
		return nil, err
	}
	sender, err := net.ListenUDP("udp", nil)
	if err != nil{
		return nil, err
	}

	return &Listener{domain:domain, queryReceiver: receiver, echoSender:sender, pingerAddress:receiveAddress,
		exitChan: make(chan bool)}, nil
}

func (listener *Listener)AddService(serviceType, protocol, address string, port int) error{
	listener.services = append(listener.services, Service{serviceType, protocol, address, port})
	return nil
}

func (listener *Listener)Start() error{
	go listener.routine()
	return nil
}

func (listener *Listener)Stop() error{
	listener.queryReceiver.Close()
	<- listener.exitChan
	return nil
}

func (listener *Listener)routine(){
	var buffer = make([]byte, DefaultBufferSize)
	if listener.OutputLog{
		log.Printf("start listen at %s", listener.queryReceiver.LocalAddr().String())
	}
	for  {
		bytes, source, err := listener.queryReceiver.ReadFromUDP(buffer)
		if err != nil{
			if listener.OutputLog{
				log.Printf("stop listener due to %s", err.Error())
			}
			break
		}
		var ping Message
		if err = json.Unmarshal(buffer[:bytes], &ping);err != nil{
			if listener.OutputLog{
				log.Printf("invalid ping packet received from %s: %s", source.String(), err.Error())
			}
			continue
		}
		if err = listener.handlePingRequest(ping, source); err != nil{
			if listener.OutputLog{
				log.Printf("handle ping from %s fail: %s", source.String(), err.Error())
			}
		}else if listener.OutputLog{
			log.Printf("response to ping %d from %s", ping.ID, source.String())
		}
	}
	listener.exitChan <- true
}

func (listener *Listener)handlePingRequest(ping Message, source *net.UDPAddr) error{
	if ping.Type != SonarPing{
		return fmt.Errorf("invalid message type %d", ping.Type)
	}
	if ping.Domain != listener.domain{
		return fmt.Errorf("invalid domain %s", ping.Domain)
	}
	var requestAddress = source.IP.String()
	var echo = Message{Type:SonarEcho, ID:ping.ID, Requestor:requestAddress, Services:listener.services}
	packet, err := json.Marshal(echo)
	if err != nil{
		return err
	}
	if _, err = listener.echoSender.WriteToUDP(packet, listener.pingerAddress);err != nil{
		return err
	}
	return nil
}
