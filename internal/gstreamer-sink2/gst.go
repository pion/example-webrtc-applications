package gst

import (
	"fmt"

	gst_ "github.com/notedit/gst"
	"github.com/pion/webrtc/v2"
)

// StartMainLoop starts GLib's main loop
// It needs to be called from the process' main thread
// Because many gstreamer plugins require access to the main thread
// See: https://golang.org/pkg/runtime/#LockOSThread
func StartMainLoop() {
	gst_.MainLoopNew().Run()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	pipeline *gst_.Pipeline
	appsrc   *gst_.Element
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(codecName string) *Pipeline {
	pipelineStr := "appsrc format=time is-live=true do-timestamp=true name=src ! application/x-rtp"
	switch codecName {
	case webrtc.VP8:
		pipelineStr += ", encoding-name=VP8-DRAFT-IETF-01 ! rtpvp8depay ! decodebin ! autovideosink"
	case webrtc.Opus:
		pipelineStr += ", payload=96, encoding-name=OPUS ! rtpopusdepay ! decodebin ! autoaudiosink"
	case webrtc.VP9:
		pipelineStr += " ! rtpvp9depay ! decodebin ! autovideosink"
	case webrtc.H264:
		pipelineStr += " ! rtph264depay ! decodebin ! autovideosink"
	case webrtc.G722:
		pipelineStr += " clock-rate=8000 ! rtpg722depay ! decodebin ! autoaudiosink"
	default:
		panic("Unhandled codec " + codecName)
	}

	pipeline_, err := gst_.ParseLaunch(pipelineStr)

	if err != nil {
		fmt.Println(err)
		panic("pipeline init error")
	}

	pipeline := &Pipeline{
		pipeline: pipeline_,
		appsrc:   pipeline_.GetByName("src"),
	}

	return pipeline
}

// Start starts the GStreamer Pipeline
func (p *Pipeline) Start() {
	p.pipeline.SetState(gst_.StatePlaying)
}

// Stop stops the GStreamer Pipeline
func (p *Pipeline) Stop() {
	p.pipeline.SetState(gst_.StateNull)
}

// Push pushes a buffer on the appsrc of the GStreamer Pipeline
func (p *Pipeline) Push(buffer []byte) {

	err := p.appsrc.PushBuffer(buffer)

	if err != nil {
		panic(err)
	}
}
