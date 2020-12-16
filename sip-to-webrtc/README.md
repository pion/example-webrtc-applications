# Project : succeed in interfacing the pion/webrtc and the SIP

### For that, there are 3 libraries for 3 examples:
- sip-1 : classic sip with the lib [go-sip-ua](https://github.com/cloudwebrtc/go-sip-ua)
- sip-2 : classic sip with the [Kalbi](https://github.com/KalbiProject/Kalbi)
- sip-websocket : sip over websocket with a fork of : [ringcentral-softphone-go](https://github.com/ringcentral/ringcentral-softphone-go)

## WARNING :
1. if I forked the projects, it's because they didn't work originally, I corrected the problems so that they could work in this case!
2. This exemple use [webrtc/v3](https://github.com/pion/webrtc), which is in beta, please be careful
3. We use an ipbx freeswitch, installed with fusionpbx (overlay that adds a web interface. PS: DON'T FORGET TO SAVE THE ADMIN PASSWORD, OR CHANGE IT AT THE FIRST CONNECTION)
## FUSIONPBX / IPBX
How to install : [Fusionpbx](https://docs.fusionpbx.com/en/latest/getting_started/quick_install.html)
For this exemple we need 2 SIP account :
- 1 for a softphone (like linphone (softphone for android, windows etc ) : [Linphone](https://www.linphone.org/))
- 1 for the Golang
 

To Create a SIP account : ```Accounts > Extensions > Add``` (for the exemple : login : ```100``` & password ```100``` and login : ```101``` & password : ```101```)
To Open a websocket for sip-websocket : ```Advanced > SIP profile > Internal``` : Add a new key : ```ws-binding``` / ```:5066```
To configure correctly a classic sip for the pion/webrtc you need :
- Advanced > Variables : Add a new variable :  Category : ```Others```/ Name : ```media_webrtc``` / Value: ```true``` then ```ctrl+s```
- Add a keys in freeswitch file : ```/etc/freeswitch/vars.xml``` and add : ```<X-PRE-PROCESS cmd="set" data="add_ice_candidate=true" />```
- In ```Advanced > SIP profile > Internal``` : Set all nat to true or enabled to true and in ext-rtp-ip and ext-sip-ip set ```auto-nat```
When all are configured, you can ```systemctl restart freeswitch```

### To use it just go to the example folder and go install

In the config : ```192.168.1.74``` is my local ip and ```192.168.1.10``` is my ipbx ip (in a VM)


##### exemple:
```
cd sip-1
go install
### CHANGE THE CONFIG IN MAIN FILE (username / password / domain ip) and :
go run sip-go-sip-ua.go
```

