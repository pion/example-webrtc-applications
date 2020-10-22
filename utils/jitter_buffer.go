package utils

import (
	"github.com/pion/rtp"
)

type JitterBuffer struct {
	nextStart int64
	end       int64
	packets   []*rtp.Packet
	size      int64
}

func NewJitterBuffer(size int) *JitterBuffer {
	return &JitterBuffer{
		packets: make([]*rtp.Packet, size),
		size:    int64(size),
	}
}

func (s *JitterBuffer) Add(seq int64, packet *rtp.Packet) bool {
	if s.end-seq >= s.size {
		return false
	}

	if seq <= s.end && s.packets[seq%s.size] != nil {
		return false
	}

	if s.nextStart == 0 {
		s.nextStart = seq
	}

	if seq > s.end {
		if seq-s.end >= s.size {
			s.packets = make([]*rtp.Packet, s.size)
		} else {
			for i := s.end + 1; i < seq; i++ {
				s.packets[i%s.size] = nil
			}
		}
		s.end = seq
	}

	if s.nextStart < s.end-s.size+1 {
		s.nextStart = s.end - s.size + 1
	}

	s.packets[seq%s.size] = packet

	return true
}

func (s *JitterBuffer) NextPackets() []*rtp.Packet {
	if s.nextStart > s.end {
		return nil
	}

	if s.packets[s.nextStart%s.size] == nil {
		return nil
	}

	end := s.end // return until sequence end unless a there is a missing packet
	for i := s.nextStart + 1; i <= s.end; i++ {
		if s.packets[i%s.size] == nil {
			end = i - 1 // missing packet found, return until previous packet
			break
		}
	}

	packets := make([]*rtp.Packet, 0, end-s.nextStart+1)
	for i := s.nextStart; i <= end; i++ {
		packets = append(packets, s.packets[i%s.size])
		s.packets[i%s.size] = nil
	}

	s.nextStart = end + 1

	return packets
}

func (s *JitterBuffer) SetNextPacketsStart(nextPacketsStart int64) {
	if s.nextStart < nextPacketsStart {
		for i := s.nextStart; i < nextPacketsStart; i++ {
			s.packets[i%s.size] = nil
		}
		s.nextStart = nextPacketsStart
	}
}
