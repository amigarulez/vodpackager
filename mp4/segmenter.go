package mp4

import (
	"errors"
	"fmt"
	"github.com/Eyevinn/mp4ff/mp4"
	"github.com/rs/zerolog/log"
	"io"
	"vodpackager/segmenter"
)

type MP4Segmenter struct {
}

func (s *MP4Segmenter) SupportedFormats() []string {
	return []string{"mp4"}
}

func (s *MP4Segmenter) SegmentVideo(file io.ReadSeeker, outDirPath string, segmentDurationSeconds uint32) (*segmenter.SegmentedMediaInfo, error) {

	var mp4File *mp4.File
	mp4File, err := mp4.DecodeFile(file)
	if err != nil {
		return nil, err
	}

	videoTrackInfo, err := getVideoTrackInfo(mp4File, segmentDurationSeconds)
	if err != nil {
		return nil, err
	}

	if videoTrackInfo == nil {
		return nil, errors.New("videoTrackInfo is nil")
	}

	audioTrackInfo, err := getAudioTrackInfo(mp4File, segmentDurationSeconds)
	if err != nil {
		return nil, err
	}

	if audioTrackInfo == nil {
		return nil, errors.New("audioTrackInfo is nil")
	}

	err = createAndWriteInitSegments(mp4File, outDirPath)
	if err != nil {
		return nil, err
	}

	err = createAndWriteSegments(mp4File, file, outDirPath, segmentDurationSeconds*1000)
	if err != nil {
		return nil, err
	}

	durationSeconds := uint32(mp4File.Moov.Mvhd.Duration) / mp4File.Moov.Mvhd.Timescale

	return &segmenter.SegmentedMediaInfo{Duration: durationSeconds, Video: *videoTrackInfo, Audio: *audioTrackInfo}, nil
}

func getVideoTrackInfo(mp4File *mp4.File, segmentDurationSeconds uint32) (*segmenter.SegmentedVideoTrackInfo, error) {
	var info *segmenter.SegmentedVideoTrackInfo
	for _, trak := range mp4File.Moov.Traks {
		trackType, err := getTrackType(trak.Mdia.Hdlr)

		if err != nil {
			return nil, err
		}

		if trackType == "video" {
			info = &segmenter.SegmentedVideoTrackInfo{
				SegmentedTrackInfo: segmenter.SegmentedTrackInfo{
					TrackId:         fmt.Sprintf("%d", trak.Tkhd.TrackID),
					SegmentDuration: segmentDurationSeconds,
					TimeScale:       1,
					Codec:           "avc1.640028",
					InitSegmentName: "init-$RepresentationID$.m4s",
					SegmentName:     "segment-$RepresentationID$-$Number%05d$.m4s",
				},
				// convert Fixed32 to uint32, see Fixed32.String()
				Width:   uint32(trak.Tkhd.Width) >> 16,
				Height:  uint32(trak.Tkhd.Height) >> 16,
				Bitrate: 2972886, // TODO
			}
		}
	}

	return info, nil
}

func getAudioTrackInfo(mp4File *mp4.File, segmentDurationSeconds uint32) (*segmenter.SegmentedAudioTrackInfo, error) {
	var info *segmenter.SegmentedAudioTrackInfo
	for _, trak := range mp4File.Moov.Traks {
		trackType, err := getTrackType(trak.Mdia.Hdlr)

		if err != nil {
			return nil, err
		}

		if trackType == "audio" {
			info = &segmenter.SegmentedAudioTrackInfo{
				SegmentedTrackInfo: segmenter.SegmentedTrackInfo{
					TrackId:         fmt.Sprintf("%d", trak.Tkhd.TrackID),
					SegmentDuration: segmentDurationSeconds,
					TimeScale:       1,
					Codec:           "mp4a.40.2",
					InitSegmentName: "init-$RepresentationID$.m4s",
					SegmentName:     "segment-$RepresentationID$-$Number%05d$.m4s",
				},
				Volume: uint32(trak.Tkhd.Volume),
			}
		}
	}

	return info, nil
}

// Create initialization segments for all tracks
func createAndWriteInitSegments(mp4File *mp4.File, outDirPath string) error {
	var initSegments []*mp4.InitSegment
	for _, trak := range mp4File.Moov.Traks {
		is := mp4.CreateEmptyInit()
		is.Moov.Mvhd.Timescale = mp4File.Moov.Mvhd.Timescale
		is.Moov.Mvex.AddChild(&mp4.MehdBox{FragmentDuration: int64(mp4File.Moov.Mvhd.Duration)})

		trackType, err := getTrackType(trak.Mdia.Hdlr)
		if err != nil {
			return err
		}

		is.AddEmptyTrack(trak.Mdia.Mdhd.Timescale, trackType, trak.Mdia.Mdhd.GetLanguage())

		is.Moov.Trak.Tkhd.TrackID = trak.Tkhd.TrackID

		inStsd := trak.Mdia.Minf.Stbl.Stsd
		outStsd := is.Moov.Trak.Mdia.Minf.Stbl.Stsd

		switch trackType {
		case "audio":
			if inStsd.Mp4a != nil {
				outStsd.AddChild(inStsd.Mp4a)
			} else if inStsd.AC3 != nil {
				outStsd.AddChild(inStsd.AC3)
			} else if inStsd.EC3 != nil {
				outStsd.AddChild(inStsd.EC3)
			}
		case "video":
			if inStsd.AvcX != nil {
				outStsd.AddChild(inStsd.AvcX)
			} else if inStsd.HvcX != nil {
				outStsd.AddChild(inStsd.HvcX)
			}
		default:
			return fmt.Errorf("unsupported tracktype: %s", trackType)
		}

		initSegments = append(initSegments, is)
	}

	for _, initSegment := range initSegments {
		outPath := fmt.Sprintf("%s/init-%d.m4s", outDirPath, initSegment.Moov.Trak.Tkhd.TrackID)
		err := mp4.WriteToFile(initSegment, outPath)
		if err != nil {
			log.Err(err).Msg("Error writing init segment")
			// TODO cleanup partial data
			return err
		}
	}

	return nil
}

func createAndWriteSegments(mp4File *mp4.File, rs io.ReadSeeker, outDirPath string, segmentDurationMs uint32) error {
	tracks := make([]track, 0, len(mp4File.Moov.Traks))

	timeScale, syncPoints, err := getSegmentStartsFromVideo(mp4File, segmentDurationMs)

	numberOfSegments := len(syncPoints)

	if err != nil {
		return err
	}

	for _, trak := range mp4File.Moov.Traks {

		segmentIntervals, err := getSegmentIntervals(timeScale, syncPoints, trak)
		if err != nil {
			return err
		}

		trackType, err := getTrackType(trak.Mdia.Hdlr)
		if err != nil {
			return err
		}

		segmentTrack := track{
			trackType: trackType,
			inTrak:    trak,
			timeScale: timeScale,
			trackID:   trak.Tkhd.TrackID,
			segments:  segmentIntervals,
		}

		tracks = append(tracks, segmentTrack)
	}

	for _, track := range tracks {
		log.Info().Msgf("Writing segments for track %d\n", track.trackID)

		segmentNo := 1
		for {
			startSampleNr, endSampleNr := track.segments[segmentNo-1].startNr, track.segments[segmentNo-1].endNr
			fullSamples, err := getFullSamplesForInterval(mp4File, &track, startSampleNr, endSampleNr, rs)

			if err != nil {
				return err
			}

			if len(fullSamples) == 0 {
				log.Debug().Msgf("No more samples for %s\n", track.trackType)
				continue
			}

			segment := mp4.NewMediaSegment()
			fragment, err := mp4.CreateFragment(uint32(segmentNo), track.trackID)
			if err != nil {
				return err
			}

			segment.AddFragment(fragment)

			for _, fullSample := range fullSamples {
				err = fragment.AddFullSampleToTrack(fullSample, track.trackID)
				if err != nil {
					return err
				}
			}

			outPath := fmt.Sprintf("%s/segment-%d-%05d.m4s", outDirPath, track.trackID, segmentNo)
			err = mp4.WriteToFile(segment, outPath)
			if err != nil {
				// TODO cleanup partial data
				return err
			}
			log.Debug().Msgf("Data segment written to `%s`\n", outPath)

			segmentNo++
			if segmentNo > numberOfSegments {
				log.Info().Msgf("Writing segments for track %d completed\n", track.trackID)
				break
			}
		}
	}

	return nil
}

func getFullSamplesForInterval(file *mp4.File, track *track, startSampleNo uint32, endSampleNo uint32, rs io.ReadSeeker) ([]mp4.FullSample, error) {
	stbl := track.inTrak.Mdia.Minf.Stbl
	mdat := file.Mdat                                // the 'mdat' box holds the raw media data
	mdatPayloadStart := mdat.PayloadAbsoluteOffset() // absolute file offset where mdat payload begins

	samples := make([]mp4.FullSample, 0, endSampleNo-startSampleNo+1)
	for sampleNo := startSampleNo; sampleNo <= endSampleNo; sampleNo++ {

		// Find the chunk that contains the current sample and also sample number of the first sample in the chunk
		// Stsc maps a sample index -> the chunk index
		chunkNr, sampleNrAtChunkStart, err := stbl.Stsc.ChunkNrFromSampleNr(int(sampleNo))
		if err != nil {
			return nil, err
		}

		// Compute the byte offset of the first sample in the chunk
		// MP4 can store chunk offsets either in a 32‑bit (stco) or a 64‑bit (co64) table.
		// We pick whichever is present and fetch the absolute file offset of the first byte of the chunk.
		var offset int64
		if stbl.Stco != nil { // 32‑bit chunk offsets
			offset = int64(stbl.Stco.ChunkOffset[chunkNr-1])
		} else if stbl.Co64 != nil { // 64‑bit chunk offsets
			offset = int64(stbl.Co64.ChunkOffset[chunkNr-1])
		}

		// Advance the offset to the exact sample within the chunk
		// stsz holds the size of every individual sample.
		// Starting at the first sample of the chunk, we sum the sizes of all preceding samples until we reach the target sampleNr.
		// After the loop, offset points to the first byte of the desired sample.
		for sNr := sampleNrAtChunkStart; sNr < int(sampleNo); sNr++ {
			offset += int64(stbl.Stsz.GetSampleSize(sNr))
		}

		// Read the raw sample data
		// If the mdat box is “lazy”, we seek to the computed file offset and read exactly size bytes.
		// Otherwise we calculate the relative offset inside the already‑loaded mdat.Data slice and take a sub‑slice—no extra I/O.
		size := stbl.Stsz.GetSampleSize(int(sampleNo))              // how many bytes to read.
		decodingTime, duration := stbl.Stts.GetDecodeTime(sampleNo) // decoding timestamp (in the track’s time‑scale).
		var compositionTimeOffset int32 = 0                         // composition‑time offset (used for B‑frames, may be zero).
		if stbl.Ctts != nil {
			compositionTimeOffset = stbl.Ctts.GetCompositionTimeOffset(sampleNo)
		}

		var sampleData []byte
		if mdat.GetLazyDataSize() > 0 { // data not yet in memory -> read from file
			_, err := rs.Seek(offset, io.SeekStart)
			if err != nil {
				return nil, err
			}
			sampleData = make([]byte, size)
			_, err = io.ReadFull(rs, sampleData)
			if err != nil {
				return nil, err
			}
		} else { // data already in memory -> slice it
			offsetInMdatData := uint64(offset) - mdatPayloadStart
			sampleData = mdat.Data[offsetInMdatData : offsetInMdatData+uint64(size)]
		}

		// Build the FullSample struct
		sc := mp4.FullSample{
			Sample: mp4.Sample{
				Flags:                 translateSampleFlagsForFragment(stbl, sampleNo),
				Size:                  size,
				Dur:                   duration,
				CompositionTimeOffset: compositionTimeOffset,
			},
			DecodeTime: decodingTime,
			Data:       sampleData,
		}

		samples = append(samples, sc)
	}

	return samples, nil
}

// TranslateSampleFlagsForFragment - translate sample flags from stss and sdtp to what is needed in trun
func translateSampleFlagsForFragment(stbl *mp4.StblBox, sampleNr uint32) (flags uint32) {
	var sampleFlags mp4.SampleFlags
	if stbl.Stss != nil {
		isSync := stbl.Stss.IsSyncSample(uint32(sampleNr))
		sampleFlags.SampleIsNonSync = !isSync
		if isSync {
			sampleFlags.SampleDependsOn = 2 //2 == does not depend on others (I-picture). May be overridden by sdtp entry
		}
	}
	if stbl.Sdtp != nil {
		entry := stbl.Sdtp.Entries[uint32(sampleNr)-1] // table starts at 0, but sampleNr is one-based
		sampleFlags.IsLeading = entry.IsLeading()
		sampleFlags.SampleDependsOn = entry.SampleDependsOn()
		sampleFlags.SampleHasRedundancy = entry.SampleHasRedundancy()
		sampleFlags.SampleIsDependedOn = entry.SampleIsDependedOn()
	}
	return sampleFlags.Encode()
}

// Translates a wall‑clock segment duration into the native time‑unit space of the video track and discovers the exact
// sample numbers that can safely start each segment.
func getSegmentStartsFromVideo(parsedMp4 *mp4.File, segmentDurationMs uint32) (uint32, []syncPoint, error) {
	var refTrak *mp4.TrakBox
	for _, trak := range parsedMp4.Moov.Traks {
		hdlrType := trak.Mdia.Hdlr.HandlerType
		if hdlrType == "vide" {
			refTrak = trak
			break
		}
	}

	if refTrak == nil {
		return 0, nil, fmt.Errorf("Cannot handle media files with no video track")
	}

	stts := refTrak.Mdia.Minf.Stbl.Stts
	stss := refTrak.Mdia.Minf.Stbl.Stss
	ctts := refTrak.Mdia.Minf.Stbl.Ctts

	timeScale := refTrak.Mdia.Mdhd.Timescale
	var segmentStep = uint32(uint64(segmentDurationMs) * uint64(timeScale) / 1000) // Segment duration in timescale units

	syncPoints := make([]syncPoint, 0, stss.EntryCount())

	var nextSegmentStart uint32 = 0
	for _, sampleNr := range stss.SampleNumber {

		decodeTime, _ := stts.GetDecodeTime(sampleNr) // Time deltas between successive decoding times
		presentationTime := int64(decodeTime)         // The moment the decoded picture is actually shown

		if ctts != nil {
			presentationTime += int64(ctts.GetCompositionTimeOffset(sampleNr))
		}
		// Ensure that a new segment starts at the first presentation‑ready sync sample whose presentation timestamp
		// reaches or exceeds the target segment boundary
		if presentationTime >= int64(nextSegmentStart) {
			// This sync sample is the first one that will be shown at or after the desired segment start time.
			syncPoints = append(syncPoints, syncPoint{sampleNr, decodeTime, uint64(presentationTime)})

			// Move to the next segment boundary
			nextSegmentStart += segmentStep
		}
	}
	return timeScale, syncPoints, nil
}

// The function returns a slice of SampleInterval structs, each describing the range of sample numbers that belong to a
// particular segment, plus an error if something goes wrong while looking up timestamps.
// [syncTimescale] The timescale (ticks/second) that was used when the SyncPoints were computed (normally the video track’s timescale).
// [syncPoints] Points that mark the first key‑frame of each segment
// [trak] Pointer to the video track
func getSegmentIntervals(syncTimescale uint32, syncPoints []syncPoint, trak *mp4.TrakBox) ([]sampleInterval, error) {
	totNrSamples := trak.Mdia.Minf.Stbl.Stsz.SampleNumber

	var startSampleNr uint32 = 1     // The first interval always starts at sample #1 (MP4 samples are 1‑based).
	var nextStartSampleNr uint32 = 0 // Will hold the first sample of the *next* interval
	var endSampleNr uint32           // Last sample of the current interval
	var err error

	// One interval per sync point.
	sampleIntervals := make([]sampleInterval, len(syncPoints))

	for i := range syncPoints {
		// If we already know where the *next* segment begins, that becomes
		// the start of the current interval.
		if nextStartSampleNr != 0 {
			startSampleNr = nextStartSampleNr
		}
		// Determine the end of the current interval.
		if i == len(syncPoints)-1 {
			// Last segment -> runs to the very last sample of the track.
			endSampleNr = totNrSamples - 1
		} else {
			// Decode‑time (DTS) of the *next* segment’s first sync sample.
			nextSyncStart := syncPoints[i+1].decodeTime

			// Convert that DTS from the sync‑point timescale to the track’s own timescale.
			// (Both are linear tick counts, so a simple ratio works.)
			nextStartTime := nextSyncStart * uint64(trak.Mdia.Mdhd.Timescale) / uint64(syncTimescale)

			// Map the track‑timescale timestamp to a concrete sample number.
			// GetSampleNrAtTime returns the first sample whose decode time >= nextStartTime.
			nextStartSampleNr, err = trak.Mdia.Minf.Stbl.Stts.GetSampleNrAtTime(nextStartTime)

			if err != nil {
				return nil, err
			}

			// The current interval ends right before the first sample of the next segment.
			endSampleNr = nextStartSampleNr - 1
		}
		// Store the interval for this segment.
		sampleIntervals[i] = sampleInterval{startSampleNr, endSampleNr}
	}

	return sampleIntervals, nil
}
