package packager

import "vodpackager/dash"

type PackagedVideo[T any] struct {
	Data T
}

type DashPackagedVideo struct {
	PackagedVideo[dash.MPD]
}

//type HlsPackagedVideo struct {
//	PackagedVideo[hls.HLS]
//}
