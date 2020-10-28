package buffer

import (
	"testing"
)

const (
	maxBufLen = 80
	maxBufAdd = 245
)

type simpleDataPoint struct {
	data float64
}

func (s simpleDataPoint) Value() float64 {
	return s.data
}

func TestNewBuffer(t *testing.T) {
	for i := 0; i < maxBufLen; i++ {
		buf := NewDataBuffer(i)
		if buf.cap != i || len(buf.data) != i || cap(buf.data) != i {
			t.Fatalf("Unexpected buffer length(s): %d, %d, %d", buf.cap, len(buf.data), cap(buf.data))
		}
	}
}

func TestAddToBuffer(t *testing.T) {
	for bufLen := 1; bufLen < maxBufLen; bufLen++ {
		buf := NewDataBuffer(bufLen)

		for bufAdd := 1; bufAdd < maxBufAdd; bufAdd++ {
			buf.Append(simpleDataPoint{data: float64(bufAdd)})

			if buf.Last().Value() != float64(bufAdd) {
				t.Fatalf("Unexpected value after adding element to buffer, want %d, have %.2f", bufAdd, buf.Last().Value())
			}

			if buf.Last().(simpleDataPoint).data != float64(bufAdd) {
				t.Fatalf("Unexpected asserted value after adding element to buffer, want %d, have %.2f", bufAdd, buf.Last().Value())
			}
		}
	}
}

func TestAddToAndRetrieveFromBuffer(t *testing.T) {
	for bufLen := 1; bufLen < maxBufLen; bufLen++ {
		buf := NewDataBuffer(bufLen)

		for bufAdd := 1; bufAdd <= maxBufAdd; bufAdd++ {
			buf.Append(simpleDataPoint{data: float64(bufAdd)})

			if buf.Last().Value() != float64(bufAdd) {
				t.Fatalf("Unexpected value after adding element to buffer, want %d, have %.2f", bufAdd, buf.Last().Value())
			}

			if buf.Last().(simpleDataPoint).data != float64(bufAdd) {
				t.Fatalf("Unexpected asserted value after adding element to buffer, want %d, have %.2f", bufAdd, buf.Last().Value())
			}

			for k := 1; k <= bufLen; k++ {
				lastN := buf.LastN(k)

				if len(lastN) != k {
					t.Fatalf("Unexpected length of buffer extraction, want %d, have %d", k, len(lastN))
				}

				for l := 0; l < k; l++ {
					pos := buf.ptr - k + l
					if pos < 0 {
						pos = buf.cap + pos
					}
					if lastN[l] != buf.data[pos] {
						t.Fatalf("Unexpected lth out of k (l = %d, k = %d) data retrieved from buffer for (bufLen=%d, bufAdd=%d), want %v, have %v", l+1, k, bufLen, bufAdd, buf.data[pos], lastN[l])
					}
				}
			}

		}
	}
}

func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || r.(string) != "Cannot retrieve more buffer elements then its capacity" {
			t.Errorf("The code did not panic as expected: %v", r)
		}
	}()

	buf := NewDataBuffer(1)
	buf.LastN(2)
}
