package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
)

const VersionString = "Pre-Alpha build, v0.0.01 08162020-0909"

//Greeting
const DefaultGreetFile = "greet.txt"

//Window
const DefaultWindowWidth = 1280
const DefaultWindowHeight = 720
const DefaultWindowDPI = 72

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
}

//Scroll
type ScrollData struct {
	ScrollPos int
}

//Font
const DefaultFontFile = "unispace rg.ttf"
const glyphCacheSize = 512
const DefaultVerticalSpace = 6
const DefaultFontSize = 30

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

type Game struct {
	text    string
	counter int
}

func (g *Game) Update(screen *ebiten.Image) error {
	// Add a string from InputChars, that returns string input by users.
	// Note that InputChars result changes every frame, so you need to call this
	// every frame.
	g.text += string(ebiten.InputChars())

	// Adjust the string to be at most x lines.
	ss := strings.Split(g.text, "\n")
	numLines := ActiveWin.Height / (ActiveWin.Font.Size + ActiveWin.Font.VerticalSpace)

	if len(ss) > numLines {
		g.text = strings.Join(ss[len(ss)-numLines:], "\n")
	}

	// If the enter key is pressed, add a line break.
	if repeatingKeyPressed(ebiten.KeyEnter) || repeatingKeyPressed(ebiten.KeyKPEnter) {
		g.text += "\n"
	}

	// If the backspace key is pressed, remove one character.
	if repeatingKeyPressed(ebiten.KeyBackspace) {
		if len(g.text) >= 1 {
			g.text = g.text[:len(g.text)-1]
		}
	}

	g.counter++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Blink the cursor.
	t := g.text
	if g.counter%60 < 30 {
		t += "_"
	}

	lines := strings.Split(t, "\n")
	for x, l := range lines {
		text.Draw(screen, l, ActiveWin.Font.Face, 0, ActiveWin.Font.Size+(x*(ActiveWin.Font.Size+ActiveWin.Font.VerticalSpace)), color.White)
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
	MainWin.Width = DefaultWindowWidth
	MainWin.Height = DefaultWindowHeight
	MainWin.DPI = DefaultWindowDPI

	MainWin.Font.VerticalSpace = DefaultVerticalSpace
	MainWin.Font.Size = DefaultFontSize

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
	g := &Game{
		text:    greetString,
		counter: 0,
	}

	ebiten.SetWindowSize(MainWin.Width, MainWin.Height)
	ebiten.SetWindowTitle(MainWin.Title)
	ebiten.SetWindowResizable(true)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
