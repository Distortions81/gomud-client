package main

import (
	"strings"
)

func AddLine(text string) {

	go func() {
		MainWin.lines.rawTextLock.Lock()
		MainWin.lines.rawText += text
		MainWin.lines.rawTextLock.Unlock()
	}()
}

func textToLines() {
	MainWin.lines.rawTextLock.Lock()
	lines := strings.Split(MainWin.lines.rawText, "\n")
	MainWin.lines.rawTextLock.Unlock()

	numLines := len(lines) - 1
	if numLines > 0 {

		x := 0
		for i := MainWin.lines.head + 1; i < MAX_SCROLL_LINES && x <= numLines; i++ {
			MainWin.lines.lines[i] = lines[x]
			MainWin.lines.colors[i] = AnsiColor(lines[x])
			x++
		}
		MainWin.lines.head += x
		MainWin.lines.rawText = ""
		updateNow()
	}
}
