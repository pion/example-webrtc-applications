# gocv-to-webrtc

gocv-to-webrtc demonstrates how to stream live webcam video to a browser using GoCV for camera capture,
FFmpeg for real‑time VP8 encoding, and Pion WebRTC for media transport.

This project could be a starting point to building a computer vision project that is viewable via WebRTC.
This is implemented using the following pieces.

- **Capture**: Uses [GoCV](https://gocv.io/) to access a webcam and read raw BGR frames.
- **Encode**: Pipes raw frames into `ffmpeg` for VP8 encoding in IVF format.
- **Stream**: Uses [Pion WebRTC](https://github.com/pion/webrtc) to send encoded video frames to a browser client.
- **Frontend**: Minimal HTML/JS page that negotiates WebRTC Offer/Answer and displays incoming video.

## Instructions
### Install Dependencies
This example requires you have GoCV and ffmpeg installed, these are the supported platforms
#### Debian/Ubuntu
* Follow the setup instructions for [GoCV](https://github.com/hybridgroup/gocv)
* `sudo apt-get install ffmpeg`
#### macOS
* `brew install ffmpeg opencv`

### Build gocv-to-webrtc
```
go build -tags gocv
```

### Run gocv-to-webrtc
```bash
gocv-to-webrtc
```
2. Open your browser at `http://localhost:8080`.
3. Click **Start Session** to initiate WebRTC negotiation.
4. After ICE connects, you should see your webcam video in the page.

## How It Works

### Server (`main.go`)

1. **HTTP Server**
   - Serves `index.html` port 8080.
   - Handles `/offer` endpoint for SDP exchange.

2. **WebRTC Setup**
   - Reads the browser’s SDP Offer.
   - Creates a Pion `PeerConnection` with a VP8 track (`TrackLocalStaticSample`).
   - Sets remote description, creates Answer, and returns it once ICE gathering completes.
   - Starts the camera stream after ICE connection.

3. **Video Pipeline** (`startCameraAndStream`)
   - Opens webcam via GoCV (`gocv.OpenVideoCapture`).
   - Pipes raw BGR frames into FFmpeg:
     ```bash
     ffmpeg -y \
       -f rawvideo -pixel_format bgr24 -video_size 640x480 -framerate 30 -i pipe:0 \
       -c:v libvpx -b:v 1M -f ivf pipe:1
     ```
   - Reads VP8 IVF frames from FFmpeg’s stdout with `ivfreader`.
   - Writes frames into the WebRTC track.

### Frontend (`index.html`)

1. Creates an `RTCPeerConnection` with STUN.
2. Adds a `recvonly` video transceiver.
3. Sends SDP Offer to server.
4. Sets remote Answer.
5. Attaches incoming stream to a `<video>` element.

---

## Configuration

- **Camera Device**: Change `gocv.OpenVideoCapture(2)` to the appropriate index (e.g., `0`).
- **Resolution & Frame Rate**: Adjust GoCV settings and FFmpeg flags (`-video_size`, `-framerate`).
- **Frame Rate**: Adjust `-framerate 30` in FFmpeg and ticker interval in Go.
- **Bitrate & Codec**: Modify `-b:v 1M` or swap codecs (H264/Opus).

---

## Troubleshooting

- **No Video**: Check camera index and FFmpeg installation.
- **ICE Fails**: Verify STUN server and network/firewall.
- **High CPU**: Lower resolution/bitrate or tune FFmpeg CPU usage.
