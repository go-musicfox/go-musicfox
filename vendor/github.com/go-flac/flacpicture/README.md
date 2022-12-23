# flacpicture

[![Documentation](https://godoc.org/github.com/go-flac/flacpicture?status.svg)](https://godoc.org/github.com/go-flac/flacpicture)
[![Build Status](https://travis-ci.org/go-flac/flacpicture.svg?branch=master)](https://travis-ci.org/go-flac/flacpicture)

FLAC picture metablock manipulation for [go-flac](https://www.github.com/go-flac/go-flac)

## Examples

The following example adds a jpeg image as front cover to the FLAC metadata. 

```golang
package example

import (
    "github.com/go-flac/flacpicture"
    "github.com/go-flac/go-flac"
)

func addFLACCover(fileName string, imgData []byte) {
	f, err := flac.ParseFile(fileName)
	if err != nil {
		panic(err)
	}
	picture, err := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover, "Front cover", imgData, "image/jpeg")
	if err != nil {
		panic(err)
	}
	picturemeta := picture.Marshal()
	f.Meta = append(f.Meta, &picturemeta)
	f.Save(fileName)
}
```

The following example extracts an existing cover from a FLAC file
```golang
package example

import (
    "github.com/go-flac/flacpicture"
    "github.com/go-flac/go-flac"
)

func extractFLACCover(fileName string) *MetadataBlockPicture {
	f, err := flac.ParseFile(fileName)
	if err != nil {
		panic(err)
	}
    
    var pic *MetadataBlockPicture
	for _, meta := range f.Meta {
		if meta.Type == flac.Picture {
			pic, err = ParseFromMetaDataBlock(*meta)
			if err != nil {
				panic(err)
			}
		}
    }
    return pic
}
```
