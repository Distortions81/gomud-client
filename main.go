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

const DefaultRepeatInterval = 3
const DefaultRepeatDelay = 60

const DefaultWindowTitle = "GoMud-Client"

type Window struct {
	Title       string
	Update      bool
	FrameBuffer *ebiten.Image

	Width  int
	Height int
	Scale  float64

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
var tt *truetype.Font

type Game struct {
	counter int
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

func updateScroll() {
	ss := strings.Split(ActiveWin.ScrollBack, "\n")

	//Calculate lines that will fit in window height
	numLines := ActiveWin.Height / (ActiveWin.Font.Size + ActiveWin.Font.VerticalSpace)
	//Crop scrollback if needed
	if len(ss) > numLines {
		ActiveWin.Text = strings.Join(ss[len(ss)-numLines:], "\n")
	} else {
		ActiveWin.Text = ActiveWin.ScrollBack
	}
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
		newData := string(buf[:n])
		if newData != "" {
			ActiveWin.Update = true
			ActiveWin.ScrollBack += newData
			updateScroll()
		}
	}
}

func (g *Game) Update(screen *ebiten.Image) error {
	keyPressed := false

	//Increase mag
	if repeatingKeyPressed(ebiten.KeyEqual) {
		ActiveWin.Scale = ActiveWin.Scale + 0.15
		adjustScale()
		return nil
	}

	//Decrease mag
	if repeatingKeyPressed(ebiten.KeyMinus) {
		ActiveWin.Scale = ActiveWin.Scale - 0.15
		adjustScale()
		return nil
	}

	//Add linebreaks
	if repeatingKeyPressed(ebiten.KeyEnter) || repeatingKeyPressed(ebiten.KeyKPEnter) {
		ActiveWin.InputLine += "\n"
		keyPressed = true
	}

	//Backspace
	if repeatingKeyPressed(ebiten.KeyBackspace) {
		if len(ActiveWin.InputLine) >= 1 {
			ActiveWin.InputLine = ActiveWin.InputLine[:len(ActiveWin.InputLine)-1]
			ActiveWin.ScrollBack = ActiveWin.ScrollBack[:len(ActiveWin.ScrollBack)-1]
			keyPressed = true
		}
	}
	newChars := string(ebiten.InputChars())
	if ActiveWin != nil && ActiveWin.Con != nil {

		if newChars != "" || keyPressed {
			ActiveWin.Update = true
			add := newChars
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
			updateScroll()
		}
	} else {

		fmt.Println("No connection.")
	}
	ActiveWin.Tick++
	return nil
}

//Eventually optimize, don't re-calc draws each time, just store and offset
func (g *Game) Draw(screen *ebiten.Image) {
	t := ActiveWin.Text

	if ActiveWin.Update {
		ActiveWin.Update = false

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
		err := ActiveWin.FrameBuffer.Clear()
		if err != nil {
			log.Fatal(err)
		}
		for c := 0; c < tLen; c++ {
			if t[c] == '\n' {
				y = 0
				x++
				continue
			}
			if t[c] >= 32 && t[c] < 255 {
				if textColors[c] != support.ANSI_CONTROL {
					charColor := color.RGBA64{textColors[c].Red, textColors[c].Green, textColors[c].Blue, 0xFFFF}
					text.Draw(ActiveWin.FrameBuffer, string(t[c]),
						ActiveWin.Font.Face,
						LeftMargin+charWidth+(y*charWidth),
						ActiveWin.Font.Size+(x*(ActiveWin.Font.Size+ActiveWin.Font.VerticalSpace)),
						charColor)
					y++
				}
			}
		}
	}
	err := screen.DrawImage(ActiveWin.FrameBuffer, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// The unit of outsideWidth/Height is device-independent pixels.
	// By multiplying them by the device scale factor, we can get a hi-DPI screen size.
	s := ebiten.DeviceScaleFactor()
	fmt.Println(fmt.Sprintf("Scale %v", s))
	return int(float64(outsideWidth) * s), int(float64(outsideHeight) * s)
}

func adjustScale() {

	ActiveWin.Title = DefaultWindowTitle
	ActiveWin.Width = int(math.Round(DefaultWindowWidth * ActiveWin.Scale))
	ActiveWin.Height = int(math.Round(DefaultWindowHeight * ActiveWin.Scale))

	ActiveWin.Font.VerticalSpace = int(math.Round(DefaultVerticalSpace * ActiveWin.Scale))
	ActiveWin.Font.Size = int(math.Round(DefaultFontSize * ActiveWin.Scale))

	//Init font
	ActiveWin.Font.Face = truetype.NewFace(tt, &truetype.Options{
		Size:              float64(ActiveWin.Font.Size),
		Hinting:           font.HintingFull,
		GlyphCacheEntries: glyphCacheSize,
	})

	ActiveWin.FrameBuffer, _ = ebiten.NewImage(ActiveWin.Width, ActiveWin.Height, ebiten.FilterNearest)
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
	tt, err = truetype.Parse(MainWin.Font.Data)
	if err != nil {
		log.Fatal(err)
	}

	ActiveWin = &MainWin
	//Load window defaults
	ActiveWin.Title = DefaultWindowTitle
	ActiveWin.Scale = ebiten.DeviceScaleFactor() * 2.0
	ActiveWin.Width = int(math.Round(DefaultWindowWidth * ActiveWin.Scale))
	ActiveWin.Height = int(math.Round(DefaultWindowHeight * ActiveWin.Scale))

	ActiveWin.Font.VerticalSpace = int(math.Round(DefaultVerticalSpace * ActiveWin.Scale))
	ActiveWin.Font.Size = int(math.Round(DefaultFontSize * ActiveWin.Scale))

	ActiveWin.RepeatDelay = DefaultRepeatDelay
	ActiveWin.RepeatInterval = DefaultRepeatInterval

	//Init font
	ActiveWin.Font.Face = truetype.NewFace(tt, &truetype.Options{
		Size:              float64(ActiveWin.Font.Size),
		Hinting:           font.HintingFull,
		GlyphCacheEntries: glyphCacheSize,
	})

	greetString := fmt.Sprintf("%v\n%v\n", string(greeting), VersionString)
	ActiveWin.ScrollBack = greetString
	ActiveWin.Text = greetString

	ebiten.SetWindowSize(ActiveWin.Width, ActiveWin.Height)
	ebiten.SetWindowTitle(ActiveWin.Title)
	ebiten.SetWindowResizable(true)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetMaxTPS(30)
	ActiveWin.FrameBuffer, _ = ebiten.NewImage(ActiveWin.Width, ActiveWin.Height, ebiten.FilterNearest)

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
