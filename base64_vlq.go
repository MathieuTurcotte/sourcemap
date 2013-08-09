package sourcemap

import (
	"io"
)

var decodeMap [256]byte

func init() {
	// Use a custom map to decode base64 encoded data instead of the default
	// go implementation in order to read each character into a single byte
	// instead of decoding everything into a slice of bytes where the values
	// are interleaved over a chunk of 4 bytes.
	base64 := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	for i := 0; i < len(base64); i++ {
		decodeMap[base64[i]] = byte(i)
	}
}

const vqlBaseShift = 5
const vqlBase = 1 << vqlBaseShift // 00100000, i.e. 32
const vqlBaseMask = vqlBase - 1   // 00011111
const vqlContMask = vqlBase       // 00100000

func fromVQLSigned(val int) int {
	signed := (val & 1) == 1
	val = val >> 1
	if signed {
		return -val
	} else {
		return val
	}
}

// Decode the next base 64 VQL value from the reader. An error is returned if
// the byte reader reaches the end of its input while decoding the VQL value.
func decodeVQL(reader io.ByteReader) (result int, err error) {
	continuation := true
	shift := uint(0)

	for continuation {
		b, err := reader.ReadByte()
		if err != nil {
			return -1, err
		}
		b = decodeMap[b]
		continuation = (b & vqlContMask) != 0
		result += int(b&vqlBaseMask) << shift
		// The VLQ base values are arranged most significant first in the
		// stream, so shift left by 5 more bits at each iteration.
		shift += vqlBaseShift
	}

	return fromVQLSigned(result), nil
}
