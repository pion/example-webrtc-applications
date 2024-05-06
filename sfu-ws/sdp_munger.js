/*

OFFER FROM CLIENT TO RECEIVE FLEX-FEC THAT HAS BEEN MUNGED BY RECEIVING CLIENT.

                description.sdp = `v=0\r
o=- 3936474526705670197 2 IN IP4 127.0.0.1\r
s=-\r
t=0 0\r
a=group:BUNDLE 0 1\r
a=msid-semantic: WMS 65e82b7e-1d89-485b-aac2-7d8b28c3c9e9\r
m=audio 9 UDP/TLS/RTP/SAVPF 111 63 9 0 8 13 110 126\r
c=IN IP4 0.0.0.0\r
a=rtcp:9 IN IP4 0.0.0.0\r
a=ice-ufrag:ASpm\r
a=ice-pwd:jZ30ohe+1A4YJVgqD4/8d6PY\r
a=ice-options:trickle\r
${fingerprintLine}\r
a=setup:actpass\r
a=mid:0\r
a=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r
a=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r
a=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r
a=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\r
a=recvonly\r
a=rtcp-mux\r
a=rtpmap:111 opus/48000/2\r
a=rtcp-fb:111 transport-cc\r
a=fmtp:111 minptime=10;useinbandfec=1\r
a=rtpmap:63 red/48000/2\r
a=fmtp:63 111/111\r
a=rtpmap:9 G722/8000\r
a=rtpmap:0 PCMU/8000\r
a=rtpmap:8 PCMA/8000\r
a=rtpmap:13 CN/8000\r
a=rtpmap:110 telephone-event/48000\r
a=rtpmap:126 telephone-event/8000\r
m=video 9 UDP/TLS/RTP/SAVPF 96 97 98 99 100 101 102 122 127 121 125 107 108 109 124 120 123 119 117\r
c=IN IP4 0.0.0.0\r
a=rtcp:9 IN IP4 0.0.0.0\r
a=ice-ufrag:ASpm\r
a=ice-pwd:jZ30ohe+1A4YJVgqD4/8d6PY\r
a=ice-options:trickle\r
${fingerprintLine}\r
a=setup:active\r
a=mid:1\r
a=extmap:14 urn:ietf:params:rtp-hdrext:toffset\r
a=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r
a=extmap:13 urn:3gpp:video-orientation\r
a=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r
a=extmap:12 http://www.webrtc.org/experiments/rtp-hdrext/playout-delay\r
a=extmap:11 http://www.webrtc.org/experiments/rtp-hdrext/video-content-type\r
a=extmap:7 http://www.webrtc.org/experiments/rtp-hdrext/video-timing\r
a=extmap:9 http://www.webrtc.org/experiments/rtp-hdrext/color-space\r
a=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\r
a=extmap:5 urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id\r
a=extmap:6 urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id\r
a=recvonly\r
a=msid:65e82b7e-1d89-485b-aac2-7d8b28c3c9e9 ed0d5f28-d226-47f7-9098-3c7718a73fc9\r
a=rtcp-mux\r
a=rtcp-rsize\r
a=rtpmap:96 VP8/90000\r
a=rtcp-fb:96 goog-remb\r
a=rtcp-fb:96 transport-cc\r
a=rtcp-fb:96 ccm fir\r
a=rtcp-fb:96 nack\r
a=rtcp-fb:96 nack pli\r
a=rtpmap:97 rtx/90000\r
a=fmtp:97 apt=96\r
a=rtpmap:98 VP9/90000\r
a=rtcp-fb:98 goog-remb\r
a=rtcp-fb:98 transport-cc\r
a=rtcp-fb:98 ccm fir\r
a=rtcp-fb:98 nack\r
a=rtcp-fb:98 nack pli\r
a=fmtp:98 profile-id=0\r
a=rtpmap:99 rtx/90000\r
a=fmtp:99 apt=98\r
a=rtpmap:100 VP9/90000\r
a=rtcp-fb:100 goog-remb\r
a=rtcp-fb:100 transport-cc\r
a=rtcp-fb:100 ccm fir\r
a=rtcp-fb:100 nack\r
a=rtcp-fb:100 nack pli\r
a=fmtp:100 profile-id=2\r
a=rtpmap:101 rtx/90000\r
a=fmtp:101 apt=100\r
a=rtpmap:102 H264/90000\r
a=rtcp-fb:102 goog-remb\r
a=rtcp-fb:102 transport-cc\r
a=rtcp-fb:102 ccm fir\r
a=rtcp-fb:102 nack\r
a=rtcp-fb:102 nack pli\r
a=fmtp:102 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f\r
a=rtpmap:122 rtx/90000\r
a=fmtp:122 apt=102\r
a=rtpmap:127 H264/90000\r
a=rtcp-fb:127 goog-remb\r
a=rtcp-fb:127 transport-cc\r
a=rtcp-fb:127 ccm fir\r
a=rtcp-fb:127 nack\r
a=rtcp-fb:127 nack pli\r
a=fmtp:127 level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42001f\r
a=rtpmap:121 rtx/90000\r
a=fmtp:121 apt=127\r
a=rtpmap:125 H264/90000\r
a=rtcp-fb:125 goog-remb\r
a=rtcp-fb:125 transport-cc\r
a=rtcp-fb:125 ccm fir\r
a=rtcp-fb:125 nack\r
a=rtcp-fb:125 nack pli\r
a=fmtp:125 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f\r
a=rtpmap:107 rtx/90000\r
a=fmtp:107 apt=125\r
a=rtpmap:108 H264/90000\r
a=rtcp-fb:108 goog-remb\r
a=rtcp-fb:108 transport-cc\r
a=rtcp-fb:108 ccm fir\r
a=rtcp-fb:108 nack\r
a=rtcp-fb:108 nack pli\r
a=fmtp:108 level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42e01f\r
a=rtpmap:109 rtx/90000\r
a=fmtp:109 apt=108\r
a=rtpmap:124 H264/90000\r
a=rtcp-fb:124 goog-remb\r
a=rtcp-fb:124 transport-cc\r
a=rtcp-fb:124 ccm fir\r
a=rtcp-fb:124 nack\r
a=rtcp-fb:124 nack pli\r
a=fmtp:124 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=4d001f\r
a=rtpmap:120 rtx/90000\r
a=fmtp:120 apt=124\r
a=rtpmap:123 H264/90000\r
a=rtcp-fb:123 goog-remb\r
a=rtcp-fb:123 transport-cc\r
a=rtcp-fb:123 ccm fir\r
a=rtcp-fb:123 nack\r
a=rtcp-fb:123 nack pli\r
a=fmtp:123 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=64001f\r
a=rtpmap:119 rtx/90000\r
a=fmtp:119 apt=123\r
a=rtpmap:117 flexfec-03/90000\r
a=fmtp:117 L:5; D:10; ToP:2; repair-window:200\r
a=ssrc-group:FID 2938260438 1859461734\r
a=ssrc-group:FEC-FR 2938260438 2055559999\r
a=ssrc:2938260438 cname:uwa1YcMXZcIOqtQ4\r
a=ssrc:1859461734 cname:uwa1YcMXZcIOqtQ4\r
a=ssrc:2055559999 cname:uwa1YcMXZcIOqtQ4\r\n`;



ANSWER FROM PION SERVER THAT HAS BEEN MUNGED BY RECEIVING CLIENT.

                    data = `v=0\r
${globalO}\r
s=-\r
t=0 0\r
${fingerprintLine}\r
a=group:BUNDLE 0 1\r
m=audio 9 UDP/TLS/RTP/SAVPF 111\r
c=IN IP4 0.0.0.0\r
a=setup:active\r
a=mid:0\r
${iceUfrag}\r
${icePwd}\r
a=rtcp-mux\r
a=rtcp-rsize\r
a=rtpmap:111 opus/48000/2\r
a=fmtp:111 minptime=10;useinbandfec=1\r
a=rtcp-fb:111 transport-cc\r
a=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r
a=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r
a=ssrc:${audioSsrc} cname:pion\r
a=ssrc:${audioSsrc} msid:pion audio\r
a=ssrc:${audioSsrc} mslabel:pion\r
a=ssrc:${audioSsrc} label:audio\r
a=msid:pion audio\r
a=sendonly\r

.... ice candidates were here .... I removed the ones that were here since it's my "home IP".

a=end-of-candidates\r
m=video 9 UDP/TLS/RTP/SAVPF 96 125 117\r
c=IN IP4 0.0.0.0\r
a=setup:active\r
a=mid:1\r
a=ice-ufrag:tPKLZkxskRMkgucm\r
a=ice-pwd:HcTIrvebAJeolGUJuEzvkvSbdafqYBfw\r
a=rtcp-mux\r
a=rtcp-rsize\r
a=rtpmap:96 VP8/90000\r
a=rtcp-fb:96 goog-remb\r
a=rtcp-fb:96 transport-cc\r
a=rtcp-fb:96 ccm fir\r
a=rtcp-fb:96 nack\r
a=rtcp-fb:96 nack pli\r
a=rtpmap:125 H264/90000\r
a=fmtp:125 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f\r
a=rtcp-fb:125 goog-remb\r
a=rtcp-fb:125 transport-cc\r
a=rtcp-fb:125 ccm fir\r
a=rtcp-fb:125 nack\r
a=rtcp-fb:125 nack pli\r
a=rtpmap:117 flexfec-03/90000\r
a=fmtp:117 L:5; D:10; ToP:2; repair-window:200\r
a=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r
a=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\r
a=extmap:5 urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id\r
a=extmap:6 urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id\r
a=ssrc-group:FEC-FR ${videoSsrc} 2055559999\r
a=ssrc:${videoSsrc} cname:pion\r
a=ssrc:${videoSsrc} msid:pion video\r
a=ssrc:${videoSsrc} mslabel:pion\r
a=ssrc:${videoSsrc} label:video\r
a=ssrc:2055559999 cname:pion\r
a=msid:pion video\r
a=sendonly\r\n`;
*/