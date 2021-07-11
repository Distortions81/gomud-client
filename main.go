package main

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"math"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"./support"
	_ "github.com/flopp/go-findfont"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
)

//Embeds
//default font
//go:embed "data/unispacerg.ttf"
var defaultFont []byte

const glyphCacheSize = 256
const defaultFontSize = 18.0
const clearEveryFrame = true

//Constants
const MAX_INPUT_LENGTH = 100 * 1024 //100kb, some kind of reasonable limit for net/input buffer
const NET_POLL_MS = 66              //1/15th of a second

const MAX_SCROLL_LINES = 10000 //Max scrollback
const MAX_VIEW_LINES = 250     //Maximum lines on screen

const defaultWindowTitle = "GoMud-Client"
const defaultServer = "127.0.0.1:7778"
const VersionString = "Pre-Alpha build, v0.0.031 07092021-1201a"

const defaultHorizontalSpace = 1.4
const defaultVerticalSpace = 4.0

const defaultWindowWidth = 960
const defaultWindowHeight = 540
const defaultUserScale = 1.0

const defaultRepeatInterval = 3
const defaultRepeatDelay = 30

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

//Globals
var tt *truetype.Font
var mainWin Window
var numThreads int = 1
var ebitenLock sync.Mutex

type Game struct {
	counter uint64
}

func main() {

	game := &Game{}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

	game.counter = 0
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {

	return outsideWidth, outsideHeight
}

func renderText() {

	didRender := false

	head := mainWin.lines.head
	tail := mainWin.lines.tail

	//Render old to new, so color codes can persist lines
	for a := tail; a <= head && a >= tail; a++ {
		if mainWin.lines.pixLines[a] == nil {
			mainWin.lines.pixLines[a] = renderLine(a)
			didRender = true
		}

	}

	//TODO: Optimize, only render if a line rendered falls within our viewport.
	//We only render if there is something new to draw!
	if didRender {
		renderOffscreen()
	}
}

func renderLine(pos int) *ebiten.Image {
	if mainWin.realWidth > 0 && mainWin.font.size > 0 {
		len := len(mainWin.lines.lines[pos])
		ebitenLock.Lock()
		defer ebitenLock.Unlock()

		tempImg := ebiten.NewImage(mainWin.realWidth, int(math.Round(mainWin.font.size+mainWin.font.vertSpace)))
		x := 0
		for i := 0; i < len; i++ {
			if strconv.IsPrint(rune(mainWin.lines.lines[pos][i])) && mainWin.lines.colors[pos][i] != support.ANSI_CONTROL {
				x++
				text.Draw(tempImg, string(mainWin.lines.lines[pos][i]),
					mainWin.font.face,
					int(math.Round(float64(x)*mainWin.font.charWidth)),
					int(math.Round(mainWin.font.size)),
					color.RGBA{mainWin.lines.colors[pos][i].Red, mainWin.lines.colors[pos][i].Green, mainWin.lines.colors[pos][i].Blue, 0xFF})
			}
		}
		return tempImg
	}
	return nil
}

func (g *Game) Update() error {
	return nil
}

func updateNow() {
	textToLines()
	renderText()
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

	numThreads = runtime.NumCPU()
	buf := fmt.Sprintf("%d vCPUs found.", numThreads)
	fmt.Println(buf)

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
	mainWin.font.vertSpace = mainWin.font.size / defaultVerticalSpace
	mainWin.font.charWidth = mainWin.font.size / defaultHorizontalSpace
	mainWin.font.charHeight = mainWin.font.size + mainWin.font.vertSpace

	mainWin.lines.lines[0] = ""
	mainWin.lines.pos = 0
	mainWin.lines.head = 0
	mainWin.lines.tail = 0

	mainWin.dirty = false

	mainWin.offScreen = ebiten.NewImage(int(mainWin.width), int(mainWin.height))
	ebiten.SetWindowTitle(mainWin.title)
	ebiten.SetWindowSize(int(mainWin.width), int(mainWin.height))

	ebiten.SetWindowResizable(true)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetMaxTPS(60)
	ebiten.SetRunnableOnUnfocused(true)
	if clearEveryFrame == true {
		ebiten.SetScreenClearedEveryFrame(true)
	} else {
		ebiten.SetScreenClearedEveryFrame(false)
	}

	addLine(
		"GOMud-Client " + VersionString + "\n" +
			"COPYRIGHT 2020-2021 Carl Frank Otto III (carlotto81@gmail.com)\n" +
			"License: Mozilla Public License 2.0\n" +
			"Written in Go, using Ebiten library.\n" +
			"This information must remain unmodified, fully intact and shown to end-users.\n" +
			"Source: https://github.com/Distortions81/gomud-client\n" +
			"\n")
	updateNow() //Only call when needed
	DialSSL(defaultServer)
	readNet()
}

func renderOffscreen() {

	if mainWin.dirty == false { //Check for no pending frames, or we will re-render for no reason
		ebitenLock.Lock()
		defer ebitenLock.Unlock()

		mainWin.offScreen.Clear()
		//mainWin.offScreen.Fill(color.RGBA{0x30, 0x00, 0x00, 0xFF})

		//Render our images out here
		head := mainWin.lines.head
		tail := mainWin.lines.tail
		for a := tail; a <= head && a >= tail; a++ {

			if mainWin.lines.pixLines[a] != nil {
				op := &ebiten.DrawImageOptions{}
				op.Filter = ebiten.FilterNearest
				op.GeoM.Translate(0.0, float64(a)*mainWin.font.charHeight)
				mainWin.offScreen.DrawImage(mainWin.lines.pixLines[a], op)
			} else {
				//Stop rendering here, lines after this are not yet ready.
				return
			}
		}
		mainWin.dirty = true
	}

}

func (g *Game) Draw(screen *ebiten.Image) {

	//Resize, or hidpi detected
	sx, sy := screen.Size()
	if mainWin.realWidth != sx || mainWin.realHeight != sy {
		mainWin.realWidth = sx
		mainWin.realHeight = sy
		mainWin.offScreen = ebiten.NewImage(sx, sy)

		mainWin.dirty = false
		for x := 0; x < MAX_SCROLL_LINES; x++ {
			mainWin.lines.pixLines[x] = nil
		}
		fmt.Println("Buffer resized.")
		updateNow()
	}

	if mainWin.dirty == true || clearEveryFrame {
		mainWin.dirty = false
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		screen.DrawImage(mainWin.offScreen, op)
	}
}

func DialSSL(addr string) {

	buf := fmt.Sprintf("Connecting to: %s\r\n", addr)
	addLine(buf)

	//Todo, allow timeout adjustment and connection canceling.
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", addr, conf)
	if err != nil {
		log.Println(err)

		buf := fmt.Sprintf("%s\r\n", err)
		addLine(buf)
		return
	}
	mainWin.sslCon = conn
}

func addLine(text string) {

	go func() {
		mainWin.lines.rawTextLock.Lock()
		mainWin.lines.rawText += text
		mainWin.lines.rawTextLock.Unlock()
	}()
}

func textToLines() {
	mainWin.lines.rawTextLock.Lock()
	lines := strings.Split(mainWin.lines.rawText, "\n")
	mainWin.lines.rawTextLock.Unlock()

	numLines := len(lines) - 1
	if numLines > 0 {

		x := 0
		for i := mainWin.lines.head + 1; i < MAX_SCROLL_LINES && x <= numLines; i++ {
			mainWin.lines.lines[i] = lines[x]
			mainWin.lines.colors[i] = support.AnsiColor(lines[x])
			x++
		}
		mainWin.lines.head += x
		mainWin.lines.rawText = ""
		updateNow()
	}
}

func readNet() {
	go func() {
		for {
			buf := make([]byte, MAX_INPUT_LENGTH)
			if mainWin.sslCon != nil {
				n, err := mainWin.sslCon.Read(buf)
				if err != nil {
					log.Println(n, err)
					mainWin.sslCon.Close()

					buf := fmt.Sprintf("Lost connection to %s: %s\r\n", mainWin.serverAddr, err)
					addLine(buf)
					mainWin.sslCon = nil
				}
				newData := string(buf[:n])
				addLine(newData)
			}
			time.Sleep(time.Millisecond * NET_POLL_MS)
		}
	}()
}
