package utils

import "sync"

type SequenceUnwrapper struct {
	m                     *sync.Mutex
	wrapArounds           int64
	highestSequenceNumber uint64
	inMax                 uint64
	started               bool
}

func NewSequenceUnwrapper(base int) *SequenceUnwrapper {
	return &SequenceUnwrapper{
		m:     &sync.Mutex{},
		inMax: 1 << uint64(base),
	}
}

// Unwrap accepts a uint<base> values which form a sequence, and converts the values into
// int64 while adjusting the sequence every time a wraparound happens. For example:
// unwrapper := NewSequenceUnwrapper(16)
// ...
// unwrapper.Unwrap(65534) == 65534
// unwrapper.Unwrap(65535) == 65535
// unwrapper.Unwrap(0) == 65536
// unwrapper.Unwrap(1) == 65537
// ...
// unwrapper.Unwrap(65534) == 131070
// unwrapper.Unwrap(65535) == 131071
// unwrapper.Unwrap(0) == 131072
// unwrapper.Unwrap(1) == 131073
// ...
func (sw *SequenceUnwrapper) Unwrap(n uint64) int64 {
	sw.m.Lock()
	defer sw.m.Unlock()

	if !sw.started {
		sw.started = true
		sw.highestSequenceNumber = n
		return int64(n)
	}

	if n == sw.highestSequenceNumber {
		return sw.wrapArounds + int64(n)
	}

	if n < sw.highestSequenceNumber {
		if sw.highestSequenceNumber-n > sw.inMax/2 {
			sw.wrapArounds += int64(sw.inMax)
			sw.highestSequenceNumber = n
			return sw.wrapArounds + int64(n)
		}

		return sw.wrapArounds + int64(n)
	}

	if n-sw.highestSequenceNumber > sw.inMax/2 {
		return sw.wrapArounds - int64(sw.inMax) + int64(n)
	}

	sw.highestSequenceNumber = n
	return sw.wrapArounds + int64(n)
}
