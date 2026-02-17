package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/dh1tw/gosamplerate"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"gopkg.in/hraban/opus.v2"
)

type WebRTCMessage struct {
	SDP string `json:"sdp"`
}

type TTSRequest struct {
	Text string `json:"text"`
}

type TTSResponse struct {
	Audio []byte
}

var peerConnection *webrtc.PeerConnection

var audioTrack *webrtc.TrackLocalStaticSample

var ticker *time.Ticker

var samplesNeeded int

var resampleLength, srcSampleRate int

var speechBuffer []int16

var resampler gosamplerate.Src

var err error

var mu sync.Mutex

func main() {
	srcSampleRate = 24000

	mu = sync.Mutex{}

	ticker = time.NewTicker(20 * time.Millisecond)

	samplesNeeded = int(float32(srcSampleRate) * 0.02) // 16000 Hz * 0.02 seconds

	resampleLength = int(math.Ceil(float64(48000) / float64(srcSampleRate) * float64(samplesNeeded)))

	speechBuffer = make([]int16, 0)

	// Create a new RTCPeerConnection
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	resampler, err = gosamplerate.New(gosamplerate.SRC_SINC_MEDIUM_QUALITY, 1, resampleLength)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/webrtc", handleWebRTC(&config))
	http.HandleFunc("/tts", handleTTS)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebRTC(config *webrtc.Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg WebRTCMessage
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var err error
		peerConnection, err = webrtc.NewPeerConnection(*config)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create a new audio track
		audioTrack, err = webrtc.NewTrackLocalStaticSample(
			webrtc.RTPCodecCapability{MimeType: "audio/opus"},
			"audio",
			"pion",
		)

		go processSpeechBuffer(context.Background(), audioTrack)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err = peerConnection.AddTrack(audioTrack); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		offer := webrtc.SessionDescription{}
		if err := json.Unmarshal([]byte(msg.SDP), &offer); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := peerConnection.SetRemoteDescription(offer); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create answer
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := peerConnection.SetLocalDescription(answer); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(answer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func handleTTS(w http.ResponseWriter, r *http.Request) {
	var req TTSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call OpenAI TTS API
	ttsReq := map[string]interface{}{
		"model":           "tts-1",
		"response_format": "pcm",
		"input":           req.Text,
		"voice":           "alloy",
	}

	jsonData, err := json.Marshal(ttsReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	request, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/speech", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	request.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	audio, err := io.ReadAll(response.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reader := bytes.NewReader(audio)
	var sample int16
	var samples []int16

	for {
		// Read each sample (2 bytes for int16)
		err := binary.Read(reader, binary.LittleEndian, &sample)
		if err != nil {
			if err == io.EOF {
				break // End of stream
			}
			log.Println("Failed to read sample:", err)
			return
		}

		// Append the int16 sample directly
		samples = append(samples, sample)
	}

	mu.Lock()
	speechBuffer = append(speechBuffer, samples...)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "TTS request received"}`))
}

func processSpeechBuffer(ctx context.Context, localTrack *webrtc.TrackLocalStaticSample) {
	ctxx, cancel := context.WithCancel(ctx)
	defer cancel()

	encoder, err := opus.NewEncoder(48000, 2, opus.AppVoIP)
	if err != nil {
		log.Printf("Failed to create opus encoder: %v", err)
		return
	}

	audioPayload := make([]byte, 1450)

	for {
		select {
		case <-ctxx.Done():
			return
		case <-ticker.C:
			mu.Lock()
			if len(speechBuffer) > 0 {
				// get earliest samples needed from buffer
				samplesToTake := min(len(speechBuffer), samplesNeeded)
				samples := speechBuffer[:samplesToTake]
				speechBuffer = speechBuffer[samplesToTake:]

				mu.Unlock()

				// convert int16 samples to float32
				floatSamples := convertToFloat32(samples)

				// resamples audio data to target sample rate
				resampled, err := resample(floatSamples, srcSampleRate, 48000)
				if err != nil {
					log.Printf("Failed to resample audio: %v", err)
					continue
				}

				// convert from mono to stereo
				stereoSamples := make([]float32, len(resampled)*2)
				for i, sample := range resampled {
					stereoSamples[i*2] = sample   // Left channel
					stereoSamples[i*2+1] = sample // Right channel
				}

				// encode pcm data to opus
				n, err := encoder.EncodeFloat32(stereoSamples, audioPayload)
				if err != nil {
					log.Printf("Failed to encode audio: %v", err)
					continue
				}

				if n > 0 {
					sampleDuration := time.Duration((float32(len(resampled))/48000)*1000) * time.Millisecond

					// write opus packet to local track
					if err = localTrack.WriteSample(media.Sample{
						Data:     audioPayload[:n],
						Duration: sampleDuration,
					}); err != nil {
						log.Printf("Failed to write audio to track: %v", err)
						continue
					}
				}
			} else {
				mu.Unlock()
				// write a silence opus sample to local track
				silencePayload := []byte{0xf8, 0xff, 0xfe} // Opus silence frame

				err = localTrack.WriteSample(media.Sample{
					Data:     silencePayload,
					Duration: 20 * time.Millisecond,
				})
				if err != nil {
					log.Printf("Failed to write silence to track: %v", err)
				}
			}
		}
	}
}

func convertToFloat32(pcm []int16) []float32 {
	samples := make([]float32, len(pcm))
	for i, v := range pcm {
		samples[i] = float32(v) / float32(1<<15-1)
	}
	return samples
}

func resample(samples []float32, sourceRate int, targetRate int) ([]float32, error) {
	ratio := float64(targetRate) / float64(sourceRate)
	resampled, err := resampler.Process(samples, ratio, false)
	if err != nil {
		return nil, err
	}

	return resampled, nil
}
