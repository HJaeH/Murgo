package server

import (
	"time"
)

const (
	MaxFrameSlots  uint64 = 3600
	HalfFrameSlots uint64 = MaxFrameSlots / 2
)

type Bandwidth struct {
	frameNo        uint64
	bandwidthTimer int64
	sizeSum        uint64
	idleTimer      int64
	onlineTimer    int64
	fps            uint64 // frame per seconds
	bandwidth      uint32
}

func newBandwidth() *Bandwidth {
	now := nowMicrosec()
	bw := new(Bandwidth)
	bw.frameNo = 1
	bw.bandwidthTimer = now
	bw.sizeSum = 0
	bw.idleTimer = now
	bw.onlineTimer = now
	bw.fps = 0

	return bw
}

func calcBandwidth(bandwidthRecord *Bandwidth, packetSize uint32) {
	now := nowMicrosec()
	elapsed := uint64(now-bandwidthRecord.bandwidthTimer) / 1000000.0 // sec
	if elapsed == 0 {
		return
	}

	bandwidthRecord.frameNo++
	bandwidthRecord.sizeSum += uint64(packetSize)

	bandwidthRecord.bandwidth = uint32(bandwidthRecord.sizeSum / elapsed)
	bandwidthRecord.fps = bandwidthRecord.frameNo / elapsed
}

func nowMicrosec() int64 {
	return int64(time.Now().UnixNano() / int64(time.Microsecond))
}

func (o *Bandwidth) copyFrom(f *Bandwidth) {
	o.frameNo = f.frameNo
	o.bandwidthTimer = f.bandwidthTimer
	o.sizeSum = f.sizeSum
	o.idleTimer = f.idleTimer
	o.onlineTimer = f.onlineTimer
	o.fps = f.fps
	o.bandwidth = f.bandwidth
}

func (o *Bandwidth) reset() {
	now := nowMicrosec()

	o.frameNo = 1
	o.bandwidthTimer = now
	o.sizeSum = 0
	o.fps = 0

}
