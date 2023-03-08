package byteutil

// IsNumChar returns true if the given character is a numeric, otherwise false.
func IsNumChar(c byte) bool { return c >= '0' && c <= '9' }
