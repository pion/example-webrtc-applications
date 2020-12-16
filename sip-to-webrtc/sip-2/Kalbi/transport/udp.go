package transport

import (
	"fmt"
	"Kalbi/interfaces"
	"Kalbi/log"
	"Kalbi/sip"
	"Kalbi/sip/event"
	reuse "github.com/libp2p/go-reuseport"
	"net"
)

//UDPTransport is a network protocol listening point for the EventDispatcher
type UDPTransport struct {
	Host             string
	Port             int
	Address          net.UDPAddr
	Connection       net.PacketConn
	TransportChannel chan interfaces.SipEventObject
}

//Read from UDP Socket
func (ut *UDPTransport) Read() interfaces.SipEventObject {
	buffer := make([]byte, 2048)
	n, _, err := ut.Connection.ReadFrom(buffer)
	if err != nil {
		log.Log.Error(err)
	}
	
	request := sip.Parse(buffer[:n])
	event := new(event.SipEvent)
	event.SetSipMessage(&request)
	return event
}

//GetHost returns ip interface address
func (ut *UDPTransport) GetHost() string {
	return ut.Host
}

//GetPort returns ip interface port
func (ut *UDPTransport) GetPort() int {
	return ut.Port
}

//Build initializes the UDPTransport object
func (ut *UDPTransport) Build(host string, port int) {
	ut.Host = host
	ut.Port = port
	ut.Address = net.UDPAddr{
		IP:   net.ParseIP(host),
		Port: port,
	}

	var err error
	ut.Connection, err = reuse.ListenPacket("udp", ut.Address.String())
	if err != nil {
		panic(err)
	}

}

//Start starts the ListeningPoint
func (ut *UDPTransport) Start() {
	log.Log.Info("Starting UDP Listening Point ")
	for {
		msg := ut.Read()
		ut.TransportChannel <- msg
	}
}

//SetTransportChannel setter that allows to set SipStack's Transport Channel
func (ut *UDPTransport) SetTransportChannel(channel chan interfaces.SipEventObject) {
	ut.TransportChannel = channel
}

//Send allows you to send a SIP message
func (ut *UDPTransport) Send(host string, port string, msg string) error {
	addr, err := net.ResolveUDPAddr("udp", host+":"+port)
	if err != nil {
		log.Log.Error(err)
	}
	log.Log.Info("Sending message to " + host + ":" + port)
	conn, err := reuse.Dial("udp", ut.Address.String(), addr.String())



	
	if err != nil {
		fmt.Printf("Some error %v", err)
		return err
	}
	_, err = conn.Write([]byte(msg))
	if err != nil {
		log.Log.Error(err)
	}
	conn.Close()
	return nil
}
