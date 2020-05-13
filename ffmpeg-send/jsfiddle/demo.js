/* eslint-env browser */

let pc = new RTCPeerConnection({
  iceServers: [
    {
      urls: "stun:stun.l.google.com:19302",
    },
  ],
});
let log = (msg) => {
  document.getElementById("logs").innerHTML += msg + "<br>";
};
let displayVideo = (videoStream) => {
  var el = document.createElement("video");
  el.srcObject = videoStream;
  el.autoplay = true;
  el.muted = true;
  el.width = 1920;
  el.height = 1080;

  document.getElementById("localVideos").appendChild(el);
  return video;
};

pc.addTransceiver("video", { direction: "sendrecv" });
pc.ontrack = (event) => {
  if (event.streams && event.streams[0]) {
    displayVideo(event.streams[0]);
  }
};

pc.createOffer()
  .then((d) => pc.setLocalDescription(d))
  .catch(log);

pc.oniceconnectionstatechange = (e) => log(pc.iceConnectionState);
pc.onicecandidate = (event) => {
  if (event.candidate === null) {
    document.getElementById("localSessionDescription").value = btoa(
      JSON.stringify(pc.localDescription)
    );
  }
};

window.startSession = () => {
  let sd = document.getElementById("remoteSessionDescription").value;
  if (sd === "") {
    return alert("Session Description must not be empty");
  }

  try {
    pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(sd))));
  } catch (e) {
    alert(e);
  }
};
