package main

import (
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"math"
	"sync"
	"time"

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

const MAX_SCROLL_LINES = 10000
const MAX_VIEW_LINES = 512

const defaultWindowTitle = "GoMud-Client"
const defaultServer = "127.0.0.1:7778"
const VersionString = "Pre-Alpha build, v0.0.02 07052021937p"

const defaultHorizontalSpaceRatio = 1.5
const DefaultVerticalSpace = 2.0

const defaultWindowWidth = 960
const defaultWindowHeight = 540
const defaultUserScale = 1.0

const defaultRepeatInterval = 3
const defaultRepeatDelay = 30

var renderLock sync.Mutex

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

	lines TextHistory

	dirty bool
}

type TextHistory struct {
	lines    [MAX_SCROLL_LINES]string
	pixLines [MAX_SCROLL_LINES]*ebiten.Image
	rendered [MAX_SCROLL_LINES]bool
	pos      int
	head     int
	tail     int
}

type FontData struct {
	vertSpace  int
	charWidth  int
	charHeight int
	size       int
	data       []byte
	face       font.Face

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

func renderText() {
	//LOCK
	renderLock.Lock()
	//LOCK

	head := mainWin.lines.head
	tail := mainWin.lines.tail
	for a := tail; a <= head && a >= tail; a++ {
		if mainWin.lines.rendered[a] == false {
			go func(a int) {
				mainWin.lines.pixLines[a] = renderLine(a)
			}(a)
		}

	}

	//UNLOCK
	renderLock.Unlock()
	//UNLOCK
}

func renderLine(pos int) *ebiten.Image {
	if mainWin.realWidth > 0 && mainWin.font.size > 0 {

		tempImg := ebiten.NewImage(mainWin.realWidth, int(mainWin.font.size)+mainWin.font.vertSpace)
		text.Draw(tempImg, mainWin.lines.lines[pos],
			mainWin.font.face,
			mainWin.font.size,
			int(mainWin.font.size)+mainWin.font.vertSpace,
			color.RGBA{0xFF, 0x00, 0x00, 0xFF})

		mainWin.dirty = true
		mainWin.lines.rendered[pos] = true
		fmt.Println("renderLine: good")
		return tempImg
	} else {
		fmt.Println("renderLine: invalid size")
	}
	mainWin.lines.rendered[pos] = false
	return nil
}

func (g *Game) Update() error {
	return nil
}

func asyncUpdate() {
	go func() {
		for {
			renderText()
			renderOffscreen()
			time.Sleep(1 * time.Millisecond)
		}
	}()
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
		Size:              float64(mainWin.font.size),
		Hinting:           font.HintingFull,
		GlyphCacheEntries: glyphCacheSize,
	})

	//Font setup
	mainWin.font.vertSpace = mainWin.font.size / DefaultVerticalSpace
	mainWin.font.charWidth = int(math.Round(float64(mainWin.font.size) / float64(defaultHorizontalSpaceRatio)))
	mainWin.font.charHeight = int(math.Round(float64(mainWin.font.size) + float64(mainWin.font.vertSpace)))

	mainWin.lines.lines[0] = "This"
	mainWin.lines.lines[1] = "is"
	mainWin.lines.lines[2] = "a"
	mainWin.lines.lines[3] = "test"
	mainWin.lines.pos = 0
	mainWin.lines.head = 3
	mainWin.lines.tail = 0

	mainWin.offScreen = ebiten.NewImage(mainWin.width, mainWin.height)
	mainWin.dirty = false

	ebiten.SetWindowTitle(mainWin.title)
	ebiten.SetWindowSize(mainWin.width, mainWin.height)
	ebiten.SetMaxTPS(60)

	asyncUpdate()
}

func renderOffscreen() {
	if mainWin.dirty {
		//LOCK
		renderLock.Lock()
		//LOCK
		mainWin.dirty = false

		mainWin.offScreen.Clear()
		mainWin.offScreen.Fill(color.RGBA{0x30, 0x00, 0x00, 0xFF})

		//Render our images out here
		head := mainWin.lines.head
		tail := mainWin.lines.tail
		for a := tail; a <= head && a >= tail; a++ {

			if mainWin.lines.rendered[a] == true {
				op := &ebiten.DrawImageOptions{}
				op.Filter = ebiten.FilterNearest
				op.GeoM.Translate(0.0, float64(a*mainWin.font.charHeight))
				mainWin.offScreen.DrawImage(mainWin.lines.pixLines[a], op)
			} else {
				fmt.Println("renderOffsreen: Nothing to draw.")
			}
		}
		//UNLOCK
		renderLock.Unlock()
		//UNLOCK
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
		renderOffscreen()
		fmt.Println("Buffer resized.")
	}

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	screen.DrawImage(mainWin.offScreen, op)
}
