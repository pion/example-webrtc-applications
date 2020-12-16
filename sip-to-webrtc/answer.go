package sip

import (
	"fmt"

	"github.com/pion/webrtc/v3"
	log "github.com/sirupsen/logrus"

	"strconv"
	"strings"
)

var myLocalIP = "192.168.1.74"

var audioMimeType = webrtc.MimeTypePCMU
//var audioMimeType = webrtc.MimeTypeOpus //IF YOU CHOOSE OPUS, DONT FORGET TO ACTIVATE mod_opus and configure it in the ipbx

//TOOLS TO MANAGE SDP ARE ON BOTTOM OF THIS FILE

func Answer(offerSDP string, answerSDP chan string) {
	sdp := CompleteTheOfferSDP(offerSDP)
	fmt.Printf("### MY OFFER SDP \n %s\n###   END  ###\n", sdp)
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp,
	}

	mediaEngine := webrtc.MediaEngine{}

	var PayloadTypeAudio uint8
	PayloadTypeAudio = getPayload(offer.SDP)
	var ClockRateAudio uint32
	switch audioMimeType {
	case webrtc.MimeTypeOpus:
		ClockRateAudio = 48000
	case webrtc.MimeTypePCMU:
		ClockRateAudio = 8000
	default:
		panic("NO CODEC AVAILABLE")
	}

	if err := mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: audioMimeType, ClockRate: ClockRateAudio, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        webrtc.PayloadType(PayloadTypeAudio),
	}, webrtc.RTPCodecTypeAudio); err != nil {
		log.Println("FAIL TO REGISTER CODEC AUDIO:", err)
		return
	}

	//SETING ENGINE
	settingEngine := webrtc.SettingEngine{}
	//https://github.com/Sean-Der/ringcentral-softphone-go/blob/master/softphone.go#L154
	if err := settingEngine.SetAnsweringDTLSRole(webrtc.DTLSRoleServer); err != nil { // TRYING TO CHANGE THE DTLSROLE, LIKE THE PR OF THE RINGCENTRAL-SOFTPHONE FROM Sean
		panic(err)
	}
	if err := settingEngine.SetEphemeralUDPPortRange(16384, 16484); err != nil { // DEFINE THE RANGE PORT FOR THE FIREWALL
		panic(err)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine), webrtc.WithSettingEngine(settingEngine))

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"}, //IF NO STUN, FREESWITCH FAIL
			},
		},
	}

	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil { //NEW AUDIO TRANSCEIVER
		panic(err)
	}

	audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: audioMimeType}, "audio", "audio")
	if err != nil {
		log.Printf("ERROR sendTrackVideoToCaller NewTrackLocalStaticRTP audio: %v\n", err)
	}
	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		log.Printf("ERROR sendTrackVideoToCaller AddTrack audio : %v\n", err)
	}
	//TODO :  !! DONT FORGET TO SEND AUDIO TRACK ! (With Udp and gstreamer)
	//go sendTrack(audioTrack)

	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) { //PLEASE, I NEED THIS EVENT FIRED
		fmt.Printf("OnTrack\n")
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) { //NEVER CONNECTED :(
		fmt.Printf("OnICEConnectionStateChange %s \n", connectionState.String())
	})

	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}


	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Println("SetLocalDesc error = ", err)
		fmt.Println(err)
	}
	gatherCompleat := webrtc.GatheringCompletePromise(peerConnection) //NEW EVENT TO GET FULL ANSWER !

	fmt.Printf("### MY SDP BEFORE COMPLEAT\n %s\n###   END  ###\n", answer.SDP)
	select {
	case <-gatherCompleat: //TRIGER , I CAN SEND MY NEW ANSWER
		localDesc := peerConnection.CurrentLocalDescription()
		fmt.Println("Finished Local description: ", localDesc)
		answer = *localDesc
	}

	//YOu can comment / uncomment all function bellow for test
	answer.SDP = addWMSAndChangeSetup(answer.SDP) // add msid-semantic: WMS with token and set the setup to active
	answer.SDP = addRtcp(answer.SDP) // add a=rtcp:9 IN IP4 0.0.0.0
	answer.SDP = addInfoToCandidate(answer.SDP) // ADD generation 0 network-id 1 in candidate and a=ice-options:trickle
	answer.SDP = strings.ReplaceAll(answer.SDP, "0.0.0.0", myLocalIP) //TRY TO SET MY LOCAL IP


	fmt.Printf("### MY SDP AFTER COMPLEAT AND CHANGE\n%s\n###   END  ###\n", answer.SDP)
	answerSDP <- answer.SDP

	select {}
}

/*
	TOOLS PART FOR SDP ! ANSWER AND OFFER
*/

/*
	OFFER
*/
func CompleteTheOfferSDP(sdp string) (string) {
	//ADD : a=sendrecv and a=mid if missing and a=ice-lite
	hasMid := strings.Contains(sdp, "a=mid:0")
	hasSendrecv := strings.Contains(sdp, "a=sendrecv")
	matchSDPCodecLine := ""
	switch audioMimeType {
	case webrtc.MimeTypeOpus:
		matchSDPCodecLine = "opus/48000/2"
	case webrtc.MimeTypePCMU:
		matchSDPCodecLine = "PCMU/8000"
	}
	sdpSplited := strings.Split(sdp, "\r\n")
	for i := 0; i < len(sdpSplited); i++ {
		if strings.HasPrefix(sdpSplited[i], "t=0 0") { //ADD ice-lite
			sdpSplited[i] += "\r\na=ice-lite"
		} else if strings.HasSuffix(sdpSplited[i], matchSDPCodecLine) {
			if !hasSendrecv { //add sendrecv
				sdpSplited[i] = "a=sendrecv\r\n" + sdpSplited[i]
			}
			if !hasMid { //add a=mid:0
				sdpSplited[i] = "a=mid:0\r\n" + sdpSplited[i]
			}
		}
	}
	return strings.Join(sdpSplited, "\r\n") + "\r\n"
}

func getPayload(sdp string) (uint8) {
	matchSDPCodecLine := ""
	switch audioMimeType {
	case webrtc.MimeTypeOpus:
		matchSDPCodecLine = "opus/48000/2"
	case webrtc.MimeTypePCMU:
		matchSDPCodecLine = "PCMU/8000"
	}
	sdpSplited := strings.Split(sdp, "\r\n")
	for i := 0; i < len(sdpSplited); i++ {
		if strings.HasSuffix(sdpSplited[i], matchSDPCodecLine) {
			first := strings.Split(sdpSplited[i], " ")[0]
			second := strings.Split(first, ":")[1] //correct payload
			s, _ := strconv.ParseUint(second, 10, 8)
			return uint8(s)
		}
	}
	return 0
}


/*
	ANSWER
	ALL FUNCTION HERE WAS FOR TEST
*/


func getTokenInSdp(sdp string) (string) {
	sdpTab := strings.Split(sdp, "\r\n")
	for i := 0; i < len(sdpTab); i++ {
		if strings.HasPrefix(sdpTab[i], "a=ssrc") {
			//correct line exemple : a=ssrc:1686957819 msid:hKKbwoRVShpPiDpT RrrwHaHjcXAZTXNc
			first := strings.Split(sdpTab[i], " ")[1] //msid:hKKbwoRVShpPiDpT
			return strings.Split(first, ":")[1]       //correct payload //hKKbwoRVShpPiDpT
		}
	}
	return ""
}

func addWMSAndChangeSetup(sdp string) (string) {
	token := getTokenInSdp(sdp)
	sdpTab := strings.Split(sdp, "\r\n")
	for i := 0; i < len(sdpTab); i++ {
		if strings.HasPrefix(sdpTab[i], "t=0 0") {
			sdpTab[i] += "\r\na=msid-semantic: WMS "+token
		} else if strings.HasPrefix(sdpTab[i], "a=setup:") {
			sdpTab[i] = "a=setup:active"
		}
	}
	return strings.Join(sdpTab, "\r\n")+ "\r\n"
}
func addRtcp(sdp string) (string) {
	sdpTab := strings.Split(sdp, "\r\n")
	for i := 0; i < len(sdpTab); i++ {
		if strings.HasPrefix(sdpTab[i], "c=IN") {
			sdpTab[i] += "\r\na=rtcp:9 IN IP4 0.0.0.0"
			return strings.Join(sdpTab, "\r\n")
		}
	}
	return ""
}

//BuildCorrectCandidate(answer.SDP, "192.168.1.74")
func addInfoToCandidate(sdp string) (string) {

	sdpTab := strings.Split(sdp, "\r\n")
	for i := 0; i < len(sdpTab); i++ { //
		if strings.HasPrefix(sdpTab[i], "a=candidate:") {
			sdpTab[i] += " generation 0 network-id 1"
		} else if strings.HasPrefix(sdpTab[i], "a=ice-pwd:") {
			sdpTab[i] += "\r\na=ice-options:trickle"
		}
	}
	return strings.Join(sdpTab, "\r\n") + "\r\n"
}


/*func dropallCandidate(sdp string) (string) {

	sdpTab := strings.Split(sdp, "\r\n")
	for i := 0; i < len(sdpTab); i++ {
		if strings.HasPrefix(sdpTab[i], "a=candidate:") {
			return strings.Join(sdpTab[:i], "\r\n") + "\r\n"
		}
	}
	return ""
}

func BuildCorrectCandidate(sdp string, ipAddr string) (string) { //TRYING TO CREATE CANDIDATE
	var port1, line1, line2 string
	sdpTab := strings.Split(sdp, "\r\n")
	for i := 0; i < len(sdpTab); i++ {
		if strings.HasPrefix(sdpTab[i], "a=candidate:") {
			if strings.Contains(sdpTab[i], ipAddr){
				port1 = strings.Split(sdpTab[i], " ")[5]
			} else if strings.Contains(sdpTab[i], "srflx") && len(line1) == 0 {
				line1 = sdpTab[i]
				//port2 = strings.Split(sdpTab[i], " ")[5]
			} else if strings.Contains(sdpTab[i], "srflx") {
				line2 = sdpTab[i]
			}
		}
	}
	line1Split := strings.Split(line1, " ")
	line2Split := strings.Split(line2, " ")
	line1Split[4] = ipAddr
	line2Split[4] = ipAddr
	line1Split[5] = port1
	line2Split[5] = port1
	line1Split[11] = port1
	line2Split[11] = port1

	return dropallCandidate(sdp) + strings.Join(line1Split, " ") + "\r\n" + strings.Join(line2Split, " ") + "\r\n"

}*/

func getPort(sdp string) (string) { //get a RTP port

	sdpTab := strings.Split(sdp, "\r\n")
	for i := 0; i < len(sdpTab); i++ {
		if strings.HasPrefix(sdpTab[i], "a=candidate:") {
			if strings.Contains(sdpTab[i], "srflx") {
				return strings.Split(sdpTab[i], " ")[5]
			}
		}
	}
	return ""
}

func BuildCorrectCandidate(sdp string, ipAddr string) (string) { //Trying to set the same port of all candidate and set the port in m=audio part !
	port := getPort(sdp)
	if len(port) == 0 {
		panic("NO STUN SERVER :( FREESWITCH FAIL ! ")
	}
	sdpTab := strings.Split(sdp, "\r\n")
	for i := 0; i < len(sdpTab); i++ {
		if strings.HasPrefix(sdpTab[i], "m=audio") {
			lineSplited := strings.Split(sdpTab[i], " ")
			lineSplited[1] = port
			sdpTab[i] = strings.Join(lineSplited, " ")
		} else if strings.HasPrefix(sdpTab[i], "a=candidate:") {
			if strings.Contains(sdpTab[i], "host") {
				lineSplited := strings.Split(sdpTab[i], " ")
				lineSplited[5] = port
				sdpTab[i] = strings.Join(lineSplited, " ")
				//port2 = strings.Split(sdpTab[i], " ")[5]
			}
		}
	}

	return strings.Join(sdpTab, "\r\n") + "\r\n"

}
