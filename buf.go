package brew

import "github.com/fako1024/btscale/pkg/scale"

type DataBuffer struct {
	data scale.DataPoints
	ptr  int
	cap  int
}

func NewDataBuffer(cap int) *DataBuffer {
	return &DataBuffer{
		data: make(scale.DataPoints, cap, cap),
		ptr:  0,
		cap:  cap,
	}
}

func (b *DataBuffer) Append(dataPoint scale.DataPoint) {
	b.data[b.ptr] = dataPoint
	b.ptr = (b.ptr + 1) % b.cap
}

func (b *DataBuffer) Last() scale.DataPoint {
	ptr := b.ptr - 1
	if ptr < 0 {
		ptr = b.cap - 1
	}

	return b.data[ptr]
}

func (b *DataBuffer) LastN(n int) []scale.DataPoint {
	res := make([]scale.DataPoint, n, n)

	ptr := b.ptr - 1
	for i := 0; i < n; i++ {
		pos := ptr - i
		if pos < 0 {
			pos = b.cap + pos
		}
		res[i] = b.data[pos]
	}

	return res
}
