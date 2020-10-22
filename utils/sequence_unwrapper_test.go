package utils

import (
	"fmt"
	"math"
	"testing"
)

func TestSequenceUnwrapper(t *testing.T) {
	for _, data := range []struct {
		baseIn      int
		min         int64
		max         int64
		cutOverflow func(int64) uint64
	}{
		{
			baseIn:      16,
			min:         0,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      16,
			min:         1,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      16,
			min:         2,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      16,
			min:         32765,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      16,
			min:         32766,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      16,
			min:         32767,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      16,
			min:         32768,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      16,
			min:         32769,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      16,
			min:         32770,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      16,
			min:         32771,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      16,
			min:         65529,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      16,
			min:         65530,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      16,
			min:         65531,
			max:         131100,
			cutOverflow: func(i int64) uint64 { return uint64(uint16(i)) },
		},
		{
			baseIn:      32,
			min:         4294967290,
			max:         4294967310,
			cutOverflow: func(i int64) uint64 { return uint64(uint32(i)) },
		},
		{
			baseIn:      32,
			min:         4294967291,
			max:         4294967310,
			cutOverflow: func(i int64) uint64 { return uint64(uint32(i)) },
		},
		{
			baseIn:      32,
			min:         4294967292,
			max:         4294967310,
			cutOverflow: func(i int64) uint64 { return uint64(uint32(i)) },
		},
	} {
		data := data
		t.Run(fmt.Sprintf("%d-%d", data.min, data.max), func(t *testing.T) {
			seqNorm := NewSequenceUnwrapper(data.baseIn)
			for want := data.min; want < data.max; {
				got := seqNorm.Unwrap(data.cutOverflow(want))
				if want != got {
					t.Fatalf("step: -1, want: %d, got: %d", want, got)
				}

				want += 3
				got = seqNorm.Unwrap(data.cutOverflow(want))
				if want != got {
					t.Fatalf("step: +3, want: %d, got: %d", want, got)
				}

				want--
			}
		})
	}
}

func TestSequenceUnwrapperBackConvert(t *testing.T) {
	seqNorm := NewSequenceUnwrapper(16)
	for i := 0; i < 3; i++ {
		for ui := uint16(0); ; ui++ {
			unwrapped := seqNorm.Unwrap(uint64(ui))

			if ui != uint16(unwrapped) {
				t.Fatalf("%d != %d (%d)", ui, uint16(unwrapped), unwrapped)
			}

			if ui == math.MaxUint16 {
				break
			}
		}
	}
}
