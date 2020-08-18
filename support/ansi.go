package support

type ANSIData struct {
	Red   uint16
	Green uint16
	Blue  uint16
	Alpha uint16

	Style int
}

const ANSI_STYLE_RESET = 0
const ANSI_STYLE_ITALIC = 1
const ANSI_STYLE_UNDERLINE = 2
const ANSI_STYLE_INVERSE = 3
const ANSI_STYLE_STRIKE = 4
const ANSI_STYLE_ERROR = 5
const ANSI_STYLE_CONTROL = 6

var ANSI_CONTROL = ANSIData{Style: ANSI_STYLE_CONTROL, Red: 0xAAAA}
var ANSI_RESET = ANSIData{Style: ANSI_STYLE_RESET, Red: 0xFFFF, Green: 0xFFFF, Blue: 0xFFFF}
var ANSI_ITALIC = ANSIData{Style: ANSI_STYLE_ITALIC}
var ANSI_UNDERLINE = ANSIData{Style: ANSI_STYLE_UNDERLINE}
var ANSI_INVERSE = ANSIData{Style: ANSI_STYLE_INVERSE}
var ANSI_STRIKE = ANSIData{Style: ANSI_STYLE_STRIKE}
var ANSI_ERROR = ANSIData{Style: ANSI_STYLE_ERROR, Red: 0xFFFF}

var ANSI_DEFAULT = ANSIData{Red: 0xFFFF, Green: 0xFFFF, Blue: 0xFFFF}
var ANSI_BLACK = ANSIData{Red: 0x0000, Green: 0x0000, Blue: 0x0000}
var ANSI_RED = ANSIData{Red: 0x7FFF}
var ANSI_GREEN = ANSIData{Green: 0x7FFF}
var ANSI_YELLOW = ANSIData{Red: 0x7FFF, Green: 0x7FFF}
var ANSI_BLUE = ANSIData{Blue: 0x7FFF}
var ANSI_MAGENTA = ANSIData{Red: 0x7FFF, Blue: 0x7FFF}
var ANSI_CYAN = ANSIData{Green: 0x7FFF, Blue: 0x7FFF}
var ANSI_GRAY = ANSIData{Red: 0x7FFF, Green: 0x7FFF, Blue: 0x7FFF}

var ANSI_LGRAY = ANSIData{Red: 0xAAAA, Green: 0xAAAA, Blue: 0xAAAA}
var ANSI_LRED = ANSIData{Red: 0xFFFF}
var ANSI_LGREEN = ANSIData{Green: 0xFFFF}
var ANSI_LYELLOW = ANSIData{Red: 0xFFFF, Green: 0xFFFF}
var ANSI_LBLUE = ANSIData{Blue: 0xFFFF}
var ANSI_LMAGENTA = ANSIData{Red: 0xFFFF, Blue: 0xFFFF}
var ANSI_LCYAN = ANSIData{Green: 0xFFFF, Blue: 0xFFFF}
var ANSI_WHITE = ANSIData{Red: 0xFFFF, Green: 0xFFFF, Blue: 0xFFFF}

func DecodeANSI(c string) ANSIData {

	if c == "\033[0m" {
		return ANSI_RESET
	} else if c == "\033[0;3m" {
		return ANSI_ITALIC
	} else if c == "\033[0;4m" {
		return ANSI_UNDERLINE
	} else if c == "\033[0;7m" {
		return ANSI_INVERSE
	} else if c == "\033[0;9m" {
		return ANSI_STRIKE

	} else if c == "\033[0;30m" {
		return ANSI_BLACK
	} else if c == "\033[0;31m" {
		return ANSI_RED
	} else if c == "\033[0;32m" {
		return ANSI_GREEN
	} else if c == "\033[0;33m" {
		return ANSI_YELLOW
	} else if c == "\033[0;34m" {
		return ANSI_BLUE
	} else if c == "\033[0;35m" {
		return ANSI_MAGENTA
	} else if c == "\033[0;36m" {
		return ANSI_CYAN
	} else if c == "\033[0;37m" {
		return ANSI_GRAY

	} else if c == "\033[1;30m" {
		return ANSI_LGRAY
	} else if c == "\033[1;31m" {
		return ANSI_LRED
	} else if c == "\033[1;32m" {
		return ANSI_LGREEN
	} else if c == "\033[1;33m" {
		return ANSI_LYELLOW
	} else if c == "\033[1;34m" {
		return ANSI_LBLUE
	} else if c == "\033[1;35m" {
		return ANSI_LMAGENTA
	} else if c == "\033[1;36m" {
		return ANSI_LCYAN
	} else if c == "\033[1;37m" {
		return ANSI_WHITE
	}

	return ANSI_ERROR
}

func StripANSI(c string) string {

	foundColor := false
	startColor := 0
	endColor := 0

	slen := len(c)
	for z := 0; z < slen; z++ {
		if c[z] == '\033' {
			foundColor = true
			startColor = z
		} else if foundColor && c[z] == 'm' {
			endColor = z
			c = c[0:startColor-1] + c[endColor+1:]
			slen = len(c)
		}
	}
	return c
}
