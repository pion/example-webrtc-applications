package gst

import (
	"fmt"
	"sync"

	gst_ "github.com/notedit/gst"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
)

// Pipeline is a wrapper
type Pipeline struct {
	pipeline  *gst_.Pipeline
	tracks    []*webrtc.Track
	id        int
	codecName string
	clockRate float32
}

var pipelines = make(map[int]*Pipeline)
var pipelinesLock sync.Mutex

const (
	videoClockRate = 90000
	audioClockRate = 48000
	pcmClockRate   = 8000
)

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(codecName string, tracks []*webrtc.Track, pipelineSrc string) *Pipeline {
	pipelineStr := "appsink name=appsink"
	var clockRate float32

	switch codecName {
	case webrtc.VP8:
		pipelineStr = pipelineSrc + " ! vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1 ! " + pipelineStr
		clockRate = videoClockRate

	case webrtc.VP9:
		pipelineStr = pipelineSrc + " ! vp9enc ! " + pipelineStr
		clockRate = videoClockRate

	case webrtc.H264:
		pipelineStr = pipelineSrc + " ! video/x-raw,format=I420 ! x264enc bframes=0 speed-preset=veryfast key-int-max=60 ! video/x-h264,stream-format=byte-stream ! " + pipelineStr
		clockRate = videoClockRate

	case webrtc.Opus:
		pipelineStr = pipelineSrc + " ! opusenc ! " + pipelineStr
		clockRate = audioClockRate

	case webrtc.G722:
		pipelineStr = pipelineSrc + " ! avenc_g722 ! " + pipelineStr
		clockRate = audioClockRate

	case webrtc.PCMU:
		pipelineStr = pipelineSrc + " ! audio/x-raw, rate=8000 ! mulawenc ! " + pipelineStr
		clockRate = pcmClockRate

	case webrtc.PCMA:
		pipelineStr = pipelineSrc + " ! audio/x-raw, rate=8000 ! alawenc ! " + pipelineStr
		clockRate = pcmClockRate

	default:
		panic("Unhandled codec " + codecName)
	}

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	pipeline_, err := gst_.ParseLaunch(pipelineStr)

	if err != nil {
		fmt.Println(err)
		panic("pipeline init error")
	}

	pipeline := &Pipeline{
		pipeline:  pipeline_,
		tracks:    tracks,
		id:        len(pipelines),
		codecName: codecName,
		clockRate: clockRate,
	}

	pipelines[pipeline.id] = pipeline
	return pipeline
}

// Start starts the GStreamer Pipeline
func (p *Pipeline) Start() {

	p.pipeline.SetState(gst_.StatePlaying)

	go pullSample(p)
}

// Stop stops the GStreamer Pipeline
func (p *Pipeline) Stop() {

	p.pipeline.SetState(gst_.StateNull)
}

func pullSample(p *Pipeline) {

	appsink := p.pipeline.GetByName("appsink")

	for {

		sample, err := appsink.PullSample()
		if err != nil {
			if appsink.IsEOS() == true {
				fmt.Println("eos")
				return
			} else {
				fmt.Println(err)
				return
			}
		}

		samples := uint32(p.clockRate * (float32(sample.Duration) / 1000000000))
		for _, t := range p.tracks {
			if err := t.WriteSample(media.Sample{Data: sample.Data, Samples: samples}); err != nil {
				panic(err)
			}
		}
	}
}
