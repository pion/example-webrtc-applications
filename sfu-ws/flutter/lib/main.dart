// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

void main() => runApp(const MyApp());

class MyApp extends StatefulWidget {
  const MyApp({super.key});

  @override
  State<MyApp> createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  static const _textStyle = TextStyle(fontSize: 24);

  // Local media
  final _localRenderer = RTCVideoRenderer();
  List<RTCVideoRenderer> _remoteRenderers = [];

  WebSocketChannel? _socket;
  late final RTCPeerConnection _peerConnection;

  _MyAppState();

  @override
  void initState() {
    super.initState();
    connect();
  }

  Future<void> connect() async {
    _peerConnection = await createPeerConnection({}, {});

    await _localRenderer.initialize();
    final localStream = await navigator.mediaDevices
        .getUserMedia({'audio': true, 'video': true});
    _localRenderer.srcObject = localStream;

    localStream.getTracks().forEach((track) async {
      await _peerConnection.addTrack(track, localStream);
    });

    _peerConnection.onIceCandidate = (candidate) {
      _socket?.sink.add(jsonEncode({
        "event": "candidate",
        "data": jsonEncode({
          'sdpMLineIndex': candidate.sdpMLineIndex,
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

        setState(() {
          _remoteRenderers.add(renderer);
        });
      }
    };

    _peerConnection.onRemoveStream = (stream) {
      RTCVideoRenderer? rendererToRemove;
      final newRenderList = <RTCVideoRenderer>[];

      // Filter existing renderers for the stream that has been stopped
      for (final r in _remoteRenderers) {
        if (r.srcObject?.id == stream.id) {
          rendererToRemove = r;
        } else {
          newRenderList.add(r);
        }
      }

      // Set the new renderer list
      setState(() {
        _remoteRenderers = newRenderList;
      });

      // Dispose the renderer we are done with
      if (rendererToRemove != null) {
        rendererToRemove.dispose();
      }
    };

    final socket =
        WebSocketChannel.connect(Uri.parse('ws://localhost:8080/websocket'));
    _socket = socket;
    socket.stream.listen((raw) async {
      Map<String, dynamic> msg = jsonDecode(raw);

      switch (msg['event']) {
        case 'candidate':
          final parsed = jsonDecode(msg['data']);
          _peerConnection
              .addCandidate(RTCIceCandidate(parsed['candidate'], '', 0));
          return;
        case 'offer':
          final offer = jsonDecode(msg['data']);

          // SetRemoteDescription and create answer
          await _peerConnection.setRemoteDescription(
              RTCSessionDescription(offer['sdp'], offer['type']));
          RTCSessionDescription answer = await _peerConnection.createAnswer({});
          await _peerConnection.setLocalDescription(answer);

          // Send answer over WebSocket
          _socket?.sink.add(jsonEncode({
            'event': 'answer',
            'data': jsonEncode({'type': answer.type, 'sdp': answer.sdp}),
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
          title: const Text('sfu-ws'),
        ),
        body: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Local Video', style: _textStyle),
            SizedBox(
              width: 160,
              height: 120,
              child: RTCVideoView(_localRenderer, mirror: true),
            ),
            const Text('Remote Video', style: _textStyle),
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
            const Text('Logs Video', style: _textStyle),
          ],
        ),
      ),
    );
  }
}
