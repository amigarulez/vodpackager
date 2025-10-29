package segmenter

import (
	"io"
)

type Segmenter interface {
	SupportedFormats() []string
	SegmentVideo(file io.ReadSeeker, outDirPath string, segmentDurationSeconds uint32) (*SegmentedMediaInfo, error)
}
