package segmenter

type SegmentedMediaInfo struct {
	// This isn't general enough, there can be multiple tracks of the same type for example audio track for different
	// languages.
	// Also, there is no  support for other track types like subtitle tracks.
	// This should be addressed before extending this PoC to the full-blown packager.
	Video    SegmentedVideoTrackInfo
	Audio    SegmentedAudioTrackInfo
	Duration uint32
}

type SegmentedVideoTrackInfo struct {
	SegmentedTrackInfo
	Width   uint32
	Height  uint32
	Bitrate uint32
}

type SegmentedAudioTrackInfo struct {
	SegmentedTrackInfo
	Volume uint32
}

type SegmentedTrackInfo struct {
	TrackId         string
	SegmentDuration uint32
	TimeScale       uint32
	Codec           string
	InitSegmentName string
	SegmentName     string
}
