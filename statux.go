package statux

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/buger/goterm"
	"golang.org/x/term"
)

type Statux struct {
	count    int
	line     int
	maxWidth int
	mutex    sync.Mutex
	finished bool
}

type lineWriter struct {
	parent *Statux
	line   int
}

func New(count int) (stat *Statux, err error) {
	if count <= 0 {
		return nil, errors.New("invalid count")
	}

	x, y, err := getConsoleSize()
	if err != nil {
		return nil, err
	}

	if count > y {
		return nil, fmt.Errorf("terminal too small for %d line(s)", count)
	}

	// make room
	fmt.Print(strings.Repeat("\n", count-1))

	// move to top of space
	goterm.MoveCursorUp(count - 1)
	goterm.Print("\033[?25l") // hide cursor
	goterm.Flush()

	return &Statux{
		count:    count,
		line:     0,
		maxWidth: x,
		mutex:    sync.Mutex{},
		finished: false,
	}, nil
}

func (stat *Statux) WriteString(index int, str string) (n int, err error) {
	if stat.finished {
		return 0, nil
	}

	if index < 0 || index >= stat.count {
		return 0, fmt.Errorf("invalid index: %d", index)
	}

	stat.mutex.Lock()
	defer stat.mutex.Unlock()

	if stat.line > index {
		// need to move up to the target line
		goterm.MoveCursorUp(stat.line - index)
		stat.line = index
	}

	if stat.line < index {
		// need to move down to the target line
		goterm.MoveCursorDown(index - stat.line)
		stat.line = index
	}

	goterm.MoveCursorBackward(stat.maxWidth)
	goterm.Flush()

	str = strings.ReplaceAll(str, "\n", " ")

	if len(str) > stat.maxWidth {
		str = str[:stat.maxWidth-1]
		str += "$"
	}

	if len(str) < stat.maxWidth {
		// data = append(data, bytes.Repeat([]byte(" "), stat.maxWidth-len(data)-1)...)
		str += strings.Repeat(" ", stat.maxWidth-len(str)-1)
	}

	return fmt.Print(str)
}

func (stat *Statux) BuildLineWriters() (lines []io.StringWriter) {
	if stat.finished {
		return nil
	}

	lines = make([]io.StringWriter, stat.count)
	for i := 0; i < stat.count; i++ {
		lines[i] = lineWriter{
			parent: stat,
			line:   i,
		}
	}

	return lines
}

func (stat *Statux) Finish() {
	if stat.finished {
		return
	}

	stat.mutex.Lock()
	defer stat.mutex.Unlock()

	if stat.line < stat.count {
		goterm.MoveCursorDown(stat.count - stat.line)
		goterm.Flush()
	}

	goterm.MoveCursorBackward(stat.maxWidth)
	fmt.Println()
	fmt.Print("\033[?25h") // show cursor
	goterm.Flush()

	stat.finished = true
}

func (stat *Statux) IsFinished() bool {
	return stat.finished
}

func (line lineWriter) WriteString(str string) (n int, err error) {
	return line.parent.WriteString(line.line, str)
}

func (line lineWriter) Write(data []byte) (n int, err error) {
	return line.WriteString(string(data))
}

func getConsoleSize() (x int, y int, err error) {
	return term.GetSize(int(os.Stdout.Fd()))
}
