package ioutilx

// Zero is an io.Reader which always reads zero bytes.
var Zero zero

// zero is an io.Reader which always reads zero bytes.
type zero struct {
}

// Read reads len(b) zero bytes into b. It returns the number of bytes read and
// a nil error value.
func (zero) Read(b []byte) (n int, err error) {
	for i := range b {
		b[i] = 0
	}
	return len(b), nil
}
