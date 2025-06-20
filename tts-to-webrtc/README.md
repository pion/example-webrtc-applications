# WebRTC Text-to-Speech Example

This is an example app that combines WebRTC with OpenAI's Text-to-Speech API to stream audio in real-time.

## Prerequisites

- Go 1.20 or later
- An OpenAI API key
- Web browser with WebRTC support (Chrome, Firefox, Safari, etc.)

## Installation

1. Clone the repository:
```bash
git clone <https://github.com/pion/example-webrtc-applications>
cd tts-to-webrtc
```

2. Install module dependencies:

[Resampler](https://github.com/dh1tw/gosamplerate) and [opus encoder](https://github.com/hraban/opus) packages are using  cgo modules and need to setup. Follow the instructions below to install the required packages.

Linux:
using apt (Ubuntu), yum (Centos)...etc.
```bash
    $ sudo apt install libsamplerate0 pkg-config libopus-dev libopusfile-dev
```

MacOS
using Homebrew:
```bash
    $ brew install libsamplerate pkg-config opus opusfile
```

3. Install Go dependencies:
```bash
export GO111MODULE=on
go install github.com/pion/example-webrtc-applications/v4/tts-to-webrtc@latest
```

## Configuration

Set your OpenAI API key as an environment variable:

```bash
export OPENAI_API_KEY=your_api_key_here
```

## Running the Application

1. Start the server:
```bash
go run main.go
```

2. Open your web browser and navigate to:
```
http://localhost:8080
```

## Usage

1. Click the "Connect" button to establish a WebRTC connection
2. Wait for the connection status to show "connected"
3. Type some text in the textarea
4. Click "Convert to Speech" to hear the text being spoken

## Technical Details

- The application uses OpenAI's TTS API to convert text to speech
- Audio is streamed using WebRTC with Opus codec
- Sample rate conversion is handled automatically (24kHz to 48kHz)
- The server implements a simple audio buffer to handle streaming



## License

This project is licensed under the MIT License - see the LICENSE file for details.