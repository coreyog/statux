package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/coreyog/statux"

	"github.com/cheggaaa/pb/v3"
)

func main() {
	// seed random
	rand.Seed(time.Now().Unix())

	// watch for ctrl+c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		fmt.Print("\033[?25h") // show cursor
		os.Exit(0)
	}()

	// setup printer
	lineCount := 5
	stat, err := statux.New(lineCount)
	if err != nil {
		panic(err)
	}

	// wait for fibers to finish
	wg := &sync.WaitGroup{}
	wg.Add(lineCount)

	// start fibers, one per line
	lines := stat.BuildLineWriters()
	for i := 0; i < lineCount; i++ {
		go counter(lines[i], wg)
	}

	// wait for fibers to end
	wg.Wait()

	// clean up
	stat.Finish()
	fmt.Println("DONE")
}

func counter(liner io.StringWriter, wg *sync.WaitGroup) {
	temp := `{{bar . "[" "#" ">" " " "]" }}`
	bar := pb.New(100).SetTemplateString(temp).SetMaxWidth(50)

	speed := rand.Intn(3) + 1

	for bar.Current() < bar.Total() {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(25)*speed))

		bar.Increment()
		liner.WriteString(bar.String())
	}

	wg.Done()
}
