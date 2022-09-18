// Copyright 2018 The GoMPD Authors. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package mpd

import "fmt"

// Quoted is a string that do no need to be quoted.
type Quoted string

// Command returns a command that can be sent to MPD sever.
// It enables low-level access to MPD protocol and should be avoided if
// the user is not familiar with MPD protocol.
//
// Strings in args are automatically quoted so that spaces are preserved.
// Pass strings as Quoted type if this is not desired.
func (c *Client) Command(format string, args ...interface{}) *Command {
	for i := range args {
		switch s := args[i].(type) {
		case Quoted: // ignore
		case string:
			args[i] = quote(s)
		}
	}
	return &Command{
		client: c,
		cmd:    fmt.Sprintf(format, args...),
	}
}

// A Command represents a MPD command.
type Command struct {
	client *Client
	cmd    string
}

// String returns the encoded command.
func (cmd *Command) String() string {
	return cmd.cmd
}

// OK sends command to server and checks for error.
func (cmd *Command) OK() error {
	id, err := cmd.client.cmd("%v", cmd.cmd)
	if err != nil {
		return err
	}
	cmd.client.text.StartResponse(id)
	defer cmd.client.text.EndResponse(id)
	return cmd.client.readOKLine("OK")
}

// Attrs sends command to server and reads attributes returned in response.
func (cmd *Command) Attrs() (Attrs, error) {
	id, err := cmd.client.cmd(cmd.cmd)
	if err != nil {
		return nil, err
	}
	cmd.client.text.StartResponse(id)
	defer cmd.client.text.EndResponse(id)
	return cmd.client.readAttrs("OK")
}

// AttrsList sends command to server and reads a list of attributes returned in response.
// Each attribute group starts with key startKey.
func (cmd *Command) AttrsList(startKey string) ([]Attrs, error) {
	id, err := cmd.client.cmd(cmd.cmd)
	if err != nil {
		return nil, err
	}
	cmd.client.text.StartResponse(id)
	defer cmd.client.text.EndResponse(id)
	return cmd.client.readAttrsList(startKey)
}

// Strings sends command to server and reads a list of strings returned in response.
// Each string have the key key.
func (cmd *Command) Strings(key string) ([]string, error) {
	id, err := cmd.client.cmd(cmd.cmd)
	if err != nil {
		return nil, err
	}
	cmd.client.text.StartResponse(id)
	defer cmd.client.text.EndResponse(id)
	return cmd.client.readList(key)
}

// Binary sends command to server and reads its binary response, returning the data and its total size (which can be
// greater than the returned chunk).
func (cmd *Command) Binary() ([]byte, int, error) {
	id, err := cmd.client.cmd(cmd.cmd)
	if err != nil {
		return nil, 0, err
	}
	cmd.client.text.StartResponse(id)
	defer cmd.client.text.EndResponse(id)
	return cmd.client.readBinary()
}
