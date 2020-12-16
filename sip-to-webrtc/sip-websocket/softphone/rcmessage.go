package softphone

import "encoding/xml"

/*
<Msg>
<Hdr SID="35482343635848" Req="{17EA740D-D207-401D-A449-62ABE81CB56C}" From="#1032016@sip.ringcentral.com:5060" To="17203861294*115" Cmd="6"/>
<Bdy SrvLvl="-149699523" SrvLvlExt="406" Phn="+16504306662" Nm="WIRELESS CALLER" ToPhn="+16504223279" ToNm="Tyler Liu" RecUrl=""/>
</Msg>
*/

type Msg struct {
	XMLName xml.Name `xml:"Msg"`
	Hdr     Hdr      `xml:"Hdr"`
	Bdy     Bdy      `xml:"Bdy"`
}

type Hdr struct {
	XMLName xml.Name `xml:"Hdr"`
	SID     string   `xml:"SID,attr"`
	Req     string   `xml:"Req,attr"`
	From    string   `xml:"From,attr"`
	To      string   `xml:"To,attr"`
	Cmd     string   `xml:"Cmd,attr"`
}

type Bdy struct {
	XMLName   xml.Name `xml:"Bdy"`
	SrvLvl    string   `xml:"SrvLvl,attr"`
	SrvLvlExt string   `xml:"SrvLvlExt,attr"`
	Phn       string   `xml:"Phn,attr"`
	Nm        string   `xml:"Nm,attr"`
	ToPhn     string   `xml:"ToPhn,attr"`
	ToNm      string   `xml:"ToNm,attr"`
	RecUrl    string   `xml:"RecUrl,attr"`
}
