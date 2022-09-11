package meta

import (
	"encoding/binary"
	"io/ioutil"
)

// Application contains third party application specific data.
//
// ref: https://www.xiph.org/flac/format.html#metadata_block_application
type Application struct {
	// Registered application ID.
	//
	// ref: https://www.xiph.org/flac/id.html
	ID uint32
	// Application data.
	Data []byte
}

// parseApplication reads and parses the body of an Application metadata block.
func (block *Block) parseApplication() error {
	// 32 bits: ID.
	app := new(Application)
	block.Body = app
	err := binary.Read(block.lr, binary.BigEndian, &app.ID)
	if err != nil {
		return unexpected(err)
	}

	// Check if the Application block only contains an ID.
	if block.Length == 4 {
		return nil
	}

	// (block length)-4 bytes: Data.
	app.Data, err = ioutil.ReadAll(block.lr)
	return unexpected(err)
}
