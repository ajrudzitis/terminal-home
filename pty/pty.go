package pty

import (
	"github.com/gdamore/tcell/v2"
	"golang.org/x/crypto/ssh"
)

// ptysRequestMsg is a message requesting a pseudo-terminal
// that we may recieve from the client.
type PtyRequestMsg struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

// Pty contains the information used to create a pseudo-terminal
// over ssh. It implements the tcell.Tty interface to allow
// that library to be used with the ssh channel. It is mostly
// a wrapper around the ssh.Channel.
type Pty struct {
	channel    ssh.Channel
	windowSize tcell.WindowSize
	windowCb   func()
}

// TODO: make thread safe

var _ tcell.Tty = &Pty{}

func NewPty(channel ssh.Channel) *Pty {
	return &Pty{
		channel: channel,
	}
}

// Close implements tcell.Tty.
func (p *Pty) Close() error {
	return p.channel.Close()
}

// Read implements tcell.Tty.
func (p *Pty) Read(b []byte) (n int, err error) {
	return p.channel.Read(b)
}

// Write implements tcell.Tty.
func (p *Pty) Write(b []byte) (n int, err error) {
	return p.channel.Write(b)
}

func (p *Pty) UpdateWindow(size tcell.WindowSize) {
	p.windowSize = size
	if p.windowCb != nil {
		p.windowCb()
	}
}

// Drain implements tcell.Tty.
func (p *Pty) Drain() error {
	// TODO: not sure what this should do, so I have
	// it make a write to the channel
	_, err := p.channel.Write([]byte{})
	return err
}

// NotifyResize implements tcell.Tty.
func (p *Pty) NotifyResize(cb func()) {
	p.windowCb = cb
}

// Start implements tcell.Tty.
func (p *Pty) Start() error {
	// there is nothing to do here, since we already have the channel
	return nil
}

// Stop implements tcell.Tty.
func (p *Pty) Stop() error {
	return nil
}

// WindowSize implements tcell.Tty.
func (p *Pty) WindowSize() (tcell.WindowSize, error) {
	return p.windowSize, nil
}
