import 'dart:async';
import 'dart:convert';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:flutter/material.dart';
import 'package:websocket/websocket.dart';

void main() => runApp(MyApp());

class MyApp extends StatefulWidget {
  @override
  _GetMyAppState createState() => _GetMyAppState();
}

class _GetMyAppState extends State<MyApp> {
  // Local media
  final _localRenderer = RTCVideoRenderer();
  List _remoteRenderers = [];

  WebSocket _socket;
  RTCPeerConnection _peerConnection;

  @override
  void initState() {
    super.initState();
    connect();
  }

  Future<void> connect() async {
    _peerConnection = await createPeerConnection({}, {});

    await _localRenderer.initialize();
    var localStream = await navigator.mediaDevices
        .getUserMedia({'audio': true, 'video': true});
    _localRenderer.srcObject = localStream;

    localStream.getTracks().forEach((track) async {
      await _peerConnection.addTrack(track, localStream);
    });

    _peerConnection.onIceCandidate = (candidate) {
      if (candidate == null) {
        return;
      }

      _socket.add(JsonEncoder().convert({
        "event": "candidate",
        "data": JsonEncoder().convert({
          'sdpMLineIndex': candidate.sdpMlineIndex,
          'sdpMid': candidate.sdpMid,
          'candidate': candidate.candidate,
        })
      }));
    };

    _peerConnection.onTrack = (event) async {
      if (event.track.kind == 'video' && event.streams.isNotEmpty) {
	var renderer = RTCVideoRenderer();
	await renderer.initialize();
	renderer.srcObject = event.streams[0];

	setState(() { _remoteRenderers.add(renderer); });
      }
    };

    _peerConnection.onRemoveStream = (stream) {
      var rendererToRemove;
      var newRenderList = [];

      // Filter existing renderers for the stream that has been stopped
      _remoteRenderers.forEach((r) {
        if (r.srcObject.id == stream.id) {
		rendererToRemove = r;
	} else {
	  newRenderList.add(r);
	}
      });

      // Set the new renderer list
      setState(() { _remoteRenderers = newRenderList; });

      // Dispose the renderer we are done with
      if (rendererToRemove != null) {
        rendererToRemove.dispose();
      }
    };

    _socket = await WebSocket.connect('ws://localhost:8080/websocket');
    _socket.stream.listen((raw) async {
      Map<String, dynamic> msg = jsonDecode(raw);

      switch (msg['event']) {
        case 'candidate':
          Map<String, dynamic> parsed = jsonDecode(msg['data']);
          _peerConnection
              .addCandidate(RTCIceCandidate(parsed['candidate'], null, 0));
          return;
        case 'offer':
          Map<String, dynamic> offer = jsonDecode(msg['data']);

          // SetRemoteDescription and create answer
          await _peerConnection.setRemoteDescription(
              RTCSessionDescription(offer['sdp'], offer['type']));
          RTCSessionDescription answer = await _peerConnection.createAnswer({});
          await _peerConnection.setLocalDescription(answer);

          // Send answer over WebSocket
          _socket.add(JsonEncoder().convert({
            'event': 'answer',
            'data':
                JsonEncoder().convert({'type': answer.type, 'sdp': answer.sdp})
          }));
          return;
      }
    }, onDone: () {
      print('Closed by server!');
    });
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
        title: 'sfu-ws',
        home: Scaffold(
            appBar: AppBar(
              title: Text('sfu-ws'),
            ),
            body: OrientationBuilder(builder: (context, orientation) {
              return Column(
                children: [
                  Row(
                    children: [
                      Text('Local Video', style: TextStyle(fontSize: 50.0))
                    ],
                  ),
                  Row(
                    children: [
                      SizedBox(
                          width: 160,
                          height: 120,
                          child: RTCVideoView(_localRenderer, mirror: true))
                    ],
                  ),
                  Row(
                    children: [
                      Text('Remote Video', style: TextStyle(fontSize: 50.0))
                    ],
                  ),
                  Row(
                    children: [
                      ..._remoteRenderers.map((remoteRenderer) {
                        return SizedBox(
                            width: 160,
                            height: 120,
                            child: RTCVideoView(remoteRenderer));
                      }).toList(),
                    ],
                  ),
                  Row(
                    children: [
                      Text('Logs Video', style: TextStyle(fontSize: 50.0))
                    ],
                  ),
                ],
              );
            })));
  }
}
