package pager

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"log"
	"regexp"
	"strings"
)

const (
	DEBUG = false

	QUIT      = 0
	NEXT_FILE = 1
	NO_ACTION = 2
)

type Pager struct {
	str          string   // contents to display
	lines        int      // num of lines in str
	Files        []string // files
	ignoreY      int      // ignore lines
	ignoreX      int      // ignore columns
	Index        int      // file index
	File         string   // current file
	isSlashOn    bool     // input search string mode
	isSearchMode bool     // search mode
	searchIndex  int      // current search index
	searchStr    string   // string to search
}

func (p *Pager) SetContent(s string) {
	p.str = s
	lines := regexp.MustCompile("(?m)$").FindAllString(p.str, -1)
	p.ignoreX = 0
	p.ignoreY = 0
	p.lines = len(lines)
}

func (p *Pager) AddContent(s string) {
	p.str += s
	lines := regexp.MustCompile("(?m)$").FindAllString(p.str, -1)
	p.lines = len(lines)
}

func (p *Pager) drawLine(x, y int, str string, canSkip bool) {
	color := termbox.ColorDefault
	backgroundColor := termbox.ColorDefault
	foundIndex := 0
	runes := []rune(str)

	minusX := p.ignoreX

	colorMap := map[string]termbox.Attribute{
		"0m": termbox.ColorBlack,
		"1m": termbox.ColorRed,
		"2m": termbox.ColorGreen,
		"3m": termbox.ColorYellow,
		"4m": termbox.ColorBlue,
		"5m": termbox.ColorMagenta,
		"6m": termbox.ColorCyan,
		"7m": termbox.ColorWhite,
	}
	attrMap := map[string]termbox.Attribute{
		"1m": termbox.AttrBold,
		"3m": termbox.AttrReverse,
		"4m": termbox.AttrUnderline,
	}

	if DEBUG {
		for i := 0; i < len(runes); i++ {
			if runes[i] == '\033' {
				log.Println("ESC")
			} else {
				log.Println(string(runes[i]))
			}
		}
		panic(1)
	}

	searchString := p.searchStr
	searchStringLen := len(searchString)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\n' {
			y++
			minusX = i + (1 + p.ignoreX)
		}
		if searchStringLen > 0 && i+searchStringLen < len(runes) { // highlight search string
			t := strings.ToUpper(string(runes[i : i+searchStringLen]))
			m := strings.ToUpper(searchString[foundIndex:searchStringLen])
			if t == m {
				backgroundColor = termbox.AttrReverse
				foundIndex = searchStringLen - 1
			} else if foundIndex == 0 {
				backgroundColor = termbox.ColorDefault
			} else {
				foundIndex--
			}
		}
		if runes[i] == '\033' { // not good
			minusX++
			if i+2 < len(runes) && string(runes[i:i+3]) == "\033[m" { // reset?
				color = termbox.ColorDefault
				backgroundColor = termbox.ColorDefault
				i += 2
				minusX += 2
				// .[H.[J.[m.top  // .[m.top -
			} else if i+6 < len(runes) && string(runes[i:i+6]) == "\033[?25l" { // clear \033[J
				runes = runes[i+12 : len(runes)]
				if DEBUG {
					log.Println("1:" + string(runes))
				}
				p.SetContent(string(runes))
				p.drawLine(0, 0, p.str, false)
				p.Clear()
				break
			} else if i+3 < len(runes) && string(runes[i:i+3]) == "\033[J" { // clear \033[J
				runes = runes[i+4 : len(runes)]
				if DEBUG {
					log.Println("2:" + string(runes))
				}
				p.SetContent(string(runes))
				termbox.Sync()
				p.Draw()
				break
			} else if i+3 < len(runes) && runes[i+1] == '[' && runes[i+3] == 'm' {
				t := string(runes[i+2 : i+4])
				if t == "0m" { // reset
					color = termbox.ColorDefault
					backgroundColor = termbox.ColorDefault
				} else { // attribute
					if a, ok := attrMap[t]; ok {
						color |= a
					} else {
						// not supported attribute
					}
				}
				i += 3
				minusX += 3
			} else if i+4 < len(runes) && string(runes[i:i+2]) == "\033[" && (runes[i+2] == '0' || runes[i+2] == '3' || runes[i+2] == '4') && runes[i+4] == 'm' { // \033[30m or  \033[40m
				// color
				c := string(runes[i+3 : i+5])
				if runes[i+2] == '3' {
					color = colorMap[c]
				} else if runes[i+2] == '4' {
					backgroundColor = colorMap[c]
				} else {
					panic(1)
					color |= termbox.AttrBold
				}
				i += 4
				minusX += 4
			} else if i+7 < len(runes) && string(runes[i:i+2]) == "\033[" && runes[i+4] == ';' && (runes[i+5] == '3' || runes[i+2] == '4') && runes[i+7] == 'm' { // \033[01;30m or \033[01;40m
				// attr + color
				a := string(runes[i+2 : i+4])
				c := string(runes[i+6 : i+8])
				if runes[i+2] == '3' {
					color = colorMap[c]
				} else {
					backgroundColor = colorMap[c]
				}
				if a == "01" {
					color |= termbox.AttrBold
				}
				i += 7
				minusX += 7
			} else if i+1 < len(runes) && string(runes[i:i+2]) == "\033[" { // \033[K
				i += 2
				minusX += 2
			}
			continue
		}
		if canSkip {
			termbox.SetCell(x+i-minusX, y-(p.ignoreY)+1, runes[i], color, backgroundColor)
		} else {
			termbox.SetCell(x+i, y, runes[i], termbox.ColorBlue, termbox.ColorWhite)
		}
	}
}

func (p *Pager) Clear() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()
	termbox.Sync()
	p.Draw()
}

func (p *Pager) Draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	p.drawLine(0, 0, p.str, true)
	maxX, _ := termbox.Size()
	empty := make([]byte, maxX)
	mode := ""
	file := ""
	nextFileUsage := ""
	if p.File != "" {
		file = " :: [file: " + p.File + " ]"
	}
	if p.isSearchMode {
		mode = fmt.Sprintf(file+" :: [searching: %s (lines: %d)] :: [forward search: n] [backward search: N] [exit search: ESC/Ctrl-C]", p.searchStr, p.ignoreY)
	} else if p.isSlashOn {
		mode = fmt.Sprintf(file+" :: [search string: %s ]", p.searchStr)
	} else if file != "" {
		mode = file
	}
	if len(p.Files) > 1 {
		nextFileUsage = "[next file: Ctrl-h,Ctrl-l]"
	}
	p.drawLine(0, 0, "USAGE [exit: ESC/q] [scroll: j,k/C-n,C-p] "+nextFileUsage+mode+string(empty), false)
	termbox.Flush()
}

func (p *Pager) viewModeKey(ev termbox.Event) int {
	switch ev.Key {
	case termbox.KeyEsc, termbox.KeyCtrlC:
		termbox.Flush()
		return QUIT
	case termbox.KeyArrowRight:
		p.scrollRight()
	case termbox.KeyArrowLeft:
		p.scrollLeft()
	case termbox.KeyCtrlL:
		if p.isMaxIndex() {
			p.Index--
		} else {
			p.ignoreY = 0
			termbox.Sync()
			return NEXT_FILE
		}
	case termbox.KeyCtrlH:
		if p.Index >= 1 {
			p.Index -= 2
			p.ignoreY = 0
			termbox.Sync()
			return NEXT_FILE
		}
	case termbox.KeyCtrlN, termbox.KeyArrowDown, termbox.KeyEnter:
		p.scrollDown()
	case termbox.KeyCtrlP, termbox.KeyArrowUp:
		p.scrollUp()
	case termbox.KeyCtrlD, termbox.KeySpace:
		_, y := termbox.Size()
		if p.ignoreY+29 < (p.lines - y + 1) {
			p.ignoreY += 29
		} else {
			p.ignoreY = p.lines - y + 1
		}
		p.Draw()
	case termbox.KeyCtrlU:
		p.ignoreY -= 29
		if p.ignoreY < 0 {
			p.ignoreY = 0
		}
		p.Draw()
	default:
		switch ev.Ch {
		case 'j':
			p.scrollDown()
		case 'k':
			p.scrollUp()
		case 'l':
			p.scrollRight()
		case 'h':
			p.scrollLeft()
		case 'q':
			termbox.Sync()
			return QUIT
		case '<':
			p.ignoreY = 0
			p.ignoreX = 0
			termbox.Sync()
			p.Draw()
		case '>':
			_, y := termbox.Size()
			p.ignoreY = p.lines - y + 1
			p.ignoreX = 0
			if p.ignoreY < 0 {
				p.ignoreY = 0
			}
			termbox.Sync()
			p.Draw()
		case '/':
			p.isSlashOn = true
			p.isSearchMode = false
			p.Draw()
		default:
			p.Draw()
		}
	}
	return NO_ACTION
}

func (p *Pager) PollEvent() bool {
	p.Draw()
	for {
		if p.isSlashOn == false {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				ret := p.viewModeKey(ev)
				if ret == 1 {
					return true
				} else if ret == 0 {
					return false
				}
				p.Draw()
			default:
				p.Draw()
			}
		} else if p.isSearchMode {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyEnter:
					// nothing to do
				case termbox.KeyDelete, termbox.KeyCtrlD, termbox.KeyBackspace, termbox.KeyBackspace2:
					p.deleteSearchString()
				case termbox.KeyEsc, termbox.KeyCtrlC:
					p.isSearchMode = false
					p.isSlashOn = false
					p.searchStr = ""
				default:
					if ev.Ch == 'q' {
						p.isSearchMode = false
						p.isSlashOn = false
						p.searchStr = ""
					} else if ev.Ch == 'n' {
						p.searchForward()
					} else if ev.Ch == 'N' {
						p.searchBackward()
					} else {
						p.viewModeKey(ev)
					}
				}
			}
			p.Draw()
		} else { // isSlashOn
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyDelete, termbox.KeyCtrlD, termbox.KeyBackspace, termbox.KeyBackspace2:
					p.deleteSearchString()
				case termbox.KeyEnter:
					p.isSearchMode = true
					p.searchIndex = 1
					p.searchForward()
				case termbox.KeyEsc:
					p.isSlashOn = false
					p.searchStr = ""
				default:
					p.searchStr += string(ev.Ch)
				}
			}
			p.Draw()
		}
	}
	return false
}

func (p *Pager) deleteSearchString() {
	if len(p.searchStr) > 0 {
		p.searchStr = p.searchStr[0 : len(p.searchStr)-1]
	}
}

func (p *Pager) searchString() [][]int {
	return regexp.MustCompile("(?mi)^.*"+p.searchStr+".*$").FindAllStringIndex(regexp.MustCompile("\\033\\[\\d+\\[m(.+?)0m").ReplaceAllString(p.str, "$1"), -1)
}

func (p *Pager) searchForward() {
	matched := p.searchString()
	if len(matched) >= p.searchIndex {
		p.ignoreY = p.getLines(p.str[0:matched[p.searchIndex-1][1]]) - 1
		if len(matched) > p.searchIndex {
			p.searchIndex++
		}
	}
}

func (p *Pager) searchBackward() {
	matched := p.searchString()
	if len(matched) > 0 && p.searchIndex > 1 {
		p.searchIndex--
		p.ignoreY = p.getLines(p.str[0:matched[p.searchIndex-1][1]]) - 1
	}
}

func (p *Pager) getLines(s string) (l int) {
	lines := regexp.MustCompile("(?m)^.*$").FindAllString(s, -1)
	l = len(lines)
	return
}

func (p *Pager) scrollDown() {
	lines := p.getLines(p.str)
	_, y := termbox.Size()
	if p.ignoreY < lines-y {
		p.ignoreY++
	}
	p.Draw()
}

func (p *Pager) scrollUp() {
	if p.ignoreY > 0 {
		p.ignoreY--
	}
	p.Draw()
}

func (p *Pager) scrollRight() {
	x, _ := termbox.Size()
	if p.ignoreX < x {
		p.ignoreX++
	}
	p.Draw()
}

func (p *Pager) scrollLeft() {
	if p.ignoreX > 0 {
		p.ignoreX--
	}
	p.Draw()
}

func (p *Pager) Init() {
	err := termbox.Init()
	p.isSlashOn = false
	p.isSearchMode = false
	p.searchIndex = 1
	if err != nil {
		panic(err)
	}
}

func (p *Pager) Close() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()
	termbox.Sync()
	termbox.Close()
}

func (p *Pager) isMaxIndex() bool {
	return len(p.Files) == (p.Index + 1)
}
