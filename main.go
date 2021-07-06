package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"log"

	_ "github.com/flopp/go-findfont"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
)

//Embeds
//default font
//go:embed "unispacerg.ttf"
var defaultFont []byte

const glyphCacheSize = 256
const defaultFontSize = 18

//default greeting
//go:embed "greet.txt"
var greeting []byte

//Constants
const MAX_INPUT_LENGTH = 1024 * 100
const MAX_LINE_LENGTH = 1024
const MAX_VIEW_LINES = 512

const defaultWindowTitle = "GoMud-Client"
const defaultServer = "127.0.0.1:7778"
const VersionString = "Pre-Alpha build, v0.0.02 07052021937p"

const defaultWindowWidth = 960
const defaultWindowHeight = 540
const defaultUserScale = 1.0

const defaultRepeatInterval = 3
const defaultRepeatDelay = 30

const MAX_SCROLL_LINES = 10000

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

	font FontData

	repeatDelay    int
	repeatInterval int

	scrollBack ScrollBack
	viewport   ViewPort

	viewPortStrings [MAX_VIEW_LINES]string
	viewPortPixels  [MAX_VIEW_LINES]string

	dirty bool
}

type ScrollBack struct {
	lines [MAX_SCROLL_LINES]string
	pos   int
	head  int
	tail  int
}

type ViewPort struct {
}

type FontData struct {
	vertSpace int
	charWidth int
	size      float64
	data      []byte
	face      font.Face

	dirty bool
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

	RenderOffscreen()
	return nil
}

// repeatingKeyPressed return true when key is pressed considering the repeat state.
func repeatingKeyPressed(key ebiten.Key) bool {
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= mainWin.repeatDelay && (d-mainWin.repeatDelay)%mainWin.repeatInterval == 0 {
		return true
	}
	return false
}

func init() {
	var err error

	//Load font
	tt, err = truetype.Parse(defaultFont)
	if err != nil {
		log.Fatal(err)
	}

	mainWin.serverAddr = defaultServer
	mainWin.isConnected = false
	mainWin.title = defaultWindowTitle
	mainWin.width = defaultWindowWidth
	mainWin.height = defaultWindowHeight
	mainWin.userScale = defaultUserScale

	//Init font
	mainWin.font.size = defaultFontSize
	mainWin.font.face = truetype.NewFace(tt, &truetype.Options{
		Size:              mainWin.font.size,
		Hinting:           font.HintingFull,
		GlyphCacheEntries: glyphCacheSize,
	})

	mainWin.scrollBack.lines[0] = ""
	mainWin.scrollBack.pos = 0
	mainWin.scrollBack.head = 0
	mainWin.scrollBack.tail = 0

	mainWin.offScreen = ebiten.NewImage(mainWin.width, mainWin.height)
	mainWin.dirty = true

	ebiten.SetWindowTitle(mainWin.title)
	ebiten.SetWindowSize(mainWin.width, mainWin.height)
	ebiten.SetMaxTPS(60)
}

func ScrollbackToViewport() {

}

func PreRenderLines() {

}

func RenderOffscreen() {
	if mainWin.dirty {
		mainWin.dirty = false

		mainWin.offScreen.Clear()
		mainWin.offScreen.Fill(color.RGBA{0x30, 0x00, 0x00, 0xFF})

		charColor := color.RGBA{0xFF, 0x00, 0x00, 0xFF}
		text.Draw(mainWin.offScreen, "testing",
			mainWin.font.face,
			50,
			50,
			charColor)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {

	sx, sy := screen.Size()
	if mainWin.realWidth != sx || mainWin.realHeight != sy {
		mainWin.realWidth = sx
		mainWin.realHeight = sy
		mainWin.offScreen = ebiten.NewImage(sx, sy)

		//Re-render
		mainWin.dirty = true
		RenderOffscreen()
		fmt.Println("Buffer resized.")
	}

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	screen.DrawImage(mainWin.offScreen, op)
}
