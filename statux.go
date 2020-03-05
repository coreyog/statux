package statux

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/buger/goterm"
	"golang.org/x/crypto/ssh/terminal"
)

type Statux struct {
	count    int
	line     int
	maxWidth int
	mutex    sync.Mutex
	finished bool
}

type LineWriter struct {
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

func (stat *Statux) Write(index int, data []byte) (n int, err error) {
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

	bytes.ReplaceAll(data, []byte{'\n'}, []byte{' '})

	if len(data) > stat.maxWidth {
		data = data[:stat.maxWidth-1]
		data = append(data, '$')
	}

	if len(data) < stat.maxWidth {
		data = append(data, bytes.Repeat([]byte(" "), stat.maxWidth-len(data)-1)...)
	}

	return fmt.Printf("%s", data)
}

func (stat *Statux) BuildLineWriters() (lines []LineWriter) {
	if stat.finished {
		return nil
	}

	lines = make([]LineWriter, stat.count)
	for i := 0; i < stat.count; i++ {
		lines[i] = LineWriter{
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
	goterm.Println("\033[?25h") // show cursor
	goterm.Flush()

	stat.finished = true
}

func (stat *Statux) IsFinished() bool {
	return stat.finished
}

func (line LineWriter) Write(data []byte) (n int, err error) {
	return line.parent.Write(line.line, data)
}

func getConsoleSize() (x int, y int, err error) {
	return terminal.GetSize(int(os.Stdout.Fd()))
}
