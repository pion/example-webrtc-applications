# Text-to-Speech Example

This example converts text to speech with eSpeak NG and streams the audio to a browser over WebRTC. It uses Pion's
pure Go Opus encoder.

## Install `espeak-ng`

Your system must have the `espeak-ng` executable installed and available in your `PATH`.

### macOS

```sh
brew install espeak-ng
```

### Ubuntu / Debian

```sh
sudo apt update
sudo apt install espeak-ng
```

Verify the installation:

```sh
espeak-ng --version
```

## Run the example

Clone the repository and enter the example directory:

```sh
git clone https://github.com/pion/example-webrtc-applications.git
cd example-webrtc-applications/text-to-speech
```

Start the server:

```sh
go run main.go
```

Open [http://localhost:8080](http://localhost:8080) in your browser.

## Usage

1. Wait for the ICE connection state to show `connected`.
2. Enter text in the text area.
3. Click **Convert to Speech** to hear it in the browser.

## How it works

- The browser sends text to the Go server over a WebRTC data channel.
- eSpeak NG produces mono 16-bit WAV audio at 22.05 kHz.
- The server converts the PCM samples to 48 kHz and encodes 20 ms Opus frames.
- Generated speech is queued. When no speech is available, the server continues sending encoded Opus silence.

## License

This project is licensed under the MIT License. See the repository's `LICENSE` file for details.
