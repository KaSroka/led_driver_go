package main

import (
	"math"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wayneashleyberry/truecolor/pkg/color"
)

var wg sync.WaitGroup

func print_intensity(ic chan float32) {
	defer wg.Done()

	for {
		intensity, ok := <-ic
		if !ok {
			return
		}
		i := uint8(math.Round(float64(intensity) * 255))
		print("New intensity ")
		color.Background(i, i, 0).Print("  ")
		println(" =", uint8(intensity*255))
	}
}

func main() {
	log.SetLevel(log.ErrorLevel)
	log.Debug("Start")

	int_chan := make(chan float32)

	wg.Add(1)
	go print_intensity(int_chan)

	lp := NewLedEffectProcessor(int_chan)

	lp.StartEffect(&Blink{1000 * time.Millisecond, 2000 * time.Millisecond, 0}, false)
	time.Sleep(5 * time.Second)

	lp.StartEffect(&Pulse{1000 * time.Millisecond, 2000 * time.Millisecond, 0}, true)
	time.Sleep(5 * time.Second)

	lp.StartEffect(&Blink{100 * time.Millisecond, 200 * time.Millisecond, 0}, false)
	time.Sleep(1 * time.Second)

	lp.StartEffect(&Pulse{1000 * time.Millisecond, 1000 * time.Millisecond, 0}, false)
	time.Sleep(5 * time.Second)

	lp.StartEffect(&Pulse{0, 1000 * time.Millisecond, 0}, false)
	time.Sleep(5 * time.Second)

	lp.Stop()

	wg.Wait()

	log.Debug("Finishing program")
}
