package main

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
)

// RepeatingKeyPressed return true when key is pressed considering the repeat state.
func RepeatingKeyPressed(key ebiten.Key) bool {
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= MainWin.repeatDelay && (d-MainWin.repeatDelay)%MainWin.repeatInterval == 0 {
		return true
	}
	return false
}
