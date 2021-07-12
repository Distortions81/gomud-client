package main

import (
	_ "embed"
	"fmt"
	"log"
	"sync"

	_ "github.com/flopp/go-findfont"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"golang.org/x/image/font"
)

//Embeds
//default font
//go:embed "data/unispacerg.ttf"
var defaultFont []byte

//Globals
var tt *truetype.Font
var MainWin Window
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

func (g *Game) Update() error {
	return nil
}

func updateNow() {
	textToLines()
	renderText()
}

func init() {
	var err error

	//Load font
	tt, err = truetype.Parse(defaultFont)
	if err != nil {
		log.Fatal(err)
	}

	MainWin.serverAddr = defaultServer
	MainWin.isConnected = false
	MainWin.title = defaultWindowTitle
	MainWin.width = defaultWindowWidth
	MainWin.height = defaultWindowHeight
	MainWin.userScale = defaultUserScale

	//Init font
	MainWin.font.size = defaultFontSize
	MainWin.font.face = truetype.NewFace(tt, &truetype.Options{
		Size:              float64(MainWin.font.size),
		Hinting:           font.HintingFull,
		GlyphCacheEntries: glyphCacheSize,
	})

	//Font setup
	MainWin.font.vertSpace = MainWin.font.size / defaultVerticalSpace
	MainWin.font.charWidth = MainWin.font.size / defaultHorizontalSpace
	MainWin.font.charHeight = MainWin.font.size + MainWin.font.vertSpace

	MainWin.lines.lines[0] = ""
	MainWin.lines.pos = 0
	MainWin.lines.head = 0
	MainWin.lines.tail = 0

	MainWin.dirty = false

	MainWin.offScreen = ebiten.NewImage(int(MainWin.width), int(MainWin.height))
	ebiten.SetWindowTitle(MainWin.title)
	ebiten.SetWindowSize(int(MainWin.width), int(MainWin.height))

	ebiten.SetWindowResizable(true)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetMaxTPS(60)
	ebiten.SetRunnableOnUnfocused(true)
	if clearEveryFrame == true {
		ebiten.SetScreenClearedEveryFrame(true)
	} else {
		ebiten.SetScreenClearedEveryFrame(false)
	}

	AddLine(
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

func (g *Game) Draw(screen *ebiten.Image) {

	//Resize, or hidpi detected
	sx, sy := screen.Size()
	if MainWin.realWidth != sx || MainWin.realHeight != sy {
		MainWin.realWidth = sx
		MainWin.realHeight = sy
		MainWin.offScreen = ebiten.NewImage(sx, sy)

		MainWin.dirty = false
		for x := 0; x < MAX_SCROLL_LINES; x++ {
			MainWin.lines.pixLines[x] = nil
		}
		fmt.Println("Buffer resized.")
		updateNow()
	}

	if MainWin.dirty == true || clearEveryFrame {
		MainWin.dirty = false
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest

		screen.DrawImage(MainWin.offScreen, op)
	}
}
