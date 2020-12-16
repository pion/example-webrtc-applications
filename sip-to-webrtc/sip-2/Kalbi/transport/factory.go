package transport

import (
	"Kalbi/interfaces"
	"Kalbi/log"
	"strconv"
)

// NewTransportListenPoint creates listen
func NewTransportListenPoint(protocol string, host string, port int) interfaces.ListeningPoint {
	switch protocol {
	case "udp":
		log.Log.Info("Creating UDP listening point")
		listner := new(UDPTransport)
		log.Log.Info("Binding to " + host + ":" + strconv.Itoa(port))
		listner.Build(host, port)
		return listner
	case "tcp":
		log.Log.Info("Creating TCP listening point")
		listner := new(UDPTransport)
		log.Log.Info("Binding to " + host + ":" + strconv.Itoa(port))
		listner.Build(host, port)
		return listner
	case "tls":
		log.Log.Info("Creating TLS listening point")
		listner := new(UDPTransport)
		log.Log.Info("Binding to " + host + ":" + strconv.Itoa(port))
		listner.Build(host, port)
		return listner
	default:
		log.Log.Info("Unknown protocol specified")
		panic("Unknown protocol specified")

	}

}
