package dash

import "encoding/xml"

/**
 * This is an AI generated subset of MPEG-MPD definition. For production generate these structs from MPEG-MPD schema.
 */
type MPD struct {
	XMLName                   xml.Name `xml:"MPD"`
	Xmlns                     string   `xml:"xmlns,attr"`
	MinBufferTime             string   `xml:"minBufferTime,attr"`
	Type                      string   `xml:"type,attr"`
	MediaPresentationDuration string   `xml:"mediaPresentationDuration,attr"`
	BaseURL                   string   `xml:"BaseURL"`
	Period                    Period   `xml:"Period"`
}

type Period struct {
	AdaptationSets []AdaptationSet `xml:"AdaptationSet"`
}

type AdaptationSet struct {
	MimeType         string           `xml:"mimeType,attr"`
	SegmentAlignment bool             `xml:"segmentAlignment,attr"`
	Representations  []Representation `xml:"Representation"`
}

type Representation struct {
	ID        string `xml:"id,attr"`
	Bandwidth uint32 `xml:"bandwidth,attr"`
	Width     uint32 `xml:"width,attr,omitempty"`
	Height    uint32 `xml:"height,attr,omitempty"`
	Codec     string `xml:"codecs,attr"`
	//BaseURL         string          `xml:"BaseURL"`
	SegmentTemplate SegmentTemplate `xml:"SegmentTemplate"`
}

type SegmentTemplate struct {
	Timescale      uint32 `xml:"timescale,attr"`
	Duration       uint32 `xml:"duration,attr"`
	Media          string `xml:"media,attr"`
	Initialization string `xml:"initialization,attr"`
}
