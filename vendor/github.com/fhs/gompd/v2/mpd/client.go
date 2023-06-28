// Copyright 2009 The GoMPD Authors. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Package mpd provides the client side interface to MPD (Music Player Daemon).
// The protocol reference can be found at http://www.musicpd.org/doc/protocol/index.html
package mpd

import (
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

// Quote quotes string VALUES in the format understood by MPD.
// See: https://github.com/MusicPlayerDaemon/MPD/blob/master/src/util/Tokenizer.cxx
// NB: this function shouldn't be used on the PROTOCOL LEVEL because it considers single quotes special chars and
// escapes them.
func quote(s string) string {
	// TODO: We are using strings.Builder even tough it's not ideal.
	// When unsafe.{String,Slice}{,Data} is available, we should use buffer+unsafe.
	//  q := make([]byte, 2+2*len(s))
	//  return unsafe.String(unsafe.SliceData(q), len(q))
	// [issue53003]: https://github.com/golang/go/issues/53003
	var q strings.Builder
	q.Grow(2 + 2*len(s))
	q.WriteByte('"')
	for _, c := range []byte(s) {
		// We need to escape single/double quotes and a backslash by prepending them with a '\'
		switch c {
		case '"', '\\', '\'':
			q.WriteByte('\\')
		}
		q.WriteByte(c)
	}
	q.WriteByte('"')
	return q.String()
}

// Quote quotes each string of args in the format understood by MPD.
// See: https://github.com/MusicPlayerDaemon/MPD/blob/master/src/util/Tokenizer.cxx
func quoteArgs(args []string) string {
	quoted := make([]string, len(args))
	for index, arg := range args {
		quoted[index] = quote(arg)
	}
	return strings.Join(quoted, " ")
}

// Client represents a client connection to a MPD server.
type Client struct {
	text    *textproto.Conn
	version string
}

// Error represents an error returned by the MPD server.
// It contains the error number, the index of the causing command in the command list,
// the name of the command in the command list and the error message.
type Error struct {
	Code             ErrorCode
	CommandListIndex int
	CommandName      string
	Message          string
}

// ErrorCode is the error code of a Error.
type ErrorCode int

// ErrorCodes as defined in MPD source (https://www.musicpd.org/doc/api/html/Ack_8hxx_source.html)
// version 0.21.
const (
	ErrorNotList       ErrorCode = 1
	ErrorArg           ErrorCode = 2
	ErrorPassword      ErrorCode = 3
	ErrorPermission    ErrorCode = 4
	ErrorUnknown       ErrorCode = 5
	ErrorNoExist       ErrorCode = 50
	ErrorPlaylistMax   ErrorCode = 51
	ErrorSystem        ErrorCode = 52
	ErrorPlaylistLoad  ErrorCode = 53
	ErrorUpdateAlready ErrorCode = 54
	ErrorPlayerSync    ErrorCode = 55
	ErrorExist         ErrorCode = 56
)

func (e Error) Error() string {
	if e.CommandName != "" {
		return fmt.Sprintf("command '%s' failed: %s", e.CommandName, e.Message)
	}
	return e.Message
}

// Attrs is a set of attributes returned by MPD.
type Attrs map[string]string

// Dial connects to MPD listening on address addr (e.g. "127.0.0.1:6600")
// on network network (e.g. "tcp").
func Dial(network, addr string) (c *Client, err error) {
	text, err := textproto.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	line, err := text.ReadLine()
	if err != nil {
		return nil, err
	}
	if line[0:6] != "OK MPD" {
		return nil, textproto.ProtocolError("no greeting")
	}
	return &Client{text: text, version: line[7:]}, nil
}

// DialAuthenticated connects to MPD listening on address addr (e.g. "127.0.0.1:6600")
// on network network (e.g. "tcp"). It then authenticates with MPD
// using the plaintext password password if it's not empty.
func DialAuthenticated(network, addr, password string) (c *Client, err error) {
	c, err = Dial(network, addr)
	if err == nil && len(password) > 0 {
		err = c.Command("password %s", password).OK()
	}
	return c, err
}

// Version returns the protocol version used as provided during the handshake.
func (c *Client) Version() string {
	return c.version
}

// We are reimplemeting Cmd() and PrintfLine() from textproto here, because
// the original functions append CR-LF to the end of commands. This behavior
// violates the MPD protocol: Commands must be terminated by '\n'.
func (c *Client) cmd(format string, args ...interface{}) (uint, error) {
	id := c.text.Next()
	c.text.StartRequest(id)
	defer c.text.EndRequest(id)
	if err := c.printfLine(format, args...); err != nil {
		return 0, err
	}
	return id, nil
}

func (c *Client) printfLine(format string, args ...interface{}) error {
	fmt.Fprintf(c.text.W, format, args...)
	c.text.W.WriteByte('\n')
	return c.text.W.Flush()
}

// Close terminates the connection with MPD.
func (c *Client) Close() (err error) {
	if c.text != nil {
		c.printfLine("close")
		err = c.text.Close()
		c.text = nil
	}
	return
}

// Ping sends a no-op message to MPD. It's useful for keeping the connection alive.
func (c *Client) Ping() error {
	return c.Command("ping").OK()
}

func (c *Client) readList(key string) (list []string, err error) {
	list = []string{}
	key += ": "
	for {
		line, err := c.readLine()
		if err != nil {
			return nil, err
		}
		if line == "OK" {
			break
		}
		if !strings.HasPrefix(line, key) {
			return nil, textproto.ProtocolError("unexpected: " + line)
		}
		list = append(list, line[len(key):])
	}
	return
}

func (c *Client) readLine() (string, error) {
	line, err := c.text.ReadLine()
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(line, "ACK ") {
		cur := line[4:]
		var code, idx int
		if strings.HasPrefix(cur, "[") {
			sep := strings.Index(cur, "@")
			end := strings.Index(cur, "] ")
			if sep > 0 && end > 0 {
				code, err = strconv.Atoi(cur[1:sep])
				if err != nil {
					return "", err
				}
				idx, err = strconv.Atoi(cur[sep+1 : end])
				if err != nil {
					return "", err
				}
				cur = cur[end+2:]
			}
		}
		var cmd string
		if strings.HasPrefix(cur, "{") {
			if end := strings.Index(cur, "} "); end > 0 {
				cmd = cur[1:end]
				cur = cur[end+2:]
			}
		}
		msg := strings.TrimSpace(cur)
		return "", Error{
			Code:             ErrorCode(code),
			CommandListIndex: idx,
			CommandName:      cmd,
			Message:          msg,
		}
	}
	return line, nil
}

func (c *Client) readBytes(length int) ([]byte, error) {
	// Read the entire chunk of data. ReadFull() makes sure the data length matches the expectation
	data := make([]byte, length)
	if _, err := io.ReadFull(c.text.R, data); err != nil {
		return nil, err
	}

	// Verify there's a linebreak afterwards and skip it
	termByte, err := c.text.R.ReadByte()
	if err != nil {
		return nil, textproto.ProtocolError("failed to read binary data terminator: " + err.Error())
	}
	if termByte != '\n' {
		return nil, textproto.ProtocolError(fmt.Sprintf("wrong binary data terminator: want 0x0a, got %x", termByte))
	}
	return data, nil
}

func (c *Client) readAttrsList(startKey string) (attrs []Attrs, err error) {
	attrs = []Attrs{}
	startKey += ": "
	for {
		line, err := c.readLine()
		if err != nil {
			return nil, err
		}
		if line == "OK" {
			break
		}
		if strings.HasPrefix(line, startKey) { // new entry begins
			attrs = append(attrs, Attrs{})
		}
		if len(attrs) == 0 {
			return nil, textproto.ProtocolError("unexpected: " + line)
		}
		i := strings.Index(line, ": ")
		if i < 0 {
			return nil, textproto.ProtocolError("can't parse line: " + line)
		}
		attrs[len(attrs)-1][line[0:i]] = line[i+2:]
	}
	return attrs, nil
}

func (c *Client) readAttrs(terminator string) (attrs Attrs, err error) {
	attrs = make(Attrs)
	for {
		line, err := c.readLine()
		if err != nil {
			return nil, err
		}
		if line == terminator {
			break
		}
		z := strings.Index(line, ": ")
		if z < 0 {
			return nil, textproto.ProtocolError("can't parse line: " + line)
		}
		key := line[0:z]
		attrs[key] = line[z+2:]
	}
	return
}

func (c *Client) readBinary() ([]byte, int, error) {
	size := -1
	for {
		line, err := c.readLine()
		switch {
		case err != nil:
			return nil, 0, err

		// Check for the size key
		case strings.HasPrefix(line, "size: "):
			if size, err = strconv.Atoi(line[6:]); err != nil {
				return nil, 0, textproto.ProtocolError("failed to parse size: " + err.Error())
			}

		// Check for the binary key
		case strings.HasPrefix(line, "binary: "):
			length := -1
			if length, err = strconv.Atoi(line[8:]); err != nil {
				return nil, 0, textproto.ProtocolError("failed to parse binary: " + err.Error())
			}

			// If no size is given, assume it's equal to the provided data's length
			if size < 0 {
				size = length
			}

			// The binary data must follow the 'binary:' key
			data, err := c.readBytes(length)
			if err != nil {
				return nil, 0, err
			}

			// The binary data must be followed by the "OK" line
			if s, err := c.readLine(); err != nil {
				return nil, 0, err
			} else if s != "OK" {
				return nil, 0, textproto.ProtocolError("expected 'OK', got " + s)
			}
			return data, size, nil

		// No more data. Obviously, no binary data encountered
		case line == "", line == "OK":
			return nil, 0, textproto.ProtocolError("no binary data found in response")
		}
	}
}

// CurrentSong returns information about the current song in the playlist.
func (c *Client) CurrentSong() (Attrs, error) {
	return c.Command("currentsong").Attrs()
}

// Status returns information about the current status of MPD.
func (c *Client) Status() (Attrs, error) {
	return c.Command("status").Attrs()
}

// Stats displays statistics (number of artists, songs, playtime, etc)
func (c *Client) Stats() (Attrs, error) {
	return c.Command("stats").Attrs()
}

func (c *Client) readOKLine(terminator string) (err error) {
	line, err := c.readLine()
	if err != nil {
		return
	}
	if line == terminator {
		return nil
	}
	return textproto.ProtocolError("unexpected response: " + line)
}

func (c *Client) idle(subsystems ...string) ([]string, error) {
	return c.Command("idle %s", Quoted(strings.Join(subsystems, " "))).Strings("changed")
}

func (c *Client) noIdle() (err error) {
	id, err := c.cmd("noidle")
	if err == nil {
		c.text.StartResponse(id)
		c.text.EndResponse(id)
	}
	return
}

//
// Playback control
//

// Next plays next song in the playlist.
func (c *Client) Next() error {
	return c.Command("next").OK()
}

// Pause pauses playback if pause is true; resumes playback otherwise.
func (c *Client) Pause(pause bool) error {
	if pause {
		return c.Command("pause 1").OK()
	}
	return c.Command("pause 0").OK()
}

// Play starts playing the song at playlist position pos. If pos is negative,
// start playing at the current position in the playlist.
func (c *Client) Play(pos int) error {
	if pos < 0 {
		return c.Command("play").OK()
	}
	return c.Command("play %d", pos).OK()
}

// PlayID plays the song identified by id. If id is negative, start playing
// at the current position in playlist.
func (c *Client) PlayID(id int) error {
	if id < 0 {
		return c.Command("playid").OK()
	}
	return c.Command("playid %d", id).OK()
}

// Previous plays previous song in the playlist.
func (c *Client) Previous() error {
	return c.Command("previous").OK()
}

// Seek seeks to the position time (in seconds) of the song at playlist position pos.
// Deprecated: Use SeekPos instead.
func (c *Client) Seek(pos, time int) error {
	return c.Command("seek %d %d", pos, time).OK()
}

// SeekID is identical to Seek except the song is identified by it's id
// (not position in playlist).
// Deprecated: Use SeekSongID instead.
func (c *Client) SeekID(id, time int) error {
	return c.Command("seekid %d %d", id, time).OK()
}

// SeekPos seeks to the position d of the song at playlist position pos.
func (c *Client) SeekPos(pos int, d time.Duration) error {
	return c.Command("seek %d %f", pos, d.Seconds()).OK()
}

// SeekSongID seeks to the position d of the song identified by id.
func (c *Client) SeekSongID(id int, d time.Duration) error {
	return c.Command("seekid %d %f", id, d.Seconds()).OK()
}

// SeekCur seeks to the position d within the current song.
// If relative is true, then the time is relative to the current playing position.
func (c *Client) SeekCur(d time.Duration, relative bool) error {
	if relative {
		return c.Command("seekcur %+f", d.Seconds()).OK()
	}
	return c.Command("seekcur %f", d.Seconds()).OK()
}

// Stop stops playback.
func (c *Client) Stop() error {
	return c.Command("stop").OK()
}

// SetVolume sets the volume to volume. The range of volume is 0-100.
func (c *Client) SetVolume(volume int) error {
	return c.Command("setvol %d", volume).OK()
}

// Random enables random playback, if random is true, disables it otherwise.
func (c *Client) Random(random bool) error {
	if random {
		return c.Command("random 1").OK()
	}
	return c.Command("random 0").OK()
}

// Repeat enables repeat mode, if repeat is true, disables it otherwise.
func (c *Client) Repeat(repeat bool) error {
	if repeat {
		return c.Command("repeat 1").OK()
	}
	return c.Command("repeat 0").OK()
}

// Single enables single song mode, if single is true, disables it otherwise.
func (c *Client) Single(single bool) error {
	if single {
		return c.Command("single 1").OK()
	}
	return c.Command("single 0").OK()
}

// Consume enables consume mode, if consume is true, disables it otherwise.
func (c *Client) Consume(consume bool) error {
	if consume {
		return c.Command("consume 1").OK()
	}
	return c.Command("consume 0").OK()
}

//
// Playlist related functions
//

// PlaylistInfo returns attributes for songs in the current playlist. If
// both start and end are negative, it does this for all songs in
// playlist. If end is negative but start is positive, it does it for the
// song at position start. If both start and end are positive, it does it
// for positions in range [start, end).
func (c *Client) PlaylistInfo(start, end int) ([]Attrs, error) {
	var cmd *Command
	switch {
	case start < 0 && end < 0:
		// Request all playlist items.
		cmd = c.Command("playlistinfo")
	case start >= 0 && end >= 0:
		// Request this range of playlist items.
		cmd = c.Command("playlistinfo %d:%d", start, end)
	case start >= 0 && end < 0:
		// Request the single playlist item at this position.
		cmd = c.Command("playlistinfo %d", start)
	case start < 0 && end >= 0:
		return nil, errors.New("negative start index")
	default:
		panic("unreachable")
	}
	return cmd.AttrsList("file")
}

// SetPriority set the priority of the specified songs. If end is negative but
// start is non-negative, it does it for the song at position start. If both
// start and end are non-negative, it does it for positions in range
// [start, end).
func (c *Client) SetPriority(priority, start, end int) error {
	switch {
	case start < 0 && end < 0:
		return errors.New("negative start and end index")
	case start >= 0 && end >= 0:
		// Update the prio for this range of playlist items.
		return c.Command("prio %d %d:%d", priority, start, end).OK()
	case start >= 0 && end < 0:
		// Update the prio for a single playlist item at this position.
		return c.Command("prio %d %d", priority, start).OK()
	case start < 0 && end >= 0:
		return errors.New("negative start index")
	default:
		panic("unreachable")
	}
}

// SetPriorityID sets the prio of the song with the given id.
func (c *Client) SetPriorityID(priority, id int) error {
	return c.Command("prioid %d %d", priority, id).OK()
}

// Delete deletes songs from playlist. If both start and end are positive,
// it deletes those at positions in range [start, end). If end is negative,
// it deletes the song at position start.
func (c *Client) Delete(start, end int) error {
	if start < 0 {
		return errors.New("negative start index")
	}
	if end < 0 {
		return c.Command("delete %d", start).OK()
	}
	return c.Command("delete %d:%d", start, end).OK()
}

// DeleteID deletes the song identified by id.
func (c *Client) DeleteID(id int) error {
	return c.Command("deleteid %d", id).OK()
}

// Move moves the songs between the positions start and end to the new position
// position. If end is negative, only the song at position start is moved.
func (c *Client) Move(start, end, position int) error {
	if start < 0 {
		return errors.New("negative start index")
	}
	if end < 0 {
		return c.Command("move %d %d", start, position).OK()
	}
	return c.Command("move %d:%d %d", start, end, position).OK()
}

// MoveID moves songid to position on the plyalist.
func (c *Client) MoveID(songid, position int) error {
	return c.Command("moveid %d %d", songid, position).OK()
}

// Add adds the file/directory uri to playlist. Directories add recursively.
func (c *Client) Add(uri string) error {
	return c.Command("add %s", uri).OK()
}

// AddID adds the file/directory uri to playlist and returns the identity
// id of the song added. If pos is positive, the song is added to position
// pos.
func (c *Client) AddID(uri string, pos int) (int, error) {
	var cmd *Command
	if pos >= 0 {
		cmd = c.Command("addid %s %d", uri, pos)
	} else {
		cmd = c.Command("addid %s", uri)
	}
	attrs, err := cmd.Attrs()
	if err != nil {
		return -1, err
	}
	tok, ok := attrs["Id"]
	if !ok {
		return -1, textproto.ProtocolError("addid did not return Id")
	}
	return strconv.Atoi(tok)
}

// Clear clears the current playlist.
func (c *Client) Clear() error {
	return c.Command("clear").OK()
}

// Shuffle shuffles the tracks from position start to position end in the
// current playlist. If start or end is negative, the whole playlist is
// shuffled.
func (c *Client) Shuffle(start, end int) error {
	if start < 0 || end < 0 {
		return c.Command("shuffle").OK()
	}
	return c.Command("shuffle %d:%d", start, end).OK()
}

// Database related commands

// GetFiles returns the entire list of files in MPD database.
func (c *Client) GetFiles() ([]string, error) {
	return c.Command("list file").Strings("file")
}

// Update updates MPD's database: find new files, remove deleted files, update
// modified files. uri is a particular directory or file to update. If it is an
// empty string, everything is updated.
//
// The returned jobID identifies the update job, enqueued by MPD.
func (c *Client) Update(uri string) (jobID int, err error) {
	id, err := c.cmd("update %s", quote(uri))
	if err != nil {
		return
	}
	c.text.StartResponse(id)
	defer c.text.EndResponse(id)

	line, err := c.readLine()
	if err != nil {
		return
	}
	if !strings.HasPrefix(line, "updating_db: ") {
		return 0, textproto.ProtocolError("unexpected response: " + line)
	}
	jobID, err = strconv.Atoi(line[13:])
	if err != nil {
		return
	}
	return jobID, c.readOKLine("OK")
}

// Rescan updates MPD's database like Update, but it also rescans unmodified
// files. uri is a particular directory or file to update. If it is an empty
// string, everything is updated.
//
// The returned jobID identifies the update job, enqueued by MPD.
func (c *Client) Rescan(uri string) (jobID int, err error) {
	id, err := c.cmd("rescan %s", quote(uri))
	if err != nil {
		return
	}
	c.text.StartResponse(id)
	defer c.text.EndResponse(id)

	line, err := c.readLine()
	if err != nil {
		return
	}
	if !strings.HasPrefix(line, "updating_db: ") {
		return 0, textproto.ProtocolError("unexpected response: " + line)
	}
	jobID, err = strconv.Atoi(line[13:])
	if err != nil {
		return
	}
	return jobID, c.readOKLine("OK")
}

// ListAllInfo returns attributes for songs in the library. Information about
// any song that is either inside or matches the passed in uri is returned.
// To get information about every song in the library, pass in "/".
func (c *Client) ListAllInfo(uri string) ([]Attrs, error) {
	id, err := c.cmd("listallinfo %s ", quote(uri))
	if err != nil {
		return nil, err
	}
	c.text.StartResponse(id)
	defer c.text.EndResponse(id)

	attrs := []Attrs{}
	inEntry := false
	for {
		line, err := c.readLine()
		if err != nil {
			return nil, err
		}
		if line == "OK" {
			break
		} else if strings.HasPrefix(line, "file: ") { // new entry begins
			attrs = append(attrs, Attrs{})
			inEntry = true
		} else if strings.HasPrefix(line, "directory: ") {
			inEntry = false
		}

		if inEntry {
			i := strings.Index(line, ": ")
			if i < 0 {
				return nil, textproto.ProtocolError("can't parse line: " + line)
			}
			attrs[len(attrs)-1][line[0:i]] = line[i+2:]
		}
	}
	return attrs, nil
}

// ListInfo lists the contents of the directory URI using MPD's lsinfo command.
func (c *Client) ListInfo(uri string) ([]Attrs, error) {
	id, err := c.cmd("lsinfo %s", quote(uri))
	if err != nil {
		return nil, err
	}
	c.text.StartResponse(id)
	defer c.text.EndResponse(id)
	attrs := []Attrs{}
	for {
		line, err := c.readLine()
		if err != nil {
			return nil, err
		}
		if line == "OK" {
			break
		}
		if strings.HasPrefix(line, "file: ") ||
			strings.HasPrefix(line, "directory: ") ||
			strings.HasPrefix(line, "playlist: ") {
			attrs = append(attrs, Attrs{})
		}
		i := strings.Index(line, ": ")
		if i < 0 {
			return nil, textproto.ProtocolError("can't parse line: " + line)
		}
		attrs[len(attrs)-1][strings.ToLower(line[0:i])] = line[i+2:]
	}
	return attrs, nil
}

// ReadComments reads "comments" (audio metadata) from the song URI using
// MPD's readcomments command.
func (c *Client) ReadComments(uri string) (Attrs, error) {
	return c.Command("readcomments %s", uri).Attrs()
}

// Find searches the library for songs and returns attributes for each matching song.
// The args are the raw arguments passed to MPD. For example, to search for
// songs that belong to a specific artist and album:
//
//	Find("artist", "Artist Name", "album", "Album Name")
//
// Searches are case sensitive. Use Search for case insensitive search.
func (c *Client) Find(args ...string) ([]Attrs, error) {
	return c.Command("find " + quoteArgs(args)).AttrsList("file")
}

// Search behaves exactly the same as Find, but the searches are not case sensitive.
func (c *Client) Search(args ...string) ([]Attrs, error) {
	return c.Command("search " + quoteArgs(args)).AttrsList("file")
}

// List searches the database for your query. You can use something simple like
// `artist` for your search, or something like `artist album <Album Name>` if
// you want the artist that has an album with a specified album name.
func (c *Client) List(args ...string) ([]string, error) {
	id, err := c.cmd("list " + quoteArgs(args))
	if err != nil {
		return nil, err
	}
	c.text.StartResponse(id)
	defer c.text.EndResponse(id)

	var ret []string
	for {
		line, err := c.readLine()
		if err != nil {
			return nil, err
		}

		i := strings.Index(line, ": ")
		if i > 0 {
			ret = append(ret, line[i+2:])
		} else if line == "OK" {
			break
		} else {
			return nil, textproto.ProtocolError("can't parse line: " + line)
		}
	}
	return ret, nil
}

// Partition commands

// Partition switches the client to a different partition.
func (c *Client) Partition(name string) error {
	return c.Command("partition %s", name).OK()
}

// ListPartitions returns a list of partitions and their information.
func (c *Client) ListPartitions() ([]Attrs, error) {
	return c.Command("listpartitions").AttrsList("partition")
}

// NewPartition creates a new partition with the given name.
func (c *Client) NewPartition(name string) error {
	return c.Command("newpartition %s", name).OK()
}

// DelPartition deletes partition with the given name.
func (c *Client) DelPartition(name string) error {
	return c.Command("delpartition %s", name).OK()
}

// MoveOutput moves an output with the given name to the current partition.
func (c *Client) MoveOutput(name string) error {
	return c.Command("moveoutput %s", name).OK()
}

// Output related commands.

// ListOutputs lists all configured outputs with their name, id & enabled state.
func (c *Client) ListOutputs() ([]Attrs, error) {
	return c.Command("outputs").AttrsList("outputid")
}

// EnableOutput enables the audio output with the given id.
func (c *Client) EnableOutput(id int) error {
	return c.Command("enableoutput %d", id).OK()
}

// DisableOutput disables the audio output with the given id.
func (c *Client) DisableOutput(id int) error {
	return c.Command("disableoutput %d", id).OK()
}

// Stored playlists related commands

// ListPlaylists lists all stored playlists.
func (c *Client) ListPlaylists() ([]Attrs, error) {
	return c.Command("listplaylists").AttrsList("playlist")
}

// PlaylistContents returns a list of attributes for songs in the specified
// stored playlist.
func (c *Client) PlaylistContents(name string) ([]Attrs, error) {
	return c.Command("listplaylistinfo %s", name).AttrsList("file")
}

// PlaylistLoad loads the specfied playlist into the current queue.
// If start and end are non-negative, only songs in this range are loaded.
func (c *Client) PlaylistLoad(name string, start, end int) error {
	if start < 0 || end < 0 {
		return c.Command("load %s", name).OK()
	}
	return c.Command("load %s %d:%d", name, start, end).OK()
}

// PlaylistAdd adds a song identified by uri to a stored playlist identified
// by name.
func (c *Client) PlaylistAdd(name string, uri string) error {
	return c.Command("playlistadd %s %s", name, uri).OK()
}

// PlaylistClear clears the specified playlist.
func (c *Client) PlaylistClear(name string) error {
	return c.Command("playlistclear %s", name).OK()
}

// PlaylistDelete deletes the song at position pos from the specified playlist.
func (c *Client) PlaylistDelete(name string, pos int) error {
	return c.Command("playlistdelete %s %d", name, pos).OK()
}

// PlaylistMove moves a song identified by id in a playlist identified by name
// to the position pos.
func (c *Client) PlaylistMove(name string, id, pos int) error {
	return c.Command("playlistmove %s %d %d", name, id, pos).OK()
}

// PlaylistRename renames the playlist identified by name to newName.
func (c *Client) PlaylistRename(name, newName string) error {
	return c.Command("rename %s %s", name, newName).OK()
}

// PlaylistRemove removes the playlist identified by name from the playlist
// directory.
func (c *Client) PlaylistRemove(name string) error {
	return c.Command("rm %s", name).OK()
}

// PlaylistSave saves the current playlist as name in the playlist directory.
func (c *Client) PlaylistSave(name string) error {
	return c.Command("save %s", name).OK()
}

// A Sticker represents a name/value pair associated to a song. Stickers
// are managed and shared by MPD clients, and MPD server does not assume
// any special meaning in them.
type Sticker struct {
	Name, Value string
}

func newSticker(name, value string) *Sticker {
	return &Sticker{
		Name:  name,
		Value: value,
	}
}

func parseSticker(s string) (*Sticker, error) {
	// Since '=' can appear in the sticker name and in the sticker value,
	// it's impossible to determine where the name ends and value starts.
	// Assume that '=' is more likely to occur in the value
	// (e.g. base64 encoded data -- see #39).
	i := strings.Index(s, "=")
	if i < 0 {
		return nil, textproto.ProtocolError("parsing sticker failed")
	}
	return newSticker(s[:i], s[i+1:]), nil
}

// StickerDelete deletes sticker for the song with given URI.
func (c *Client) StickerDelete(uri string, name string) error {
	return c.Command("sticker delete song %s %s", uri, name).OK()
}

// StickerFind finds songs inside directory with URI which have a sticker with given name.
// It returns a slice of URIs of matching songs and a slice of corresponding stickers.
func (c *Client) StickerFind(uri string, name string) ([]string, []Sticker, error) {
	attrs, err := c.Command("sticker find song %s %s", uri, name).AttrsList("file")
	if err != nil {
		return nil, nil, err
	}
	files := make([]string, len(attrs))
	stks := make([]Sticker, len(attrs))
	for i, attr := range attrs {
		if _, ok := attr["file"]; !ok {
			return nil, nil, textproto.ProtocolError("file attribute not found")
		}
		if _, ok := attr["sticker"]; !ok {
			return nil, nil, textproto.ProtocolError("sticker attribute not found")
		}
		files[i] = attr["file"]
		stk, err := parseSticker(attr["sticker"])
		if err != nil {
			return nil, nil, err
		}
		stks[i] = *stk
	}
	return files, stks, nil
}

// StickerGet gets sticker value for the song with given URI.
func (c *Client) StickerGet(uri string, name string) (*Sticker, error) {
	attrs, err := c.Command("sticker get song %s %s", uri, name).Attrs()
	if err != nil {
		return nil, err
	}
	attr, ok := attrs["sticker"]
	if !ok {
		return nil, textproto.ProtocolError("sticker not found")
	}
	stk, err := parseSticker(attr)
	if stk == nil {
		return nil, err
	}
	return stk, nil
}

// StickerList returns a slice of stickers for the song with given URI.
func (c *Client) StickerList(uri string) ([]Sticker, error) {
	attrs, err := c.Command("sticker list song %s", uri).AttrsList("sticker")
	if err != nil {
		return nil, err
	}
	stks := make([]Sticker, len(attrs))
	for i, attr := range attrs {
		s, ok := attr["sticker"]
		if !ok {
			return nil, textproto.ProtocolError("sticker attribute not found")
		}
		stk, err := parseSticker(s)
		if err != nil {
			return nil, err
		}
		stks[i] = *stk
	}
	return stks, nil
}

// StickerSet sets sticker value for the song with given URI.
func (c *Client) StickerSet(uri string, name string, value string) error {
	return c.Command("sticker set song %s %s %s", uri, name, value).OK()
}

// AlbumArt retrieves an album artwork image for a song with the given URI using MPD's albumart command.
func (c *Client) AlbumArt(uri string) ([]byte, error) {
	offset := 0
	var data []byte
	for {
		// Read the data in chunks
		chunk, size, err := c.Command("albumart %s %d", uri, offset).Binary()
		if err != nil {
			return nil, err
		}

		// Accumulate the data
		data = append(data, chunk...)
		offset = len(data)
		if offset >= size {
			break
		}
	}
	return data, nil
}

// ReadPicture retrieves the embedded album artwork image for a song with the given URI using MPD's readpicture command.
func (c *Client) ReadPicture(uri string) ([]byte, error) {
	offset := 0
	var data []byte
	for {
		// Read the data in chunks
		chunk, size, err := c.Command("readpicture %s %d", uri, offset).Binary()
		if err != nil {
			return nil, err
		}

		// Accumulate the data
		data = append(data, chunk...)
		offset = len(data)
		if offset >= size {
			break
		}
	}
	return data, nil
}
