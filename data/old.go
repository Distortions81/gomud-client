package old

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"./support"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
)

//default font
//go:embed "unispacerg.ttf"
var DefaultFont []byte

//default greeting
//go:embed "greet.txt"
var greeting []byte

const DefaultServer = "127.0.0.1:7778"
const VersionString = "Pre-Alpha build, v0.0.02 07-04-2021-1055p"
const MAX_STRING_LENGTH = 1024 * 10

//Window
const DefaultWindowWidth = 640.0
const DefaultWindowHeight = 360.0

const DefaultRepeatInterval = 3
const DefaultRepeatDelay = 30

const DefaultWindowTitle = "GoMud-Client"

type FontData struct {
	VerticalSpace float64
	Size          float64
	Data          []byte
	Face          font.Face

	Color string
	Dirty bool
}

type Window struct {
	Title       string
	ConName     string
	Update      bool
	FrameBuffer *ebiten.Image
	Lock        sync.Mutex

	Width     int
	Height    int
	Scale     float64
	UserScale float64

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

	dirty bool
}

//Scroll
type ScrollData struct {
	ScrollPos int
}

const glyphCacheSize = 256
const DefaultVerticalSpace = 4.0
const HorizontalSpaceRatio = 1.5
const DefaultFontSize = 12.0
const LeftMargin = 6.0

//Data
var MainWin Window
var ActiveWin *Window
var tt *truetype.Font

type Game struct {
	counter int
}

// RepeatingKeyPressed return true when key is pressed considering the repeat state.
func RepeatingKeyPressed(key ebiten.Key) bool {
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
	numLines := int((float64(ActiveWin.Height)) / (ActiveWin.Font.Size + ActiveWin.Font.VerticalSpace) * ActiveWin.Scale)
	//Crop scrollback if needed
	if len(ss) > numLines {
		ActiveWin.Text = strings.Join(ss[len(ss)-numLines:], "\n")
	} else {
		ActiveWin.Text = ActiveWin.ScrollBack
	}
}

func AddText(newData string) {
	if newData != "" {
		ActiveWin.Lock.Lock()
		ActiveWin.Update = true
		ActiveWin.ScrollBack += newData
		updateScroll()
		ActiveWin.Lock.Unlock()
	}
}

func ReadInput() {
	for {
		buf := make([]byte, MAX_STRING_LENGTH)
		if ActiveWin.Con != nil {
			n, err := ActiveWin.Con.Read(buf)
			if err != nil {
				log.Println(n, err)
				ActiveWin.Con.Close()

				buf := fmt.Sprintf("Lost connection to %s: %s\r\n", ActiveWin.ConName, err)
				AddText(buf)
				ActiveWin.Con = nil
			}
			newData := string(buf[:n])
			AddText(newData)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (g *Game) Update() error {
	startTime := time.Now()

	keyPressed := false

	ActiveWin.Lock.Lock()
	defer ActiveWin.Lock.Unlock()

	//Increase mag
	if RepeatingKeyPressed(ebiten.KeyEqual) && ebiten.IsKeyPressed(ebiten.KeyControl) {
		ActiveWin.UserScale = ActiveWin.UserScale + 0.15
		ActiveWin.Update = true
		adjustScale()
		updateScroll()
		return nil
	} else if RepeatingKeyPressed(ebiten.KeyMinus) && ebiten.IsKeyPressed(ebiten.KeyControl) {
		ActiveWin.UserScale = ActiveWin.UserScale - 0.15
		ActiveWin.Update = true
		adjustScale()
		updateScroll()
		return nil
	} else if RepeatingKeyPressed(ebiten.KeyEnter) || RepeatingKeyPressed(ebiten.KeyKPEnter) {
		ActiveWin.InputLine += "\n"
		keyPressed = true
	} else if RepeatingKeyPressed(ebiten.KeyBackspace) {
		if len(ActiveWin.InputLine) >= 1 {
			ActiveWin.InputLine = ActiveWin.InputLine[:len(ActiveWin.InputLine)-1]
			ActiveWin.ScrollBack = ActiveWin.ScrollBack[:len(ActiveWin.ScrollBack)-1]
			keyPressed = true
		}
	}
	newChars := string(ebiten.InputChars())
	if ActiveWin != nil && ActiveWin.Con != nil && (newChars != "" || keyPressed) {

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
	}
	g.counter++

	since := time.Since(startTime).Nanoseconds()
	time.Sleep(time.Duration(16666666-since) * time.Nanosecond)
	return nil
}

//Eventually optimize, don't re-calc draws each time, just store and offset
func (g *Game) Draw(screen *ebiten.Image) {

	startTime := time.Now()
	if ActiveWin.Update {

		//Mutex
		ActiveWin.Lock.Lock()
		defer ActiveWin.Lock.Unlock()

		//Just shorthand
		t := ActiveWin.Text

		//Clear screen, auto clear disabled for performance
		//screen.Clear()

		//We are drawing, so we can turn off the need-update flag.
		ActiveWin.Update = false

		textLen := len(t)
		foundColor := false               //Mark when color code starts
		colorStart := 0                   //Color code start position
		colorEnd := 0                     //Color code end position
		drawColor := support.ANSI_DEFAULT //Var to store the colors

		var textColors [MAX_STRING_LENGTH]support.ANSIData

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

		charWidth := int(math.Round(float64(ActiveWin.Font.Size) / float64(HorizontalSpaceRatio))) // Calc charater pixel width
		tLen := len(t)
		y := 0
		x := 0
		ActiveWin.FrameBuffer.Clear() //Clear frame buffer, buffer not actually needed now.
		//ActiveWin.FrameBuffer.Fill(color.NRGBA{0xFF, 0x00, 0x00, 0xff})
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
						int(math.Round(LeftMargin+float64(charWidth)+(float64(y-1)*float64(charWidth)))),
						int(math.Round(ActiveWin.Font.Size+(float64(x)*(ActiveWin.Font.Size+ActiveWin.Font.VerticalSpace)))),
						charColor)
					y++
				}
			}
		}
		screen.DrawImage(ActiveWin.FrameBuffer, nil) //Draw to screen
	}

	since := time.Since(startTime).Nanoseconds()
	time.Sleep(time.Duration(16666666-since) * time.Nanosecond) //Sleep for the rest of the frame time
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {

	NewWidth := int(math.Round(float64(outsideWidth) * ActiveWin.Scale))
	NewHeight := int(math.Round(float64(outsideHeight) * ActiveWin.Scale))

	if NewWidth != ActiveWin.Width || NewHeight != ActiveWin.Height {
		ActiveWin.Lock.Lock()
		defer ActiveWin.Lock.Unlock()

		adjustScale()
		updateScroll()
		ActiveWin.Update = true // Layout changed, redraw screen
	}

	return NewWidth, NewHeight
}

func adjustScale() {
	x, y := ebiten.WindowSize() // Get window size
	ActiveWin.Width = x
	ActiveWin.Height = y

	ActiveWin.Title = DefaultWindowTitle
	//Re-calculate vertical line spacing, may not be needed?
	ActiveWin.Font.VerticalSpace = DefaultVerticalSpace * ActiveWin.Scale * ActiveWin.UserScale
	//Recalculate font size based on new scale
	ActiveWin.Font.Size = DefaultFontSize * ActiveWin.Scale * ActiveWin.UserScale

	//Init font
	ActiveWin.Font.Face = truetype.NewFace(tt, &truetype.Options{
		Size:              float64(ActiveWin.Font.Size),
		Hinting:           font.HintingFull,
		GlyphCacheEntries: glyphCacheSize,
	})

	//New framebuffer with new size.
	ActiveWin.FrameBuffer = ebiten.NewImage(
		ActiveWin.Width,
		ActiveWin.Height)
}

func notmain() {
	var err error

	//Load font
	tt, err = truetype.Parse(DefaultFont)
	if err != nil {
		log.Fatal(err)
	}

	ActiveWin = &MainWin
	//Load window defaults
	ActiveWin.Title = DefaultWindowTitle
	ActiveWin.UserScale = 1.0
	ActiveWin.Scale = 1.0
	ActiveWin.Width = DefaultWindowWidth
	ActiveWin.Height = DefaultWindowHeight

	ActiveWin.Font.VerticalSpace = DefaultVerticalSpace
	ActiveWin.Font.Size = DefaultFontSize * ActiveWin.Scale * ActiveWin.UserScale

	ActiveWin.RepeatDelay = DefaultRepeatDelay
	ActiveWin.RepeatInterval = DefaultRepeatInterval

	//Setup window.
	ebiten.SetWindowSize(ActiveWin.Width, ActiveWin.Height)
	ebiten.SetWindowTitle(ActiveWin.Title)
	ebiten.SetWindowResizable(true)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetMaxTPS(60)
	ebiten.SetRunnableOnUnfocused(true)

	//Setup frame buffer
	ActiveWin.FrameBuffer = ebiten.NewImage(
		ActiveWin.Width,
		ActiveWin.Height)

	//Init font
	ActiveWin.Font.Face = truetype.NewFace(tt, &truetype.Options{
		Size:              float64(ActiveWin.Font.Size),
		Hinting:           font.HintingFull,
		GlyphCacheEntries: glyphCacheSize,
	})

	//Draw greeting to window.
	greetString := fmt.Sprintf("%v\n%v\n", string(greeting), VersionString)
	ActiveWin.ScrollBack = greetString
	ActiveWin.Text = greetString

	go func() {
		DialSSL(DefaultServer) //Connnect
		go ReadInput()         //Start reading connection
	}()

	ebiten.RunGame(&Game{})
}

func DialSSL(addr string) {

	buf := fmt.Sprintf("Connecting to: %s\r\n", addr)
	AddText(buf)

	//Todo, allow timeout adjustment and connection canceling.
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	ActiveWin.ConName = addr
	conn, err := tls.Dial("tcp", addr, conf)
	if err != nil {
		log.Println(err)

		buf := fmt.Sprintf("%s\r\n", err)
		AddText(buf)
		return
	}
	ActiveWin.Con = conn
}
