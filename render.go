package main

import (
	"image/color"
	"math"
	"strconv"

	"./support"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/text"
)

func renderOffscreen() {

	if MainWin.dirty == false { //Check for no pending frames, or we will re-render for no reason
		ebitenLock.Lock()
		defer ebitenLock.Unlock()

		MainWin.offScreen.Clear()
		//MainWin.offScreen.Fill(color.RGBA{0x30, 0x00, 0x00, 0xFF})

		//Render our images out here
		head := MainWin.lines.head
		tail := MainWin.lines.tail
		for a := tail; a <= head && a >= tail; a++ {

			if MainWin.lines.pixLines[a] != nil {
				op := &ebiten.DrawImageOptions{}
				op.Filter = ebiten.FilterNearest
				op.GeoM.Translate(0.0, float64(a)*MainWin.font.charHeight)
				MainWin.offScreen.DrawImage(MainWin.lines.pixLines[a], op)
			} else {
				//Stop rendering here, lines after this are not yet ready.
				return
			}
		}
		MainWin.dirty = true
	}

}

func renderText() {

	didRender := false

	head := MainWin.lines.head
	tail := MainWin.lines.tail

	//Render old to new, so color codes can persist lines
	for a := tail; a <= head && a >= tail; a++ {
		if MainWin.lines.pixLines[a] == nil {
			MainWin.lines.pixLines[a] = renderLine(a)
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
	if MainWin.realWidth > 0 && MainWin.font.size > 0 {
		len := len(MainWin.lines.lines[pos])
		ebitenLock.Lock()
		defer ebitenLock.Unlock()

		tempImg := ebiten.NewImage(MainWin.realWidth, int(math.Round(MainWin.font.size+MainWin.font.vertSpace)))
		x := 0
		for i := 0; i < len; i++ {
			if strconv.IsPrint(rune(MainWin.lines.lines[pos][i])) && MainWin.lines.colors[pos][i] != support.ANSI_CONTROL {
				x++
				text.Draw(tempImg, string(MainWin.lines.lines[pos][i]),
					MainWin.font.face,
					int(math.Round(float64(x)*MainWin.font.charWidth)),
					int(math.Round(MainWin.font.size)),
					color.RGBA{MainWin.lines.colors[pos][i].Red, MainWin.lines.colors[pos][i].Green, MainWin.lines.colors[pos][i].Blue, 0xFF})
			}
		}
		return tempImg
	}
	return nil
}
