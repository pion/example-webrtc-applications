package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/pion/example-webrtc-applications/v3/internal/signal"

	"github.com/at-wat/ebml-go/mkvcore"
	"github.com/at-wat/ebml-go/webm"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
)

type mkvSaver struct {
	audioWriter, videoWriter       webm.BlockWriteCloser
	audioBuilder, videoBuilder     *samplebuilder.SampleBuilder
	audioTimestamp, videoTimestamp time.Duration
}

func newMkvSaver() *mkvSaver {
	return &mkvSaver{
		audioBuilder: samplebuilder.New(32, &codecs.OpusPacket{}, 48000),
		videoBuilder: samplebuilder.New(1024, &codecs.H264Packet{}, 90000),
	}
}

func (s *mkvSaver) Close() {
	fmt.Printf("Finalizing mkv...\n")
	if s.audioWriter != nil {
		if err := s.audioWriter.Close(); err != nil {
			panic(err)
		}
	}
	if s.videoWriter != nil {
		if err := s.videoWriter.Close(); err != nil {
			panic(err)
		}
	}
}
func (s *mkvSaver) PushOpus(rtpPacket *rtp.Packet) {
	s.audioBuilder.Push(rtpPacket)

	for {
		sample := s.audioBuilder.Pop()
		if sample == nil {
			return
		}
		if s.audioWriter != nil {
			s.audioTimestamp += sample.Duration
			if _, err := s.audioWriter.Write(true, int64(s.videoTimestamp/time.Millisecond), sample.Data); err != nil {
				panic(err)
			}
		}
	}
}
func (s *mkvSaver) Push264(rtpPacket *rtp.Packet) {
	s.videoBuilder.Push(rtpPacket)

	for {
		sample := s.videoBuilder.Pop()
		if sample == nil {
			return
		}
		naluType := sample.Data[4] & 0x1F
		videoKeyframe := (naluType == 7) || (naluType == 8)
		if videoKeyframe {
			if (s.videoWriter == nil || s.audioWriter == nil) && naluType == 7 {
				p := bytes.SplitN(sample.Data[4:], []byte{0x00, 0x00, 0x00, 0x01}, 2)
				if width, height, fps, ok := H264DecodeSps(p[0], uint(len(p[0]))); ok {
					log.Printf("width:%d, height:%d, fps:%d", width, height, fps)
					s.InitWriter(width, height)
				}
			}
		}
		if s.videoWriter != nil {
			s.videoTimestamp += sample.Duration
			if _, err := s.videoWriter.Write(videoKeyframe, int64(s.videoTimestamp/time.Millisecond), sample.Data); err != nil {
				panic(err)
			}
		}

	}
}
func (s *mkvSaver) InitWriter(width, height int) {
	w, err := os.OpenFile("test.mkv", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	var desc []mkvcore.TrackDescription
	desc = append(desc,
		mkvcore.TrackDescription{
			TrackNumber: uint64(1),
			TrackEntry: webm.TrackEntry{
				Name:        "Audio",
				TrackNumber: 1,
				CodecID:     "A_OPUS",
				TrackType:   2,
				Audio: &webm.Audio{
					SamplingFrequency: 48000.0,
					Channels:          2,
				},
			},
		},
	)

	desc = append(desc,
		mkvcore.TrackDescription{
			TrackNumber: uint64(2),
			TrackEntry: webm.TrackEntry{
				Name:        "Video",
				TrackNumber: 2,
				CodecID:     "V_MPEG4/ISO/AVC",
				TrackType:   1,
				Video: &webm.Video{
					PixelWidth:  uint64(width),
					PixelHeight: uint64(height),
				},
			},
		},
	)
	header := webm.DefaultEBMLHeader
	header.DocType = "matroska"
	ws, _ := mkvcore.NewSimpleBlockWriter(
		w, desc,
		mkvcore.WithEBMLHeader(header),
		mkvcore.WithSegmentInfo(webm.DefaultSegmentInfo),
		mkvcore.WithBlockInterceptor(webm.DefaultBlockInterceptor),
	)

	s.audioWriter = ws[0]
	s.videoWriter = ws[1]
}

func Ue(pBuff []byte, nLen uint, nStartBit *uint) uint {
	var nZeroNum int = 0
	for *nStartBit < nLen*8 {
		if (pBuff[*nStartBit/8] & (0x80 >> (*nStartBit % 8))) > 0 {
			break
		}
		nZeroNum++
		*nStartBit++
	}
	*nStartBit++

	var dwRet uint = 0
	for i := 0; i < nZeroNum; i++ {
		dwRet <<= 1
		if (pBuff[*nStartBit/8] & (0x80 >> (*nStartBit % 8))) > 0 {
			dwRet += 1
		}
		*nStartBit++
	}
	return (1 << nZeroNum) - 1 + dwRet
}

func Se(pBuff []byte, nLen uint, nStartBit *uint) int {
	UeVal := Ue(pBuff, nLen, nStartBit)
	var nValue int = (int)(math.Ceil((float64)(UeVal) / 2))
	if UeVal%2 == 0 {
		nValue = -nValue
	}
	return nValue
}

func u(bitCount uint, buf []byte, nStartBit *uint) uint {
	var dwRet uint = 0
	for i := uint(0); i < bitCount; i++ {
		dwRet <<= 1
		if (buf[*nStartBit/8] & (0x80 >> (*nStartBit % 8))) > 0 {
			dwRet += 1
		}
		*nStartBit++
	}
	return dwRet
}

func deEmulationPrevention(buf []byte, bufSize *uint) {
	tmpPtr := buf
	tmpBufSize := *bufSize
	for i := 0; i < (int)(tmpBufSize-2); i++ {
		val := (int)(tmpPtr[i] + tmpPtr[i+1] + tmpPtr[i+2])
		if val == 0 {
			// kick out 0x03
			for j := i + 2; j < (int)(tmpBufSize-1); j++ {
				tmpPtr[j] = tmpPtr[j+1]
			}
			// and so we should devrease bufsize
			*bufSize--
		}
	}
}

func H264DecodeSps(buf []byte, nLen uint) (int, int, uint, bool) {
	var startBit uint = 0
	var fps uint = 0
	deEmulationPrevention(buf, &nLen)

	u(1, buf, &startBit) // forbidden_zero_bit :=
	u(2, buf, &startBit) // nal_ref_idc :=
	nalUnitType := u(5, buf, &startBit)
	if nalUnitType == 7 {
		profileIdc := u(8, buf, &startBit) //
		_ = u(1, buf, &startBit)           // (buf[1] & 0x80)>>7  constraint_set0_flag
		_ = u(1, buf, &startBit)           // (buf[1] & 0x40)>>6;constraint_set1_flag
		_ = u(1, buf, &startBit)           // (buf[1] & 0x20)>>5;constraint_set2_flag
		_ = u(1, buf, &startBit)           // (buf[1] & 0x10)>>4;constraint_set3_flag
		_ = u(4, buf, &startBit)           // reserved_zero_4bits
		u(8, buf, &startBit)               // level_idc :=

		Ue(buf, nLen, &startBit) // seq_parameter_set_id :=

		if profileIdc == 100 || profileIdc == 110 || profileIdc == 122 || profileIdc == 144 {
			chromaFormatIdc := Ue(buf, nLen, &startBit)
			if chromaFormatIdc == 3 {
				u(1, buf, &startBit) // residual_colou_transform_flag :=
			}

			Ue(buf, nLen, &startBit) // bit_depth_luma_minus8 :=
			Ue(buf, nLen, &startBit) // bit_depth_chroma_minus8 :=
			u(1, buf, &startBit)     // qpprime_y_zero_transform_bypass_flag :=
			seqScalingMatrixPresentFlag := u(1, buf, &startBit)

			seqScalingListPresentFlag := make([]int, 8)
			if seqScalingMatrixPresentFlag > 0 {
				for i := 0; i < 8; i++ {
					seqScalingListPresentFlag[i] = (int)(u(1, buf, &startBit))
				}
			}
		}
		Ue(buf, nLen, &startBit) // log2_max_frame_num_minus4 :=
		picOrderCntType := Ue(buf, nLen, &startBit)
		if picOrderCntType == 0 {
			Ue(buf, nLen, &startBit) // log2_max_pic_order_cnt_lsb_minus4 :=
		} else if picOrderCntType == 1 {
			u(1, buf, &startBit)     // delta_pic_order_always_zero_flag :=
			Se(buf, nLen, &startBit) // offset_for_non_ref_pic :=
			Se(buf, nLen, &startBit) // offset_for_top_to_bottom_field :=
			numRefFramesInPicOrderCntCycle := Ue(buf, nLen, &startBit)

			offsetForRefFrame := make([]int, numRefFramesInPicOrderCntCycle)
			for i := 0; i < (int)(numRefFramesInPicOrderCntCycle); i++ {
				offsetForRefFrame[i] = Se(buf, nLen, &startBit)
			}
		}
		Ue(buf, nLen, &startBit) // num_ref_frames :=
		u(1, buf, &startBit)     // gaps_in_frame_num_value_allowed_flag :=
		picWidthInMbsMinus1 := Ue(buf, nLen, &startBit)
		picHeightInMapUnitsMinus1 := Ue(buf, nLen, &startBit)

		width := (picWidthInMbsMinus1 + 1) * 16
		height := (picHeightInMapUnitsMinus1 + 1) * 16

		frameMbsOnlyFlag := u(1, buf, &startBit)
		if frameMbsOnlyFlag <= 0 {
			u(1, buf, &startBit) // mb_adaptive_frame_field_flag :=
		}

		u(1, buf, &startBit) // direct_8x8_inference_flag :=
		frameCroppingFlag := u(1, buf, &startBit)
		if frameCroppingFlag > 0 {
			Ue(buf, nLen, &startBit) // frame_crop_left_offset:=
			Ue(buf, nLen, &startBit) // frame_crop_right_offset:=
			Ue(buf, nLen, &startBit) // frame_crop_top_offset:=
			Ue(buf, nLen, &startBit) // frame_crop_bottom_offset:=
		}
		vuiParameterPresentFlag := u(1, buf, &startBit)
		if vuiParameterPresentFlag > 0 {
			aspectRatioInfoPresentFlag := u(1, buf, &startBit)
			if aspectRatioInfoPresentFlag > 0 {
				aspectRatioIdc := u(8, buf, &startBit)
				if aspectRatioIdc == 255 {
					u(16, buf, &startBit) // sar_width:=
					u(16, buf, &startBit) // sar_height:=
				}
			}
			overscanInfoPresentFlag := u(1, buf, &startBit)
			if overscanInfoPresentFlag > 0 {
				u(1, buf, &startBit) // overscan_appropriate_flagu:=
			}
			videoSignalTypePresentFlag := u(1, buf, &startBit)
			if videoSignalTypePresentFlag > 0 {
				u(3, buf, &startBit) // video_format:=
				u(1, buf, &startBit) // video_full_range_flag:=
				colorDescriptionPresentFlag := u(1, buf, &startBit)
				if colorDescriptionPresentFlag > 0 {
					u(8, buf, &startBit) // color_primaries:=
					u(8, buf, &startBit) // transfer_characteristics:=
					u(8, buf, &startBit) // matrix_coefficients:=
				}
			}
			chromaLocInfoPresentFlag := u(1, buf, &startBit)
			if chromaLocInfoPresentFlag > 0 {
				Ue(buf, nLen, &startBit) // chroma_sample_loc_type_top_field:=
				Ue(buf, nLen, &startBit) // chroma_sample_loc_type_bottom_field:=
			}
			timingInfoPresentFlag := u(1, buf, &startBit)

			if timingInfoPresentFlag > 0 {
				numUnitsInTick := u(32, buf, &startBit)
				timeScale := u(32, buf, &startBit)
				fps = timeScale / numUnitsInTick
				fixedFrameRateFlag := u(1, buf, &startBit)
				if fixedFrameRateFlag > 0 {
					fps /= 2
				}
			}
		}
		return int(width), int(height), fps, true
	} else {
		return 0, 0, 0, false
	}
}

func main() {
	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.

	// Create a MediaEngine object to configure the supported codec
	m := &webrtc.MediaEngine{}

	// Request H264 and OPUS
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Allow us to receive 1 audio track, and 1 video track
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	} else if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	saver := newMkvSaver()
	defer saver.Close()

	// Set a handler for when a new remote track starts, this handler saves buffers to disk as
	// an ivf file, since we could have multiple video tracks we provide a counter.
	// In your application this is where you would handle/process video
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Printf("Track has started, of type %d: %s \n", track.PayloadType(), track.Codec().RTPCodecCapability.MimeType)
		for {
			// Read RTP packets being sent to Pion
			rtp, _, readErr := track.ReadRTP()
			if readErr != nil {
				if readErr == io.EOF {
					return
				}
				panic(readErr)
			}
			switch track.Kind() {
			case webrtc.RTPCodecTypeAudio:
				saver.PushOpus(rtp)
			case webrtc.RTPCodecTypeVideo:
				saver.Push264(rtp)
			}
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())

		if connectionState == webrtc.ICEConnectionStateConnected {
			fmt.Println("Ctrl+C the remote client to stop the demo")
		} else if connectionState == webrtc.ICEConnectionStateFailed {
			fmt.Println("Done writing media files")
			saver.Close()
			// Gracefully shutdown the peer connection
			if closeErr := peerConnection.Close(); closeErr != nil {
				panic(closeErr)
			}

			os.Exit(0)
		}
	})

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	signal.Decode(signal.MustReadStdin(), &offer)

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the answer in base64 so we can paste it in browser
	fmt.Println(signal.Encode(*peerConnection.LocalDescription()))

	// Block forever
	select {}
}
