package bits

// ZigZag decodes a ZigZag encoded integer and returns it.
//
// Examples of ZigZag encoded values on the left and decoded values on the
// right:
//
//    0 =>  0
//    1 => -1
//    2 =>  1
//    3 => -2
//    4 =>  2
//    5 => -3
//    6 =>  3
//
// ref: https://developers.google.com/protocol-buffers/docs/encoding
func ZigZag(x int32) int32 {
	return x>>1 ^ -(x & 1)
}
