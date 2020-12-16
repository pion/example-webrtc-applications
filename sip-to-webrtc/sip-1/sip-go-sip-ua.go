package main

import (
	"fmt"
	"github.com/pixelbender/go-sdp/sdp"
	"os"
	"os/signal"
	mySip "sip"
	"syscall"
	"time"

	"github.com/cloudwebrtc/go-sip-ua/pkg/account"
	"github.com/cloudwebrtc/go-sip-ua/pkg/endpoint"
	"github.com/cloudwebrtc/go-sip-ua/pkg/invite"
	"github.com/cloudwebrtc/go-sip-ua/pkg/ua"
	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/sip/parser"
)

const (
	username = "100"
	password = "100"
	displayName = "PionSIP"
	sipServer = "192.168.1.10"
)

var (
	logger log.Logger
)

func init() {
	logger = log.NewDefaultLogrusLogger().WithPrefix("Client")
}

func main() {
	//EXEMPLE FROM https://github.com/cloudwebrtc/go-sip-ua/blob/master/examples/client/main.go
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	endpoint := endpoint.NewEndPoint(&endpoint.EndPointConfig{Extensions: []string{"replaces", "outbound"}, Dns: "8.8.8.8"}, logger)
	// BE CAREFULL WITH THIS LIB, MULTIPLE IP ADDR Make it fail, you can change the function in endpoint.go
	//https://github.com/cloudwebrtc/go-sip-ua/issues/11
	listen := "0.0.0.0:5080"
	logger.Infof("Listen => %s", listen)

	if err := endpoint.Listen("udp", listen); err != nil {
		logger.Panic(err)
	}

	if err := endpoint.Listen("tcp", listen); err != nil {
		logger.Panic(err)
	}

	ua := ua.NewUserAgent(&ua.UserAgentConfig{
		UserAgent: "Pion Sip Client/1.0.0",
		Endpoint:  endpoint,
	}, logger)

	ua.InviteStateHandler = func(sess *invite.Session, req *sip.Request, resp *sip.Response, state invite.Status) {
		logger.Infof("InviteStateHandler: state => %v, type => %s", state, sess.Direction())
		if state == invite.InviteReceived {
			offerSDP := (*req).Body()
			offerSDPSession, _ := sdp.ParseString(offerSDP) //CHECK ERROR PLEASE!
			sess.ProvideOffer(offerSDPSession)
			answerSDPChan := make(chan string)
			mySip.Answer((*req).Body(), answerSDPChan)
			answerSDP := <-answerSDPChan
			answerSDPSession, _ := sdp.ParseString(answerSDP)
			sess.ProvideAnswer(answerSDPSession)
			sess.Accept(200)
		}
	}

	ua.RegisterStateHandler = func(state account.RegisterState) {
		logger.Infof("RegisterStateHandler: user => %s, state => %v, expires => %v", state.Account.Auth.AuthName, state.StatusCode, state.Expiration)
	}

	profile := account.NewProfile(username, displayName,
		&account.AuthInfo{
			AuthName: username,
			Password: password,
			Realm:    "",
		},
		1800,
	)

	uri := fmt.Sprintf("sip:%s@%s:5060;transport=udp", username, sipServer)
	target, err := parser.ParseSipUri(uri)
	if err != nil {
		logger.Error(err)
	}

	go ua.SendRegister(profile, target, profile.Expires)
	time.Sleep(time.Second * 10)
	go ua.SendRegister(profile, target, 0)
	/*
		sdp := mock.answer.String()
		called := "weiweiduan"
		go ua.Invite(profile, &sip.SipUri{
			FUser:      sip.String{Str: called},
			FHost:      target.Host(),
			FPort:      target.Port(),
			FUriParams: target.UriParams(),
		}, &sdp)
	*/
	<-stop

	ua.Shutdown()
}