package utils

import (
	"github.com/pion/rtcp"
)

func NackPairs(seqNums []uint16) []rtcp.NackPair {
	pairs := make([]rtcp.NackPair, 0)
	startSeq := seqNums[0]
	nackPair := &rtcp.NackPair{PacketID: startSeq}
	for i := 1; i < len(seqNums); i++ {
		m := seqNums[i]

		if m-nackPair.PacketID > 16 {
			pairs = append(pairs, *nackPair)
			nackPair = &rtcp.NackPair{PacketID: m}
			continue
		}

		nackPair.LostPackets |= 1 << (m - nackPair.PacketID - 1)
	}

	pairs = append(pairs, *nackPair)

	return pairs
}

func NackParsToSequenceNumbers(pairs []rtcp.NackPair) []uint16 {
	seqs := make([]uint16, 0)
	for _, pair := range pairs {
		startSeq := pair.PacketID
		seqs = append(seqs, startSeq)
		for i := 0; i < 16; i++ {
			if (pair.LostPackets & (1 << i)) != 0 {
				seqs = append(seqs, startSeq+uint16(i)+1)
			}
		}
	}

	return seqs
}
