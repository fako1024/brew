package buffer

// DataPoint denotes a generic interface to any data structure returning a float64
type DataPoint interface {

	// Value denotes a moethod to retrieve the current value of the data point
	Value() float64
}

// DataPoints denotes a set of data points
type DataPoints []DataPoint

// DataBuffer denotes a generic data buffer
type DataBuffer struct {
	data DataPoints
	ptr  int
	cap  int
}

// NewDataBuffer instantiates a new buffer of given length
func NewDataBuffer(cap int) *DataBuffer {
	return &DataBuffer{
		data: make(DataPoints, cap, cap),
		ptr:  0,
		cap:  cap,
	}
}

// Append adds and element to the end of the buffer
func (b *DataBuffer) Append(dataPoint DataPoint) {
	b.data[b.ptr] = dataPoint
	b.ptr = (b.ptr + 1) % b.cap
}

// Last retrieves the last / current element from the buffer
func (b *DataBuffer) Last() DataPoint {
	ptr := b.ptr - 1
	if ptr < 0 {
		ptr = b.cap - 1
	}

	return b.data[ptr]
}

// LastN retrieves the last / current n element from the buffer
func (b *DataBuffer) LastN(n int) DataPoints {

	if n > b.cap {
		panic("Cannot retrieve more buffer elements then its capacity")
	}

	res := make(DataPoints, n, n)
	ptr := b.ptr - n
	if ptr < 0 {
		step := copy(res, b.data[b.cap+ptr:])
		copy(res[step:], b.data)
	} else {
		copy(res, b.data[ptr:ptr+n])
	}

	return res
}
