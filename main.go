package main

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"math"
	"runtime"
	"strings"
	"sync"

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
//go:embed "unispacerg.ttf"
var defaultFont []byte

const glyphCacheSize = 256
const defaultFontSize = 18.0

//default greeting
//go:embed "greet.txt"
var greeting []byte

//Constants
const MAX_INPUT_LENGTH = 1024 * 1024 //Some kind of reasonable limit
const MAX_LINE_LENGTH = 1024 * 10    //Larger than needed, for ANSI codes.

const MAX_SCROLL_LINES = 10000 //Max scrollback
const MAX_VIEW_LINES = 512     //Maximum lines on screen

const defaultWindowTitle = "GoMud-Client"
const defaultServer = "127.0.0.1:7778"
const VersionString = "Pre-Alpha build, v0.0.03 07082021-0111a"

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

	head := mainWin.lines.head
	tail := mainWin.lines.tail

	for a := tail; a <= head && a >= tail; a++ {
		if mainWin.lines.pixLines[a] == nil {
			mainWin.lines.pixLines[a] = renderLine(a)
		}

	}
}

func ansiColor(t string) []support.ANSIData {
	textLen := len(t)
	foundColor := false               //Mark when color code starts
	colorStart := 0                   //Color code start position
	colorEnd := 0                     //Color code end position
	drawColor := support.ANSI_DEFAULT //Var to store the colors

	//Only alloc what we need
	textColors := make([]support.ANSIData, textLen+1)

	for z := 0; z < textLen; z++ { //Loop through all chars
		if t[z] == '\033' {
			foundColor = true                    //Found ANSI escape code
			colorStart = z                       //Record start pos
			textColors[z] = support.ANSI_CONTROL //Mark this as no-draw
			continue
		} else if z-colorStart > 10 { //Bail, this isn't a valid color code
			foundColor = false
		} else if foundColor && t[z] == 'm' { //Color code end
			colorEnd = z
			foundColor = false
			if z+1 < textLen { //Make sure we dont run off the end of the string
				drawColor = support.DecodeANSI(t[colorStart : colorEnd+1])
				textColors[z+1] = drawColor //Set color
			}
			textColors[z] = support.ANSI_CONTROL //Mark code end as so
			continue
		}
		if foundColor {
			textColors[z] = support.ANSI_CONTROL //Not valid
		} else {
			textColors[z] = drawColor //Mark all characters with current color
		}
	}
	return textColors
}

func renderLine(pos int) *ebiten.Image {
	if mainWin.realWidth > 0 && mainWin.font.size > 0 {
		len := len(mainWin.lines.lines[pos])
		tempImg := ebiten.NewImage(mainWin.realWidth, int(math.Round(mainWin.font.size+mainWin.font.vertSpace)))
		for i := 0; i < len; i++ {
			text.Draw(tempImg, string(mainWin.lines.lines[pos][i]),
				mainWin.font.face,
				int(math.Round(float64(i)*mainWin.font.charWidth)),
				int(math.Round(mainWin.font.size)),
				color.RGBA64{mainWin.lines.colors[pos][i].Red, mainWin.lines.colors[pos][i].Green, mainWin.lines.colors[pos][i].Blue, 0xFFFF})
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
	renderOffscreen()
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

	mainWin.offScreen = ebiten.NewImage(mainWin.width, mainWin.height)
	ebiten.SetWindowTitle(mainWin.title)
	ebiten.SetWindowSize(mainWin.width, mainWin.height)

	ebiten.SetWindowResizable(true)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetMaxTPS(60)
	ebiten.SetRunnableOnUnfocused(true)

	updateNow() //Only call when needed
	DialSSL(defaultServer)
	readNet()
}

func renderOffscreen() {

	mainWin.offScreen.Clear()
	mainWin.offScreen.Fill(color.RGBA{0x30, 0x00, 0x00, 0xFF})

	//Render our images out here
	head := mainWin.lines.head
	tail := mainWin.lines.tail
	for a := tail; a <= head && a >= tail; a++ {

		if mainWin.lines.pixLines[a] != nil {
			op := &ebiten.DrawImageOptions{}
			op.Filter = ebiten.FilterNearest
			op.GeoM.Translate(0.0, float64(a)*mainWin.font.charHeight)
			mainWin.offScreen.DrawImage(mainWin.lines.pixLines[a], op)
		}
	}

}

func (g *Game) Draw(screen *ebiten.Image) {

	//Resize, or hidpi detected
	sx, sy := screen.Size()
	if mainWin.realWidth != sx || mainWin.realHeight != sy {
		mainWin.realWidth = sx
		mainWin.realHeight = sy
		mainWin.offScreen = ebiten.NewImage(sx, sy)

		updateNow()
		fmt.Println("Buffer resized.")
	}

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest

	screen.DrawImage(mainWin.offScreen, op)
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
			mainWin.lines.colors[i] = ansiColor(lines[x])
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
		}
	}()
}
