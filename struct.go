package main

import (
	"crypto/tls"
	"sync"

	"./support"
	"github.com/hajimehoshi/ebiten"
	"golang.org/x/image/font"
)

type Window struct {
	sslCon      *tls.Conn
	serverAddr  string
	isConnected bool

	offScreen *ebiten.Image

	title      string
	width      int
	height     int
	realWidth  int
	realHeight int
	userScale  float64

	font FontData

	repeatDelay    int
	repeatInterval int

	lines TextHistory
	dirty bool
}

type TextHistory struct {
	rawText     string
	rawTextLock sync.Mutex

	lines    [MAX_SCROLL_LINES]string
	colors   [MAX_SCROLL_LINES][]support.ANSIData
	pixLines [MAX_SCROLL_LINES]*ebiten.Image

	pos  int
	head int
	tail int
}

type FontData struct {
	vertSpace  float64
	charWidth  float64
	charHeight float64
	size       float64
	data       []byte
	face       font.Face
}
