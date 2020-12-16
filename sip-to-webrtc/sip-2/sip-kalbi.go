package main

import (
	"Kalbi"
	"Kalbi/authentication"
	"Kalbi/interfaces"
	"Kalbi/sip/message"
	"Kalbi/sip/method"
	"Kalbi/sip/status"
	"fmt"
	"sip"
	"time"
)

const (
	title = "Kalbi Simple SIP CLient"
	prompt = "kalbiclient> ")


type ClientProperties struct {
	IP          string
	Username    string
	Domain      string
	Password    string
	Registrar   string
}

type Client struct {
	stack       *kalbi.SipStack
	properties  *ClientProperties
}

func main() {
	// DOES NOT WORK, LOOP WITH ERROR : Client Transaction already exists
	//EXEMPLE AND LIB FROM THE https://github.com/KalbiProject/Kalbi/
	// IT'S Forked because Go get fail and dont get the authentication folder
	client := new(Client)
	client.properties = &ClientProperties{IP:"192.168.1.74", Username:"100", Password:"100", Domain:"192.168.1.10", Registrar:"192.168.1.10"}
	client.Start(client.properties.IP, 5060)
	client.SendRegister()
	client.basicCliInterface()
}


func (c *Client) HandleRequests(event interfaces.SipEventObject) {

	tx := event.GetTransaction()
	switch string(tx.GetLastMessage().Req.Method) {
	case method.CANCEL:
		//handle CANCEL request
	case method.INVITE:

		trans := event.GetTransaction().GetOrigin()
		answerSDPChan := make(chan string)
		sip.Answer(string(event.GetSipMessage().Body), answerSDPChan) // BUILD  ANSWER WITH PION !
		answerSDP := <-answerSDPChan
		responseLine := message.NewResponseLine(status.OK, answerSDP)
		via := &message.SipVia{} //fmt.Sprintf("SIP/2.0/WSS %s;branch=%s", "vQDZgy4701HmD.invalid", "z9hG4bKvQDZgy4701HmD")
		msg := message.NewResponse(responseLine, via, &message.SipTo{UriType:"100@192.168.1.10"}, &message.SipFrom{UriType:"mod_sofia@192.168.1.10"}, &trans.CallId, &trans.MaxFwd)
		msg.CopyHeaders(tx.GetOrigin())
		msg.ContLen.SetValue("0")
		tx.Send(msg, string(tx.GetOrigin().Contact.Host), string(tx.GetOrigin().Contact.Port))

	case method.REGISTER:
		//handle REGISTER request
	case method.BYE:
		//handle BYE request
	case method.ACK:

	default:
		responseLine := message.NewResponseLine(status.OK, "It's Cool")
		msg := message.NewResponse(responseLine, nil, nil, nil, nil, nil)
		msg.CopyHeaders(tx.GetOrigin())
		msg.ContLen.SetValue("0")
		tx.Send(msg, string(tx.GetOrigin().Contact.Host), string(tx.GetOrigin().Contact.Port))
	}

}


func (c *Client) ReRegister(event interfaces.SipEventObject){
	//response := event.GetSipMessage()

	time.Sleep(60 * time.Second)

	originRequest :=  event.GetTransaction().GetOrigin()

	requestLine := message.NewRequestLine(method.REGISTER, "sip", c.properties.Username, c.properties.Domain, "5060") //Create requestline e.g.  REGISTER sip:1234@127.0.0.1:5060 SIP/2.0
	requestVia := message.NewViaHeader("udp", c.properties.IP, "5060") //Creates Via e.g. Via: SIP/2.0/UDP 127.0.0.1:5060
	requestVia.SetBranch(message.GenerateBranchId())//Generate Branch
	requestFrom := message.NewFromHeader(c.properties.Username, "sip", c.properties.Domain, "5060") //Creates From e.g. From: <sip:1234@127.0.0.1>
	requestFrom.SetTag("3234jhf23")
	requestTo := message.NewToHeader(c.properties.Username, "sip", c.properties.Domain, "5060") //Creates To e.g. To: <sip:5678@127.0.0.1>
	requestContact := message.NewContactHeader("sip", c.properties.Username, c.properties.IP) //Creates contact header

	requestCallID := message.NewCallID(message.GenerateNewCallID())//Creates CallID e.g. Call-ID: 123456789

	requestCseq := message.NewCSeq("1", method.REGISTER) //Creates CSeq e.g. CSeq: 1 INVITE

	requestMaxFor := message.NewMaxForwards("70") //Creates Max-Forwards e.g. Max-Forwards: 70
	requestContentLen := message.NewContentLength("0")//Creates Content Length Header
	request := message.NewRequest(requestLine, requestVia , requestTo, requestFrom, requestContact, requestCallID, requestCseq, requestMaxFor, requestContentLen)

	request.SetAuthHeader(&originRequest.Auth)

	txmng := c.stack.GetTransactionManager()
	txmng.NewClientTransaction(request)
	c.stack.ListeningPoints[0].Send(c.properties.Registrar, "5060", request.String())


}



func (c *Client) HandleResponses(event interfaces.SipEventObject) {

	response := event.GetTransaction()
	fmt.Println(string(event.GetSipMessage().Src))
	switch response.GetLastMessage().GetStatusCode() {

	case 100:
		//Handle 100 Trying
	case 180:
		//Handle 180 Ringing
	case 200:
		if string(response.GetOrigin().Req.Method) == method.REGISTER {
			go c.ReRegister(event)
		}
	case 401:
		c.HandleUnAuth(event)
	default:
		//Handle Default
	}

}


func (c *Client) HandleUnAuth(event interfaces.SipEventObject) {
	response := event.GetSipMessage()

	origin := event.GetTransaction().GetOrigin()

	//copy original auth header
	authHeader := response.Auth


	authHeader.SetCNonce("nwqlcqw80wnf")
	authHeader.SetUsername(c.properties.Username)
	authHeader.SetNc("00000001")
	authHeader.SetURI("sip:"+ c.properties.Domain)
	authHeader.SetResponse(authentication.MD5Challange(authHeader.GetUsername(), authHeader.GetRealm(), c.properties.Password, authHeader.GetURI(), authHeader.GetNonce(), authHeader.GetCNonce(), authHeader.GetNc(), authHeader.GetQoP(), string(origin.Req.Method) ))
	origin.SetAuthHeader(&authHeader)
	if string(event.GetTransaction().GetOrigin().Req.Method) != "INVITE"{
		origin.CallId.SetValue(message.GenerateNewCallID())
	}

	txmng := c.stack.GetTransactionManager()
	tx := txmng.NewClientTransaction(origin)
	tx.Send(origin, c.properties.Registrar, "5060")


}

func (c *Client) SendRegister() {

	requestLine := message.NewRequestLine(method.REGISTER, "sip", c.properties.Username, c.properties.Domain, "5060") //Create requestline e.g.  REGISTER sip:1234@127.0.0.1:5060 SIP/2.0
	requestVia := message.NewViaHeader("udp", c.properties.IP, "5060") //Creates Via e.g. Via: SIP/2.0/UDP 127.0.0.1:5060
	requestVia.SetBranch(message.GenerateBranchId())//Generate Branch
	requestFrom := message.NewFromHeader(c.properties.Username, "sip", c.properties.Domain, "5060") //Creates From e.g. From: <sip:1234@127.0.0.1>
	requestFrom.SetTag("3234jhf23")
	requestTo := message.NewToHeader(c.properties.Username, "sip", c.properties.Domain, "5060") //Creates To e.g. To: <sip:5678@127.0.0.1>
	requestContact := message.NewContactHeader("sip", c.properties.Username, c.properties.IP) //Creates contact header
	requestCallID := message.NewCallID(message.GenerateNewCallID())//Creates CallID e.g. Call-ID: 123456789
	requestCseq := message.NewCSeq("1", method.REGISTER) //Creates CSeq e.g. CSeq: 1 INVITE
	requestMaxFor := message.NewMaxForwards("70") //Creates Max-Forwards e.g. Max-Forwards: 70
	requestContentLen := message.NewContentLength("0")//Creates Content Length Header
	request := message.NewRequest(requestLine, requestVia , requestTo, requestFrom, requestContact, requestCallID, requestCseq, requestMaxFor, requestContentLen)

	txmng := c.stack.GetTransactionManager()
	txmng.NewClientTransaction(request)
	c.stack.ListeningPoints[0].Send(c.properties.Registrar, "5060", request.String())
}


func (c *Client) SendInvite(to string) {

	requestLine := message.NewRequestLine(method.INVITE, "sip", to, c.properties.Domain, "5060") //Create requestline e.g.  REGISTER sip:1234@127.0.0.1:5060 SIP/2.0
	requestVia := message.NewViaHeader("udp", c.properties.IP, "5060") //Creates Via e.g. Via: SIP/2.0/UDP 127.0.0.1:5060
	requestVia.SetBranch(message.GenerateBranchId())//Generate Branch
	requestFrom := message.NewFromHeader(c.properties.Username, "sip", c.properties.Domain, "5060") //Creates From e.g. From: <sip:1234@127.0.0.1>
	requestFrom.SetTag("3234jhf23")
	requestTo := message.NewToHeader(to, "sip", c.properties.Domain, "5060") //Creates To e.g. To: <sip:5678@127.0.0.1>
	requestContact := message.NewContactHeader("sip", c.properties.Username, c.properties.IP) //Creates contact header
	requestCallID := message.NewCallID(message.GenerateNewCallID())//Creates CallID e.g. Call-ID: 123456789
	requestCseq := message.NewCSeq("1", method.INVITE) //Creates CSeq e.g. CSeq: 1 INVITE
	requestMaxFor := message.NewMaxForwards("70") //Creates Max-Forwards e.g. Max-Forwards: 70
	requestContentLen := message.NewContentLength("0")//Creates Content Length Header
	request := message.NewRequest(requestLine, requestVia , requestTo, requestFrom, requestContact, requestCallID, requestCseq, requestMaxFor, requestContentLen)

	txmng := c.stack.GetTransactionManager()
	tx := txmng.NewClientTransaction(request)
	tx.Send(request, c.properties.Registrar, "5060")
}

func (c *Client) Start(host string, port int) {
	c.stack = kalbi.NewSipStack("Basic Client Example")
	c.stack.CreateListenPoint("udp", host, port)
	c.stack.SetSipListener(c)
	go c.stack.Start()
}



func configure() *ClientProperties {
	props := new(ClientProperties)
	fmt.Println(title + "\n")

	//Username
	fmt.Print("Username:")
	_, err := fmt.Scan(&props.Username)
	if err != nil {
		fmt.Println(err)
	}

	//Domain
	fmt.Print("Domain:")
	_, err = fmt.Scan(&props.Domain)
	if err != nil {
		fmt.Println(err)
	}

	//Password will be visable because I cba to set stty echo off using syscalls
	fmt.Print("Password:")
	_, err = fmt.Scan(&props.Password)
	if err != nil {
		fmt.Println(err)
	}

	//IP
	fmt.Print("Machine IP:")
	_, err = fmt.Scan(&props.IP)
	if err != nil {
		fmt.Println(err)
	}

	//Registrar
	fmt.Print("Registrar:")
	_, err = fmt.Scan(&props.Registrar)
	if err != nil {
		fmt.Println(err)
	}

	return props
}


func (c *Client) basicCliInterface() {


	alive := true

	for alive == true {
		var command string
		fmt.Print(prompt)

		fmt.Scanln(&command)

		switch command {
		case "exit":
			alive = false
			fmt.Println("Exiting...")
		case "":
			continue
		default:
			fmt.Println("Unknown command")

		}

	}

}
