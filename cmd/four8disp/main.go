package main

import (
	"flag"
	"log"
	"time"

	"github.com/upsampled/mmgpio"
	"github.com/upsampled/mmgpio/foursegdisp"
)

func main() {
	seg := flag.Int("seg", -1, "set the segment number")
	run := flag.Int("run", -1, "Run display driver lightig up each display for x ms. If no value for counter given, displays 12.34 for give duration")
	dur := flag.Int("dur", 5, "Duration to run for (used with run)")
	cnt := flag.Int("count", -1, "Count, one a second, till given number, driver timeout must be set (runoption)")

	flag.Parse()

	rp := mmgpio.NewRaspMMGPIO(mmgpio.RASP_ZERO)

	mm := foursegdisp.NewFourEightSegs(rp)

	err := mm.Init([...]int{9, 13, 17, 3, 2, 11, 27}, [...]int{10, 5, 6, 22}, 4)
	if err != nil {
		log.Fatal("Init Error: " + err.Error())
	}

	if *seg >= 0 && *seg < 9 {
		mm.AllDigsOn()
		mm.SetDigsSegs(*seg)
		mm.DeInit()
	}

	if *run > 0 {
		mm.SetNumsDots([...]uint32{1, 2, 3, 4}, [...]uint32{0, 1, 0, 0})
		done := mm.Run(*run)
		time.Sleep(time.Duration(*dur) * time.Second)
		mm.Stop()
		<-done
		mm.AllDigsOff()
		mm.DeInit()
	}

	if *cnt > 0 && *run > 0 {
		ctr := uint32(0)
		done := mm.Run(*run)
		for i := 0; i < *cnt+1; i++ {
			mm.SetNumsDots([...]uint32{(ctr / 1000) % 10, (ctr / 100) % 10, (ctr / 10) % 10, ctr % 10}, [...]uint32{0, 0, 0, 0})
			ctr++
			time.Sleep(time.Second)
		}
		mm.Stop()
		<-done
		mm.AllDigsOff()
		mm.DeInit()
	}

}
