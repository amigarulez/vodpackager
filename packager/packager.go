package packager

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"slices"
	"vodpackager/dash"
	"vodpackager/segmenter"
)

type Packager struct {
	baseSourcePath string
	baseTargetPath string
	baseUrl        string
	segmenters     []segmenter.Segmenter
}

func New(baseSourcePath string, baseTargetPath string, baseUrl string, segmenters []segmenter.Segmenter) *Packager {
	return &Packager{baseSourcePath: baseSourcePath, baseTargetPath: baseTargetPath, baseUrl: baseUrl, segmenters: segmenters}
}

// Segment a video and prepares MPEG-DASH manifest. As this is PoC only this implementation doesn't support other
// format than MPEG-DASH.
func (p *Packager) PackageVideo(fileName string, format string, segmentDurationSeconds uint32) (*DashPackagedVideo, error) {
	s := p.resolveSegmenter(format)
	if s == nil {
		return nil, errors.New(fmt.Sprintf("unuspported format %s", format))
	}

	segmentedInfo, err := p.segmentVideo(*s, fileName, segmentDurationSeconds)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error segmenting video %s", err))
	}

	mpd := dash.CreateMpd(segmentedInfo.Video, segmentedInfo.Audio, segmentedInfo.Duration, p.baseUrl)

	mpdWriter, err := os.Create(fmt.Sprintf("%s/manifest.mpd", p.baseTargetPath))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error creating manifest.mpd %s", err))
	}

	err = dash.WriteMpd(mpd, mpdWriter)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error writing manifest.mpd %s", err))
	}

	packagedVideo := DashPackagedVideo{PackagedVideo: PackagedVideo[dash.MPD]{Data: mpd}}

	return &packagedVideo, nil
}

func (p *Packager) segmentVideo(s segmenter.Segmenter, fileName string, segmentDurationSeconds uint32) (*segmenter.SegmentedMediaInfo, error) {
	ifd, err := os.Open(fmt.Sprintf("%s/%s", p.baseSourcePath, fileName))
	defer ifd.Close()

	if err != nil {
		log.Err(err).Msgf("Error reading file: %s\n", fileName)
		return nil, err
	}

	segmentedInfo, err := s.SegmentVideo(ifd, p.baseTargetPath, segmentDurationSeconds)
	if err != nil {
		log.Err(err).Msgf("Error segmenting file: %s\n", fileName)
		return nil, err
	}

	return segmentedInfo, nil
}

func (p *Packager) resolveSegmenter(format string) *segmenter.Segmenter {
	for _, s := range p.segmenters {
		if slices.Contains(s.SupportedFormats(), format) {
			return &s
		}
	}

	return nil
}
