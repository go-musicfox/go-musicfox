[![version: golang-version](https://img.shields.io/badge/Go-v1.17-blue.svg)](https://git.exness.io/notifications/api/-/blob/master/go.mod)
[![Build Status](https://travis-ci.org/frolovo22/tag.svg?branch=master)](https://travis-ci.org/frolovo22/tag)
[![Go Report Card](https://goreportcard.com/badge/github.com/frolovo22/tag)](https://goreportcard.com/report/github.com/frolovo22/tag)
[![codecov](https://codecov.io/gh/frolovo22/tag/branch/master/graph/badge.svg)](https://codecov.io/gh/frolovo22/tag)

# Tag

Its pure golang library for parsing and editing tags in mp3, mp4 and flac formats

# Install

```go 
go get github.com/frolovo22/tag
```

For using command line arguments

```go 
go install github.com/frolovo22/tag
```

# Supported tags

| Name              | ID3v1       | ID3v2.2 | ID3v2.3               | ID3v2.4               | MP4             | FLAC                   |
|-------------------|-------------|---------|-----------------------|-----------------------|-----------------|------------------------|
| Title             | Title       | TT2     | TIT2                  | TIT2                  | \xa9nam         | TITLE                  |       
| Artist            | Artist      | TP1     | TPE1                  | TPE1                  | \xa9art         | ARTIST                 |
| Album             | Album       | TAL     | TALB                  | TALB                  | \xa9alb         | ALBUM                  |
| Year              | Year        | TYE     | TYER                  | TDOR                  | \xa9day         | YEAR                   |
| Comment           | Comment     | COM     | COMM                  | COMM                  |                 | COMMENT                |
| Genre             | Genre       | TCO     | TCON                  | TCON                  | \xa9gen         | GENRE                  |
| Album Artist      | -           | TOA     | TPE2                  | TPE2                  | aART            | ALBUMARTIST            | 
| Date              | -           | TIM     | TYER                  | TDRC                  |                 | DATE                   |
| Arranger          | -           | -       | IPLS                  | IPLS                  |                 | ARRANGER               |
| Author            | -           | TOL     | TOLY                  | TOLY                  |                 | AUTHOR                 |
| BPM               | -           | BPM     | TBPM                  | TBPM                  |                 | BPM                    |
| Catalog Number    | -           | -       | TXXX:CATALOGNUMBER    | TXXX:CATALOGNUMBER    |                 | CATALOGNUMBER          |
| Compilation       | -           | -       | TCMP                  | TCMP                  |                 | COMPILATION            |
| Composer          | -           | TCM     | TCOM                  | TCOM                  | \xa9wrt         | COMPOSER               |
| Conductor         | -           | TP3     | TPE3                  | TPE3                  |                 | CONDUCTOR              |
| Copyright         | -           | TCR     | TCOP                  | TCOP                  | cprt            | COPYRIGHT              |
| Description       | -           | TXX     | TIT3                  | TIT3                  |                 | DESCRIPTION            |
| Disc Number       | -           | -       | TPOS                  | TPOS                  |                 | DISCNUMBER             |
| Encoded by        | -           | TEN     | TENC                  | TENC                  | \xa9too         | ENCODED-BY             |
| Track Number      | TrackNumber | TRK     | TRCK                  | TRCK                  | trkn            | TRACKNUMBER            |  
| Picture           | -           | PIC     | APIC                  | APIC                  | covr            | METADATA_BLOCK_PICTURE |

# Status

In progress  
Future features:

* Support all tags (id3 v1, v1.1, v2.2, v2.3, v2.4, mp4, flac)
* Fix errors in files (empty tags, incorrect size, tag size, tag parameters)
* Command line arguments

| Format | Read                      | Set                       | Delete                     | Save                      |
|--------|---------------------------|---------------------------|----------------------------|---------------------------|
| idv1   | <ul><li> - [x] </li></ul> | <ul><li> - [x] </li></ul> | <ul><li> - [x] </li></ul>  | <ul><li> - [x] </li></ul> |
| idv1.1 | <ul><li> - [x] </li></ul> | <ul><li> - [x] </li></ul> | <ul><li> - [x] </li></ul>  | <ul><li> - [x] </li></ul> |
| idv2.2 | <ul><li> - [x] </li></ul> | <ul><li> - [x] </li></ul> | <ul><li> - [x] </li></ul>  | <ul><li> - [ ] </li></ul> |
| idv2.3 | <ul><li> - [x] </li></ul> | <ul><li> - [x] </li></ul> | <ul><li> - [x] </li></ul>  | <ul><li> - [x] </li></ul> |
| idv2.4 | <ul><li> - [x] </li></ul> | <ul><li> - [x] </li></ul> | <ul><li> - [x] </li></ul>  | <ul><li> - [x] </li></ul> |
| mp4    | <ul><li> - [x] </li></ul> | <ul><li> - [ ] </li></ul> | <ul><li> - [ ] </li></ul>  | <ul><li> - [ ] </li></ul> |
| FLAC   | <ul><li> - [x] </li></ul> | <ul><li> - [x] </li></ul> | <ul><li> - [x] </li></ul>  | <ul><li> - [x] </li></ul> |

# Command line arguments

Cli info

```bash
tag help
```

For read tags use

```bash
tag read -in "path/to/file"
# example output
version: id3v2.4
description         : subtitle
track number        : 12/12
author              : Kitten
catalog number      : catalogcat
bmp                 : 777
conductor           : catconductor
copyright           : 2019
album               : CatAlbum
comment             : catcomment
```

for save meta information use

```bash
tag read -in "path/to/file" -out "path/to/outputfile.json"
```

Now supported only json output file

# How to use

```go
tags, err := tag.ReadFile("song.mp3")
if err != nil {
return err
}
fmt.Println(tags.GetTitle())
```

```tag.ReadFile or tag.Read``` return interface ```Metadata```:

```go
package tag

type Metadata interface {
	GetMetadata
	SetMetadata
	DeleteMetadata
	SaveMetadata
}

type GetMetadata interface {
	GetAllTagNames() []string
	GetVersion() Version
	GetFileData() []byte // all another file data

	GetTitle() (string, error)
	GetArtist() (string, error)
	GetAlbum() (string, error)
	GetYear() (int, error)
	GetComment() (string, error)
	GetGenre() (string, error)
	GetAlbumArtist() (string, error)
	GetDate() (time.Time, error)
	GetArranger() (string, error)
	GetAuthor() (string, error)
	GetBMP() (int, error)
	GetCatalogNumber() (int, error)
	GetCompilation() (string, error)
	GetComposer() (string, error)
	GetConductor() (string, error)
	GetCopyright() (string, error)
	GetDescription() (string, error)
	GetDiscNumber() (int, int, error) // number, total
	GetEncodedBy() (string, error)
	GetTrackNumber() (int, int, error) // number, total
	GetPicture() (image.Image, error)
}

type SetMetadata interface {
	SetTitle(title string) error
	SetArtist(artist string) error
	SetAlbum(album string) error
	SetYear(year int) error
	SetComment(comment string) error
	SetGenre(genre string) error
	SetAlbumArtist(albumArtist string) error
	SetDate(date time.Time) error
	SetArranger(arranger string) error
	SetAuthor(author string) error
	SetBMP(bmp int) error
	SetCatalogNumber(catalogNumber int) error
	SetCompilation(compilation string) error
	SetComposer(composer string) error
	SetConductor(conductor string) error
	SetCopyright(copyright string) error
	SetDescription(description string) error
	SetDiscNumber(number int, total int) error
	SetEncodedBy(encodedBy string) error
	SetTrackNumber(number int, total int) error
	SetPicture(picture image.Image) error
}

type DeleteMetadata interface {
	DeleteAll() error

	DeleteTitle() error
	DeleteArtist() error
	DeleteAlbum() error
	DeleteYear() error
	DeleteComment() error
	DeleteGenre() error
	DeleteAlbumArtist() error
	DeleteDate() error
	DeleteArranger() error
	DeleteAuthor() error
	DeleteBMP() error
	DeleteCatalogNumber() error
	DeleteCompilation() error
	DeleteComposer() error
	DeleteConductor() error
	DeleteCopyright() error
	DeleteDescription() error
	DeleteDiscNumber() error
	DeleteEncodedBy() error
	DeleteTrackNumber() error
	DeletePicture() error
}

type SaveMetadata interface {
	SaveFile(path string) error
	Save(input io.WriteSeeker) error
}
```   

Also you can read defined format. For Example:

```go
package main

import (
	"fmt"
	"github.com/frolovo22/tag"
	"os"
)

func Read() error {
	file, err := os.Open("path/to/file")
	if err != nil {
		return err
	}
	defer file.Close()

	id3v2, err := tag.ReadID3v24(file)
	if err != nil {
		return err
	}

	// Get tag value by name
	value, err := id3v2.GetString("TIT2")
	if err != nil {
		return err
	}
	fmt.Println("title: " + value)

	// Set tag value
	err = id3v2.SetString("TIT2", "Title")
	if err != nil {
		return err
	}

	// User defined tags
	value, err = id3v2.GetStringTXXX("MYTAG")
	if err != nil {
		return err
	}
	fmt.Println("my tag: " + value)

	// Set user tag
	err = id3v2.SetStringTXXX("MYTAG222", "Dogs")
	if err != nil {
		return err
	}

	// Save changes
	err = id3v2.SaveFile("path/to/file")
	if err != nil {
		return err
	}
}
``` 

# Contribution

