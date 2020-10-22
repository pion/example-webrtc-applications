package utils

import (
	"testing"

	"github.com/kr/pretty"
	"github.com/pion/rtcp"
)

func TestNackPairs(t *testing.T) {
	u16 := func(i int) uint16 {
		return uint16(i)
	}

	seqNums := []uint16{
		65533, 65534, u16(65547), u16(65548), u16(65549),
		u16(65550),
		u16(65580), u16(65581),
	}

	wantPairs := []rtcp.NackPair{
		{PacketID: 65533, LostPackets: 0b1110000000000001},
		{PacketID: 14, LostPackets: 0b0000000000000000},
		{PacketID: 44, LostPackets: 0b0000000000000001},
	}

	pairs := NackPairs(seqNums)

	if diff := pretty.Diff(wantPairs, pairs); len(diff) > 0 {
		t.Errorf("want/got: %v", diff)
	}
}

func TestNackParsToSequenceNumbers(t *testing.T) {
	pairs := []rtcp.NackPair{
		{PacketID: 65533, LostPackets: 0b1110000000000001},
		{PacketID: 14, LostPackets: 0b0000000000000000},
		{PacketID: 44, LostPackets: 0b0000000000000001},
	}

	u16 := func(i int) uint16 {
		return uint16(i)
	}

	wantSeqNums := []uint16{
		65533, 65534, u16(65547), u16(65548), u16(65549),
		u16(65550),
		u16(65580), u16(65581),
	}

	seqNums := NackParsToSequenceNumbers(pairs)

	if diff := pretty.Diff(wantSeqNums, seqNums); len(diff) > 0 {
		t.Errorf("want/got: %v", diff)
	}
}
