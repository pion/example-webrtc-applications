package transaction

import "time"

/*
Author - Aaron Parfitt
Date - 11th October 2020

RFC3261 - SIP: Session Initiation Protocol
https://tools.ietf.org/html/rfc3261#section-17.1.1.2


T1 timer
T1 is an estimate of the round-trip time (RTT), and
   it defaults to 500 ms.  Nearly all of the transaction timers
   described here scale with T1, and changing T1 adjusts their values.

T2 timer
The default value of T2 is 4s,
   and it represents the amount of time a non-INVITE server transaction
   will take to respond to a request, if it does not respond
   immediately.  For the default values of T1 and T2, this results in
   intervals of 500 ms, 1 s, 2 s, 4 s, 4 s, 4 s, etc.
*/

const (
	//T1 is a timer described in RFC3261.
	T1 = 500 * time.Millisecond
	//T2 is a timer described in RFC3261.
	T2 = 4 * time.Second
)
