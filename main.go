package main

import (
	_ "embed"
	"image/color"
	"log"

	_ "github.com/flopp/go-findfont"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"golang.org/x/image/font"
)

//Embeds
//default font
//go:embed "unispacerg.ttf"
var DefaultFont []byte

//default greeting
//go:embed "greet.txt"
var greeting []byte

//Constants
const MAX_INPUT_LENGTH = 1024 * 100
const MAX_LINE_LENGTH = 1024

const DefaultWindowTitle = "GoMud-Client"
const DefaultServer = "127.0.0.1:7778"
const VersionString = "Pre-Alpha build, v0.0.02 07-04-2021-1055p"

const DefaultWindowWidth = 960
const DefaultWindowHeight = 540
const DefaultUserScale = 1.0

const DefaultRepeatInterval = 3
const DefaultRepeatDelay = 30

const ScrollBackLinesMax = 10000

type Window struct {
	serverAddr  string
	isConnected bool
	offScreen   *ebiten.Image

	title      string
	width      int
	height     int
	realWidth  int
	realHeight int
	userScale  float64

	scrollBack     [ScrollBackLinesMax]string
	scrollBackPos  int
	scrollBackHead int
	scrollBackTail int

	dirty bool
}

type FontData struct {
	VertSpace int
	CharWidth int
	Size      float64
	Data      []byte
	Face      font.Face

	Dirty bool
}

//Globals
var tt *truetype.Font
var mainWin Window

type Game struct {
	//
}

func main() {

	game := &Game{}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {

	return outsideWidth, outsideHeight
}

func (g *Game) Update() error {
	return nil
}

func init() {
	var err error

	//Load font
	tt, err = truetype.Parse(DefaultFont)
	if err != nil {
		log.Fatal(err)
	}

	mainWin.serverAddr = DefaultServer
	mainWin.isConnected = false
	mainWin.title = DefaultWindowTitle
	mainWin.width = DefaultWindowWidth
	mainWin.height = DefaultWindowHeight
	mainWin.userScale = DefaultUserScale

	mainWin.scrollBack[0] = ""
	mainWin.scrollBackPos = 0
	mainWin.scrollBackHead = 0
	mainWin.scrollBackTail = 0

	mainWin.offScreen = ebiten.NewImage(mainWin.width, mainWin.height)
	mainWin.dirty = true

	ebiten.SetWindowTitle(mainWin.title)
	ebiten.SetWindowSize(mainWin.width, mainWin.height)
	ebiten.SetMaxTPS(60)
}

func (g *Game) Draw(screen *ebiten.Image) {
	if mainWin.dirty {
		mainWin.dirty = false

		sx, sy := screen.Size()
		mainWin.realWidth = sx
		mainWin.realHeight = sy

		if mainWin.width != mainWin.realWidth || mainWin.height != mainWin.realHeight {
			mainWin.offScreen = ebiten.NewImage(mainWin.realWidth, mainWin.realHeight)
		}

		mainWin.offScreen.Clear()
		mainWin.offScreen.Fill(color.RGBA{0xFF, 0x00, 0x00, 0xFF})
	}

	op := &ebiten.DrawImageOptions{}
	// Specify linear filter.
	op.Filter = ebiten.FilterLinear

	screen.DrawImage(mainWin.offScreen, op)
}
