// Copyright 2013 The GoMPD Authors. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package mpd

// Watcher represents a MPD client connection that can be watched for events.
type Watcher struct {
	conn  *Client       // client connection to MPD
	exit  chan bool     // channel used to ask loop to terminate
	done  chan bool     // channel indicating loop has terminated
	names chan []string // channel to set new subsystems to watch
	Event chan string   // event channel
	Error chan error    // error channel
}

// NewWatcher connects to MPD server and watches for changes in subsystems
// names. If no subsystem is specified, all changes are reported.
//
// See http://www.musicpd.org/doc/protocol/command_reference.html#command_idle
// for valid subsystem names.
func NewWatcher(net, addr, passwd string, names ...string) (w *Watcher, err error) {
	conn, err := DialAuthenticated(net, addr, passwd)
	if err != nil {
		return
	}
	w = &Watcher{
		conn:  conn,
		Event: make(chan string),
		Error: make(chan error),
		done:  make(chan bool),
		// Buffer channels to avoid race conditions with noIdle
		names: make(chan []string, 1),
		exit:  make(chan bool, 1),
	}
	go w.watch(names...)
	return
}

func (w *Watcher) watch(names ...string) {
	defer w.closeChans()

	// We can block in two places: idle and sending on Event/Error channels.
	// We need to check w.exit and w.names after each.
	for {
		changed, err := w.conn.idle(names...)
		select {
		case <-w.exit:
			// If Close interrupted idle with a noidle, and we don't
			// exit now, we will block trying to send on Event/Error.
			return
		case names = <-w.names:
			// Received new subsystems to watch. Ignore results.
			changed = []string{}
			err = nil
		default: // continue
		}

		switch {
		case err != nil:
			w.Error <- err
		default:
			for _, name := range changed {
				w.Event <- name
			}
		}
		select {
		case <-w.exit:
			// If Close unblocks us from sending on Event/Error channels,
			// we should exit now because noidle might be sent out
			// before we get to idle.
			return
		case names = <-w.names:
			// If method Subsystems unblocks us from sending on Event/Error
			// channels, the next call to idle should be on the new names.
		default: // continue
		}
	}
}

func (w *Watcher) closeChans() {
	close(w.Event)
	close(w.Error)
	close(w.names)
	close(w.exit)
	close(w.done)
}

func (w *Watcher) consume() {
	for {
		select {
		case <-w.Event:
		case <-w.Error:
		default:
			return
		}
	}
}

// Subsystems interrupts watching current subsystems, consumes all
// outstanding values from Event and Error channels, and then
// changes the subsystems to watch for to names.
func (w *Watcher) Subsystems(names ...string) {
	w.names <- names
	w.consume()
	w.conn.noIdle()
}

// Close closes Event and Error channels, and the connection to MPD server.
func (w *Watcher) Close() error {
	w.exit <- true
	w.consume()
	w.conn.noIdle()

	<-w.done // wait for idle to finish and channels to close
	// At this point, watch goroutine has ended,
	// so it's safe to close connection.
	return w.conn.Close()
}
