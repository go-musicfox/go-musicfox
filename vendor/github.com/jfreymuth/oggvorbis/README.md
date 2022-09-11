# oggvorbis
a native go ogg/vorbis decoder

[![GoDoc](https://godocs.io/github.com/jfreymuth/oggvorbis?status.svg)](https://godocs.io/github.com/jfreymuth/oggvorbis)

## Usage

This package provides the type oggvorbis.Reader, which can be used to read .ogg files.

	r, err := oggvorbis.NewReader(reader)
	// handle error

	fmt.Println(r.SampleRate())
	fmt.Println(r.Channels())

	buffer := make([]float32, 8192)
	for {
		n, err := r.Read(buffer)

		// use buffer[:n]

		if err == io.EOF {
			break
		}
		if err != nil {
			// handle error
		}
	}

The reader also provides methods for seeking, these will only work if the reader
was created from an io.ReadSeeker.

There are also convenience functions to read an entire (small) file, similar to ioutil.ReadAll.

	data, format, err := oggvorbis.ReadAll(reader)
