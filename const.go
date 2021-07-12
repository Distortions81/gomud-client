package main

const glyphCacheSize = 256
const defaultFontSize = 18.0
const clearEveryFrame = true

//Constants
const MAX_INPUT_LENGTH = 100 * 1024 //100kb, some kind of reasonable limit for net/input buffer
const NET_POLL_MS = 66              //1/15th of a second

const MAX_SCROLL_LINES = 10000 //Max scrollback
const MAX_VIEW_LINES = 250     //Maximum lines on screen

const defaultWindowTitle = "GoMud-Client"
const defaultServer = "127.0.0.1:7778"
const VersionString = "Pre-Alpha build, v0.0.031 07092021-1201a"

const defaultHorizontalSpace = 1.4
const defaultVerticalSpace = 4.0

const defaultWindowWidth = 960
const defaultWindowHeight = 540
const defaultUserScale = 1.0

const defaultRepeatInterval = 3
const defaultRepeatDelay = 30
