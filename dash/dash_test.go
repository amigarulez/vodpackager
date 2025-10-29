package dash

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"vodpackager/segmenter"
)

var DURATION_9m_56s uint32 = (9*60 + 56)

func TestHappyDash(t *testing.T) {
	videoTrackInfo := segmenter.SegmentedVideoTrackInfo{
		SegmentedTrackInfo: segmenter.SegmentedTrackInfo{
			TrackId:         "1",
			SegmentDuration: 5,
			TimeScale:       1,
			Codec:           "avc1.640028",
			InitSegmentName: "init-$RepresentationID$.m4s",
			SegmentName:     "segment-$RepresentationID$-$Number%05d$.m4s",
		},
		Width:   1920,
		Height:  1080,
		Bitrate: 2972886,
	}

	audioTrackInfo := segmenter.SegmentedAudioTrackInfo{
		SegmentedTrackInfo: segmenter.SegmentedTrackInfo{
			TrackId:         "2",
			SegmentDuration: 5,
			TimeScale:       1,
			Codec:           "mp4a.40.2",
			InitSegmentName: "init-$RepresentationID$.m4s",
			SegmentName:     "segment-$RepresentationID$-$Number%05d$.m4s",
		},
		Volume: 1,
	}

	mpd := CreateMpd(videoTrackInfo, audioTrackInfo, DURATION_9m_56s, "http://localhost:8080/out")

	ifd, err := os.Create("../.test/out/manifest.mpd")
	assert.Nil(t, err)
	defer ifd.Close()

	err = WriteMpd(mpd, ifd)
	assert.Nil(t, err)
}
