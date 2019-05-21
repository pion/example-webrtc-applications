import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:math';
import 'package:flutter_webrtc/webrtc.dart';
import 'package:sdp_transform/sdp_transform.dart' as sdp_transform;

class SfuWsSample {
  JsonEncoder _jsonEnc = new JsonEncoder();
  WebSocket _socket;
  MediaStream _stream;
  RTCPeerConnection _pc;
  RTCDataChannel _dc;
  dynamic onLocalStream;
  dynamic onRemoteStream;
  dynamic onOpen;
  dynamic onClose;
  dynamic onError;

  SfuWsSample();

  /*Replace the payload to adapt SFU-WS */
  _changePlyload(description) {
    var session = sdp_transform.parse(description.sdp);
    print('session => ' + _jsonEnc.convert(session));

    var videoIdx = 1;

    /*
     * DefaultPayloadTypeG722 = 9
     * DefaultPayloadTypeOpus = 111
     * DefaultPayloadTypeVP8  = 96
     * DefaultPayloadTypeVP9  = 98
     * DefaultPayloadTypeH264 = 100
    */
    /*Add VP8 and RTX only.*/
    var rtp = [
      {"payload": 96, "codec": "VP8", "rate": 90000, "encoding": null},
      {"payload": 97, "codec": "rtx", "rate": 90000, "encoding": null}
    ];

    session['media'][videoIdx]["payloads"] = "96 97";
    session['media'][videoIdx]["rtp"] = rtp;

    var fmtp = [
      {"payload": 97, "config": "apt=96"}
    ];

    session['media'][videoIdx]["fmtp"] = fmtp;

    var rtcpFB = [
      {"payload": 96, "type": "transport-cc", "subtype": null},
      {"payload": 96, "type": "ccm", "subtype": "fir"},
      {"payload": 96, "type": "nack", "subtype": null},
      {"payload": 96, "type": "nack", "subtype": "pli"}
    ];
    session['media'][videoIdx]["rtcpFb"] = rtcpFB;

    var sdp = sdp_transform.write(session, null);
    return new RTCSessionDescription(sdp, description.type);
  }

  bool _inCalling = false;

  bool get inCalling => _inCalling;

  Map<String, dynamic> configuration = {
    "iceServers": [
      {"url": "stun:stun.l.google.com:19302"},
    ]
  };

  final Map<String, dynamic> _config = {
    'mandatory': {},
    'optional': [
      {'DtlsSrtpKeyAgreement': true},
    ],
  };

  final Map<String, dynamic> _constraints = {
    'mandatory': {
      'OfferToReceiveAudio': true,
      'OfferToReceiveVideo': true,
    },
    'optional': [],
  };

  Future<void> connect(String host) async {
    if (_socket != null) {
      print('Already connected!');
      return;
    }

    try {
      Random r = new Random();
      String key = base64.encode(List<int>.generate(8, (_) => r.nextInt(255)));
      SecurityContext securityContext = new SecurityContext();
      HttpClient client = HttpClient(context: securityContext);
      client.badCertificateCallback =
          (X509Certificate cert, String host, int port) {
        print('Allow self-signed certificate => $host:$port.');
        return true;
      };
      HttpClientRequest request = await client.getUrl(
          Uri.parse('https://$host:8443/ws')); // form the correct url here
      request.headers.add('Connection', 'Upgrade');
      request.headers.add('Upgrade', 'websocket');
      request.headers.add(
          'Sec-WebSocket-Version', '13'); // insert the correct version here
      request.headers.add('Sec-WebSocket-Key', key.toLowerCase());

      HttpClientResponse response = await request.close();
      Socket socket = await response.detachSocket();
      _socket = WebSocket.fromUpgradedSocket(
        socket,
        protocol: 'pions-flutter',
        serverSide: false,
      );
      _socket.listen((data) {
        print('Recivied data: ' + data);
        _onMessage(data);
      }, onDone: () {
        print('Closed by server!');
        if (this.onClose != null) this.onClose();
        _socket = null;
      });
      if (this.onOpen != null) this.onOpen();
      return;
    } catch (e) {
      print(e.toString());
      if (this.onError != null) this.onError(e.toString());
      _socket = null;
      return;
    }
  }

  void _send(String data) {
    if (_socket != null) _socket.add(data);
    print('send: ' + data);
  }

  void _onMessage(data) {
    if (_pc == null) return;
    _pc.setRemoteDescription(new RTCSessionDescription(data, 'answer'));
  }

  void createPublisher() async {
    if (_inCalling) {
      return;
    }
    final Map<String, dynamic> mediaConstraints = {
      "audio": true,
      "video": {
        "mandatory": {
          "minWidth":
              '640', // Provide your own width, height and frame rate here
          "minHeight": '480',
          "minFrameRate": '30',
        },
        "facingMode": "user",
        "optional": [],
      }
    };
    _stream = await navigator.getUserMedia(mediaConstraints);
    if (this.onLocalStream != null) this.onLocalStream(_stream);
    _pc = await createPeerConnection(configuration, _config);
    _dc = await _pc.createDataChannel('data', RTCDataChannelInit());
    _pc.onIceGatheringState = (state) async {
      if (state == RTCIceGatheringState.RTCIceGatheringStateComplete) {
        print('RTCIceGatheringStateComplete');
        RTCSessionDescription sdp = await _pc.getLocalDescription();
        _send(sdp.sdp);
      }
    };
    _pc.addStream(_stream);
    RTCSessionDescription description = await _pc.createOffer(_constraints);
    /*Use sdp-transform to replace payload type. :(*/
    description = _changePlyload(description);
    print('Publisher createOffer');
    _pc.setLocalDescription(description);
    _inCalling = true;
  }

  void createSubscriber() async {
    if (_inCalling) {
      return;
    }

    _pc = await createPeerConnection(configuration, _config);
    _dc = await _pc.createDataChannel('data', RTCDataChannelInit());

    _pc.onIceGatheringState = (state) async {
      if (state == RTCIceGatheringState.RTCIceGatheringStateComplete) {
        print('RTCIceGatheringStateComplete');
        RTCSessionDescription sdp = await _pc.getLocalDescription();
        _send(sdp.sdp);
      }
    };

    _pc.onAddStream = (stream) {
      print('Got remote stream => ' + stream.id);
      _stream = stream;
      if (this.onRemoteStream != null) this.onRemoteStream(stream);
    };

    RTCSessionDescription description = await _pc.createOffer(_constraints);
    /*Use sdp-transform to replace payload type. :(*/
    description = _changePlyload(description);
    print('Subscriber createOffer');
    _pc.setLocalDescription(description);
    _inCalling = true;
  }

  void close() async {
    if (_stream != null) await _stream.dispose();
    if (_pc != null) await _pc.close();
    if (_socket != null) {
      await _socket.close();
      _socket = null;
    }
  }
}
