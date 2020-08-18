package main

import (
	"crypto/tls"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"

	"./support"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
)

const VersionString = "Pre-Alpha build, v0.0.01 08162020-0909"
const MAX_STRING_LENGTH = 1024 * 10

//Greeting
const DefaultGreetFile = "greet.txt"

//Window
const DefaultWindowWidth = 640.0
const DefaultWindowHeight = 360.0
const DefaultWindowDPI = 72
const DefaultMagnification = 2.25

const DefaultRepeatInterval = 3
const DefaultRepeatDelay = 60

const DefaultWindowTitle = "GoMud-Client"

type Window struct {
	Title  string
	Width  int
	Height int
	DPI    int

	TextLines   int
	TextColumns int

	Scroll ScrollData
	Font   FontData

	RepeatInterval int
	RepeatDelay    int

	Text       string
	Tick       int
	ScrollBack string
	InputLine  string

	Con *tls.Conn
}

//Scroll
type ScrollData struct {
	ScrollPos int
}

//Font
const DefaultFontFile = "unispacerg.ttf"
const glyphCacheSize = 512
const DefaultVerticalSpace = 3.0
const HorizontalSpaceRatio = 1.66666666667
const DefaultFontSize = 12.0
const LeftMargin = 4.0

type FontData struct {
	VerticalSpace int
	Size          int
	Data          []byte
	Face          font.Face

	Color string
	Dirty bool
}

//Data
var MainWin Window
var ActiveWin *Window

type Game struct {
	counter int
}

//No ASCII control characters
func StripControl(str string) string {
	b := make([]byte, len(str))
	var bl int
	for i := 0; i < len(str); i++ {
		c := str[i]
		if (c >= 32 && c < 255) || c == '\n' {
			b[bl] = c
			bl++
		}
	}
	return string(b[:bl])
}

// repeatingKeyPressed return true when key is pressed considering the repeat state.
func repeatingKeyPressed(key ebiten.Key) bool {
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= ActiveWin.RepeatDelay && (d-ActiveWin.RepeatDelay)%ActiveWin.RepeatInterval == 0 {
		return true
	}
	return false
}

func ReadInput() {
	for {
		buf := make([]byte, MAX_STRING_LENGTH)
		n, err := ActiveWin.Con.Read(buf)
		if err != nil {
			log.Println(n, err)
			ActiveWin.Con.Close()
			os.Exit(0)
		}
		ActiveWin.ScrollBack += string(buf[:n])
	}
}

func (g *Game) Update(screen *ebiten.Image) error {

	//Add linebreaks
	if repeatingKeyPressed(ebiten.KeyEnter) || repeatingKeyPressed(ebiten.KeyKPEnter) {
		ActiveWin.InputLine += "\n"
	}

	//Backspace
	if repeatingKeyPressed(ebiten.KeyBackspace) {
		if len(ActiveWin.InputLine) >= 1 {
			ActiveWin.InputLine = ActiveWin.InputLine[:len(ActiveWin.InputLine)-1]
		}
	}

	if ActiveWin != nil && ActiveWin.Con != nil {

		add := string(ebiten.InputChars())
		ActiveWin.InputLine = ActiveWin.InputLine + add
		ActiveWin.ScrollBack = ActiveWin.ScrollBack + add

		if strings.HasSuffix(ActiveWin.InputLine, "\n") {
			n, err := ActiveWin.Con.Write([]byte(ActiveWin.InputLine))
			if err != nil {
				log.Println(n, err)
				os.Exit(0)
			}
			ActiveWin.InputLine = ""
		}
	} else {

		fmt.Println("No connection.")
	}

	ss := strings.Split(ActiveWin.ScrollBack, "\n")

	//Calculate lines that will fit in window height
	numLines := ActiveWin.Height / (ActiveWin.Font.Size + ActiveWin.Font.VerticalSpace)
	//Crop scrollback if needed
	if len(ss) > numLines {
		ActiveWin.Text = strings.Join(ss[len(ss)-numLines:], "\n")
	} else {
		ActiveWin.Text = ActiveWin.ScrollBack
	}

	ActiveWin.Tick++
	return nil
}

//Eventually optimize, don't re-calc draws each time, just store and offset
func (g *Game) Draw(screen *ebiten.Image) {
	// Blink the cursor.
	t := ActiveWin.Text
	if ActiveWin.Tick%60 < 30 {
		t += "_"
	}

	textLen := len(t)
	foundColor := false
	colorStart := 0
	colorEnd := 0
	drawColor := support.ANSI_DEFAULT
	var textColors [MAX_STRING_LENGTH]support.ANSIData

	for z := 0; z < textLen; z++ {
		if t[z] == '\033' {
			foundColor = true
			colorStart = z
			textColors[z] = support.ANSI_CONTROL
			continue
		} else if foundColor && t[z] == 'm' {
			colorEnd = z
			foundColor = false
			if z+1 < textLen {
				drawColor = support.DecodeANSI(t[colorStart : colorEnd+1])
				textColors[z+1] = drawColor
			}
			textColors[z] = support.ANSI_CONTROL
			continue
		}
		if foundColor {
			textColors[z] = support.ANSI_CONTROL
		} else {
			textColors[z] = drawColor
		}
	}

	charWidth := int(math.Round(float64(ActiveWin.Font.Size) / float64(HorizontalSpaceRatio)))
	tLen := len(t)
	y := 0
	x := 0
	for c := 0; c < tLen; c++ {
		if t[c] == '\n' {
			y = 0
			x++
			continue
		}
		if t[c] >= 32 && t[c] < 255 {
			if textColors[c] != support.ANSI_CONTROL {
				charColor := color.RGBA64{textColors[c].Red, textColors[c].Green, textColors[c].Blue, 0xFFFF}
				text.Draw(screen, string(t[c]),
					ActiveWin.Font.Face,
					LeftMargin+charWidth+(y*charWidth),
					ActiveWin.Font.Size+(x*(ActiveWin.Font.Size+ActiveWin.Font.VerticalSpace)),
					charColor)
				y++
			}
		}
	}

}
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	x, y := ebiten.WindowSize()
	ActiveWin.Height = y
	ActiveWin.Width = x
	return x, y
}

func main() {
	var err error

	//Read default font
	MainWin.Font.Data, err = ioutil.ReadFile(DefaultFontFile)
	if err != nil {
		log.Fatal(err)

	}

	//Read default greeting
	var greeting []byte
	greeting, err = ioutil.ReadFile(DefaultGreetFile)
	if err != nil {
		log.Fatal(err)

	}

	//Load font
	tt, err := truetype.Parse(MainWin.Font.Data)
	if err != nil {
		log.Fatal(err)
	}

	//Load window defaults
	MainWin.Title = DefaultWindowTitle
	MainWin.Width = int(math.Round(DefaultWindowWidth * DefaultMagnification))
	MainWin.Height = int(math.Round(DefaultWindowHeight * DefaultMagnification))
	MainWin.DPI = DefaultWindowDPI

	MainWin.Font.VerticalSpace = int(math.Round(DefaultVerticalSpace * DefaultMagnification))
	MainWin.Font.Size = int(math.Round(DefaultFontSize * DefaultMagnification))

	MainWin.RepeatDelay = DefaultRepeatDelay
	MainWin.RepeatInterval = DefaultRepeatInterval
	ActiveWin = &MainWin

	//Init font
	MainWin.Font.Face = truetype.NewFace(tt, &truetype.Options{
		Size:              float64(MainWin.Font.Size),
		DPI:               float64(MainWin.DPI),
		Hinting:           font.HintingFull,
		GlyphCacheEntries: glyphCacheSize,
	})

	greetString := fmt.Sprintf("%v\n%v\n", string(greeting), VersionString)
	MainWin.ScrollBack = greetString
	MainWin.Text = greetString

	ebiten.SetWindowSize(MainWin.Width, MainWin.Height)
	ebiten.SetWindowTitle(MainWin.Title)
	ebiten.SetWindowResizable(true)

	DialSSL()
	go ReadInput()

	g := &Game{
		counter: 0,
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}

}

func DialSSL() {

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", "bhmm.net:7778", conf)
	if err != nil {
		log.Println(err)
		return
	}
	ActiveWin.Con = conn
}
