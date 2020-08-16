package main

import (
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

var screenWidth = 1280
var screenHeight = 720
var ourFont []byte

// repeatingKeyPressed return true when key is pressed considering the repeat state.
func repeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 30
		interval = 3
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= delay && (d-delay)%interval == 0 {
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

	// Adjust the string to be at most 10 lines.
	ss := strings.Split(g.text, "\n")
	numLines := screenHeight / 30

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
	//ebitenutil.DebugPrint(screen, t)

	var mplusNormalFont font.Face
	tt, err := truetype.Parse(ourFont)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	mplusNormalFont = truetype.NewFace(tt, &truetype.Options{
		Size:    30,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})

	text.Draw(screen, t, mplusNormalFont, 0, 30, color.White)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	x, y := ebiten.WindowSize()
	screenHeight = y
	screenWidth = x
	return x, y
}

func main() {
	var err error
	ourFont, err = ioutil.ReadFile("unispace rg.ttf")
	if err != nil {
		log.Fatal(err)

	}

	g := &Game{
		text:    "Type on the keyboard:\n",
		counter: 0,
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("TypeWriter (Ebiten Demo)")
	ebiten.SetWindowResizable(true)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
