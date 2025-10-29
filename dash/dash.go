package dash

import (
	"encoding/xml"
	"github.com/rs/zerolog/log"
	"io"
	"time"
	"vodpackager/helpers"
	"vodpackager/segmenter"
)

func CreateMpd(videoInfo segmenter.SegmentedVideoTrackInfo, audioInfo segmenter.SegmentedAudioTrackInfo, durationSeconds uint32, baseUrl string) MPD {
	var adaptationSets = []AdaptationSet{
		createMpdAdaptationSetForVideo(videoInfo),
		createMpdAdaptationSetForAudio(audioInfo),
	}

	var period = Period{AdaptationSets: adaptationSets}

	var mpd = MPD{
		Xmlns:                     "urn:mpeg:dash:schema:mpd:2011",
		MinBufferTime:             "PT1.5S",
		Type:                      "static",
		MediaPresentationDuration: helpers.ToISO(time.Duration(durationSeconds) * 1000000000), // Duration is in nanoseconds
		BaseURL:                   baseUrl,
		Period:                    period,
	}

	return mpd
}

// TODO refactoring - use (builder) pattern to remove duplicated code for video and audio
func createMpdAdaptationSetForVideo(info segmenter.SegmentedVideoTrackInfo) AdaptationSet {
	var baseRepresentation = Representation{
		ID:        info.TrackId,
		Bandwidth: info.Bitrate,
		Width:     info.Width,
		Height:    info.Height,
		Codec:     info.Codec,
		SegmentTemplate: SegmentTemplate{
			Timescale:      1,
			Duration:       info.SegmentDuration,
			Media:          info.SegmentName,
			Initialization: info.InitSegmentName,
		},
	}

	return AdaptationSet{
		MimeType:         MIME_VIDEO_MP4, // hardcoded as no other format than mp4 is supported, should resolve mime type instead
		SegmentAlignment: true,
		Representations:  []Representation{baseRepresentation},
	}
}

func createMpdAdaptationSetForAudio(info segmenter.SegmentedAudioTrackInfo) AdaptationSet {
	var baseRepresentation = Representation{
		ID:    info.TrackId,
		Codec: info.Codec,
		SegmentTemplate: SegmentTemplate{
			Timescale:      1,
			Duration:       info.SegmentDuration,
			Media:          info.SegmentName,
			Initialization: info.InitSegmentName,
		},
	}

	return AdaptationSet{
		MimeType:         MIME_AUDIO_MP4, // hardcoded as no other format than mp4 is supported, should resolve mime type instead
		SegmentAlignment: true,
		Representations:  []Representation{baseRepresentation},
	}
}

func WriteMpd(mpd MPD, w io.Writer) error {
	data, err := xml.MarshalIndent(mpd, "", "  ")

	if err != nil {
		log.Err(err).Msg("Error marshalling mpd")
		return err
	}

	xmlHeader := []byte(xml.Header)
	data = append(xmlHeader, data...)

	if _, err := w.Write(data); err != nil {
		log.Err(err).Msg("Error writing mpd file")
		return err
	}

	log.Info().Msgf("successfully wrote mpd: %v\n", string(data))

	return nil
}
