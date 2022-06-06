package main

import (
	"math"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type LedEffect interface {
	IsRunning() bool
	Step(time.Duration) float32
	Reset()
}

type Blink struct {
	off    time.Duration
	period time.Duration
	state  time.Duration
}

type Pulse struct {
	off    time.Duration
	period time.Duration
	state  time.Duration
}

type ledRequest struct {
	eff    LedEffect
	repeat bool
}

type LedEffectProcessor struct {
	req            chan ledRequest
	wg             *sync.WaitGroup
	last_int       float32
	intensity_chan chan float32
}

func (b *Blink) Step(progress time.Duration) float32 {
	b.state += progress

	if b.state > b.period {
		b.state = b.period
	}

	if b.state < b.off {
		return 1.0
	}
	return 0.0
}

func (b *Blink) Reset() {
	b.state = 0
}

func (b Blink) IsRunning() bool {
	return b.state < b.period
}

func (b *Pulse) Step(progress time.Duration) float32 {
	b.state += progress

	if b.state > b.period {
		b.state = b.period
	}

	if b.state < b.off {
		return float32(b.state.Seconds() / b.off.Seconds())
	} else if (b.period - b.off) <= 0 { // off time is equal or longer than period
		return 1.0
	}
	return 1.0 - float32((b.state-b.off).Seconds()/(b.period-b.off).Seconds())
}

func (b *Pulse) Reset() {
	b.state = 0
}

func (b Pulse) IsRunning() bool {
	return b.state < b.period
}

func NewLedEffectProcessor(intensity_chan chan float32) LedEffectProcessor {
	c := make(chan ledRequest)
	var wg sync.WaitGroup

	wg.Add(1)

	proc := LedEffectProcessor{c, &wg, float32(math.NaN()), intensity_chan}

	go func() {
		defer wg.Done()
		defer close(intensity_chan)

		duration := 100 * time.Millisecond

		ticker := time.NewTicker(duration)
		ticker.Stop()

		var eff *LedEffect = nil
		var repeat bool = false

		for {
			select {
			case <-ticker.C:
				// log.debug("Tick:", a)
				if eff != nil {
					proc.updateEffect(*eff, duration)
					if !(*eff).IsRunning() {
						if repeat {
							log.Debug("Restarting effect")
							(*eff).Reset()
							proc.updateEffect(*eff, 0)
						} else {
							log.Debug("Ending effect")
							ticker.Stop()
							eff = nil
						}
					}
				}

			case req, ok := <-c:
				if !ok {
					log.Info("Input channel closed, ending")
					return
				}
				log.Info("New effect:", req)
				eff = &req.eff
				repeat = req.repeat
				ticker.Reset(duration)
				proc.updateEffect(*eff, 0)
			}
		}
	}()

	return proc
}

func (lp *LedEffectProcessor) updateEffect(eff LedEffect, duration time.Duration) {
	intensity := eff.Step(duration)
	if lp.last_int != intensity {
		lp.intensity_chan <- intensity
		lp.last_int = intensity
	}
}

func (lp LedEffectProcessor) StartEffect(eff LedEffect, repeat bool) {
	lp.req <- ledRequest{eff, repeat}
}

func (lp LedEffectProcessor) Stop() {
	close(lp.req)
	lp.wg.Wait()
}
