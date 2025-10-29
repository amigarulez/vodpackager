# vodpackager
Proof of Concept for VOD packager - segments a video (only video and audio tracks, no subtitles) to n-second segments and creates MPEG-DASH manifest.
- JSON API
- Uses go-chi as HTTP server
- Only supported format is MP4
- Only supported streaming standard is MPEG-DASH
- Prometheus metrics
- Zerolog library for logging