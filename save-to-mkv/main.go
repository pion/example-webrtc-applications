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
				if width, height, fps, ok := H264_decode_sps(p[0], uint(len(p[0]))); ok {
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
	var k int64 = (int64)(UeVal)
	var nValue int = (int)(math.Ceil((float64)(k / 2)))
	if UeVal%2 == 0 {
		nValue = -nValue
	}
	return nValue
}

func u(BitCount uint, buf []byte, nStartBit *uint) uint {
	var dwRet uint = 0
	var i uint = 0
	for i = 0; i < BitCount; i++ {
		dwRet <<= 1
		if (buf[*nStartBit/8] & (0x80 >> (*nStartBit % 8))) > 0 {
			dwRet += 1
		}
		*nStartBit++
	}
	return dwRet
}

func de_emulation_prevention(buf []byte, buf_size *uint) {
	i := 0
	j := 0
	var tmp_buf_size uint = 0
	var val int = 0

	tmp_ptr := buf
	tmp_buf_size = *buf_size
	for i = 0; i < (int)(tmp_buf_size-2); i++ {
		val = (int)((tmp_ptr[i] ^ 0x00) + (tmp_ptr[i+1] ^ 0x00) + (tmp_ptr[i+2] ^ 0x03))
		if val == 0 {
			//kick out 0x03
			for j = i + 2; j < (int)(tmp_buf_size-1); j++ {
				tmp_ptr[j] = tmp_ptr[j+1]
			}

			//and so we should devrease bufsize
			*buf_size--
		}
	}
}

func H264_decode_sps(buf []byte, nLen uint) (int, int, uint, bool) {
	var StartBit uint = 0
	var fps uint = 0
	de_emulation_prevention(buf, &nLen)

	u(1, buf, &StartBit) //forbidden_zero_bit :=
	u(2, buf, &StartBit) //nal_ref_idc :=
	nal_unit_type := u(5, buf, &StartBit)
	if nal_unit_type == 7 {
		profile_idc := u(8, buf, &StartBit) //
		_ = u(1, buf, &StartBit)            //(buf[1] & 0x80)>>7  constraint_set0_flag
		_ = u(1, buf, &StartBit)            //(buf[1] & 0x40)>>6;constraint_set1_flag
		_ = u(1, buf, &StartBit)            //(buf[1] & 0x20)>>5;constraint_set2_flag
		_ = u(1, buf, &StartBit)            //(buf[1] & 0x10)>>4;constraint_set3_flag
		_ = u(4, buf, &StartBit)            //reserved_zero_4bits
		u(8, buf, &StartBit)                //level_idc :=

		Ue(buf, nLen, &StartBit) //seq_parameter_set_id :=

		if profile_idc == 100 || profile_idc == 110 || profile_idc == 122 || profile_idc == 144 {
			chroma_format_idc := Ue(buf, nLen, &StartBit)
			if chroma_format_idc == 3 {
				u(1, buf, &StartBit) //residual_colour_transform_flag :=
			}

			Ue(buf, nLen, &StartBit) //bit_depth_luma_minus8 :=
			Ue(buf, nLen, &StartBit) //bit_depth_chroma_minus8 :=
			u(1, buf, &StartBit)     //qpprime_y_zero_transform_bypass_flag :=
			seq_scaling_matrix_present_flag := u(1, buf, &StartBit)

			seq_scaling_list_present_flag := make([]int, 8)
			if seq_scaling_matrix_present_flag > 0 {
				for i := 0; i < 8; i++ {
					seq_scaling_list_present_flag[i] = (int)(u(1, buf, &StartBit))
				}
			}
		}
		Ue(buf, nLen, &StartBit) //log2_max_frame_num_minus4 :=
		pic_order_cnt_type := Ue(buf, nLen, &StartBit)
		if pic_order_cnt_type == 0 {
			Ue(buf, nLen, &StartBit) //log2_max_pic_order_cnt_lsb_minus4 :=
		} else if pic_order_cnt_type == 1 {
			u(1, buf, &StartBit)     //delta_pic_order_always_zero_flag :=
			Se(buf, nLen, &StartBit) //offset_for_non_ref_pic :=
			Se(buf, nLen, &StartBit) //offset_for_top_to_bottom_field :=
			num_ref_frames_in_pic_order_cnt_cycle := Ue(buf, nLen, &StartBit)

			offset_for_ref_frame := make([]int, num_ref_frames_in_pic_order_cnt_cycle)
			for i := 0; i < (int)(num_ref_frames_in_pic_order_cnt_cycle); i++ {
				offset_for_ref_frame[i] = Se(buf, nLen, &StartBit)
			}
		}
		Ue(buf, nLen, &StartBit) //num_ref_frames :=
		u(1, buf, &StartBit)     //gaps_in_frame_num_value_allowed_flag :=
		pic_width_in_mbs_minus1 := Ue(buf, nLen, &StartBit)
		pic_height_in_map_units_minus1 := Ue(buf, nLen, &StartBit)

		width := (pic_width_in_mbs_minus1 + 1) * 16
		height := (pic_height_in_map_units_minus1 + 1) * 16

		frame_mbs_only_flag := u(1, buf, &StartBit)
		if frame_mbs_only_flag <= 0 {
			u(1, buf, &StartBit) //mb_adaptive_frame_field_flag :=
		}

		u(1, buf, &StartBit) //direct_8x8_inference_flag :=
		frame_cropping_flag := u(1, buf, &StartBit)
		if frame_cropping_flag > 0 {
			Ue(buf, nLen, &StartBit) //frame_crop_left_offset:=
			Ue(buf, nLen, &StartBit) //frame_crop_right_offset:=
			Ue(buf, nLen, &StartBit) //frame_crop_top_offset:=
			Ue(buf, nLen, &StartBit) //frame_crop_bottom_offset:=
		}
		vui_parameter_present_flag := u(1, buf, &StartBit)
		if vui_parameter_present_flag > 0 {
			aspect_ratio_info_present_flag := u(1, buf, &StartBit)
			if aspect_ratio_info_present_flag > 0 {
				aspect_ratio_idc := u(8, buf, &StartBit)
				if aspect_ratio_idc == 255 {
					u(16, buf, &StartBit) //sar_width:=
					u(16, buf, &StartBit) //sar_height:=
				}
			}
			overscan_info_present_flag := u(1, buf, &StartBit)
			if overscan_info_present_flag > 0 {
				u(1, buf, &StartBit) //overscan_appropriate_flagu:=
			}
			video_signal_type_present_flag := u(1, buf, &StartBit)
			if video_signal_type_present_flag > 0 {
				u(3, buf, &StartBit) //video_format:=
				u(1, buf, &StartBit) //video_full_range_flag:=
				colour_description_present_flag := u(1, buf, &StartBit)
				if colour_description_present_flag > 0 {
					u(8, buf, &StartBit) //colour_primaries:=
					u(8, buf, &StartBit) //transfer_characteristics:=
					u(8, buf, &StartBit) //matrix_coefficients:=
				}
			}
			chroma_loc_info_present_flag := u(1, buf, &StartBit)
			if chroma_loc_info_present_flag > 0 {
				Ue(buf, nLen, &StartBit) //chroma_sample_loc_type_top_field:=
				Ue(buf, nLen, &StartBit) //chroma_sample_loc_type_bottom_field:=
			}
			timing_info_present_flag := u(1, buf, &StartBit)

			if timing_info_present_flag > 0 {
				num_units_in_tick := u(32, buf, &StartBit)
				time_scale := u(32, buf, &StartBit)
				fps = time_scale / num_units_in_tick
				fixed_frame_rate_flag := u(1, buf, &StartBit)
				if fixed_frame_rate_flag > 0 {
					fps = fps / 2
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

	// Setup the codecs you want to use.
	// We'll use a VP8 and Opus but you can also define your own
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
