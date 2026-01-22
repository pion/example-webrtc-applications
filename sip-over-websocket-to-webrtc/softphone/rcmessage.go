// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package softphone

import "encoding/xml"

// Msg ...
type Msg struct {
	XMLName xml.Name `xml:"Msg"`
	Hdr     Hdr      `xml:"Hdr"`
	Bdy     Bdy      `xml:"Bdy"`
}

// Hdr ...
type Hdr struct {
	XMLName xml.Name `xml:"Hdr"`
	SID     string   `xml:"SID,attr"`
	Req     string   `xml:"Req,attr"`
	From    string   `xml:"From,attr"`
	To      string   `xml:"To,attr"`
	Cmd     string   `xml:"Cmd,attr"`
}

// Bdy ...
type Bdy struct {
	XMLName   xml.Name `xml:"Bdy"`
	SrvLvl    string   `xml:"SrvLvl,attr"`
	SrvLvlExt string   `xml:"SrvLvlExt,attr"`
	Phn       string   `xml:"Phn,attr"`
	Nm        string   `xml:"Nm,attr"`
	ToPhn     string   `xml:"ToPhn,attr"`
	ToNm      string   `xml:"ToNm,attr"`
	RecURL    string   `xml:"RecUrl,attr"`
}
