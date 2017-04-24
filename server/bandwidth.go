package server

import "time"

type BandWidth struct {
	frame_no int
	size_sum int
	size_sum2 int
	bandwidth int

	bandwidth_timer time.Time
	bandwidth_timer2 time.Time

	idle_timer time.Time
	online_timer time.Time
}


func NewBandWidth()(*BandWidth){
	bandWidth := new(BandWidth)
	bandWidth.frame_no = 1
	bandWidth.size_sum = 0
	bandWidth.size_sum2 = 0
	bandWidth.bandwidth = 0
	bandWidth.bandwidth_timer = time.Now()
	bandWidth.bandwidth_timer2 = time.Now()
	bandWidth.idle_timer = time.Now()

	return bandWidth
}

