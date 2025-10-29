package mp4

import "github.com/Eyevinn/mp4ff/mp4"

type syncPoint struct {
	sampleNr         uint32
	decodeTime       uint64
	presentationTime uint64
}

type sampleInterval struct {
	startNr uint32
	endNr   uint32 // included in interval
}

type track struct {
	trackType string
	inTrak    *mp4.TrakBox
	timeScale uint32
	trackID   uint32 // trackID in segmented output
	segments  []sampleInterval
}

type segmentedTrack struct {
	trackID  uint32
	segments []*mp4.MediaSegment
}
