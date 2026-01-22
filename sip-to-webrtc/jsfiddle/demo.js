// SPDX-FileCopyrightText: 2026 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

/* eslint-env browser */

let log = msg => {
  document.getElementById('logs').innerHTML += msg + '<br>'
}

let pc = new RTCPeerConnection({
  // iceServers: [
  //   {
  //     urls: 'stun:stun.l.google.com:19302'
  //   }
  // ]
})

pc.addTransceiver('audio', {direction: 'recvonly'})
pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
pc.onicecandidate = event => {
  if (event.candidate === null) {
    document.getElementById('localSessionDescription').value = btoa(JSON.stringify(pc.localDescription))
  }
}

pc.ontrack = e => {
  document.getElementById('remote-audio').srcObject = e.streams[0]
}

pc.createOffer().then(d => pc.setLocalDescription(d)).catch(log)

window.startSession = () => {
  let sd = document.getElementById('remoteSessionDescription').value
  if (sd === '') {
    return alert('Session Description must not be empty')
  }

  try {
    pc.setRemoteDescription(JSON.parse(atob(sd)))
  } catch (e) {
    alert(e)
  }
}
