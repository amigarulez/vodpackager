package packager

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
	"vodpackager/api"
	"vodpackager/helpers"
	"vodpackager/metrics"
)

type VideoPackagerHandler struct {
	// TODO move to endpoint trait
	endpoint string

	// TODO move to Monitoring trait
	requestsMeter *prometheus.CounterVec
	durationMeter *prometheus.HistogramVec

	packager *Packager
}

func CreateVideoPackagerHandler(packager *Packager, appMetrics metrics.Metrics) VideoPackagerHandler {
	return VideoPackagerHandler{packager: packager, endpoint: "segment", requestsMeter: appMetrics.HttpRequestsMeter, durationMeter: appMetrics.HttpDurationMeter}
}

func (h *VideoPackagerHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// TODO support for traceId and some correlationId
	start := time.Now()
	defer func() {
		// Record latency & request count (labels are path & method)
		h.durationMeter.WithLabelValues(h.endpoint, r.Method).Observe(time.Since(start).Seconds())
	}()

	// Parse request
	var req api.PackageVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.requestsMeter.WithLabelValues(h.endpoint, r.Method, fmt.Sprintf("%d", http.StatusBadRequest)).Inc()
		helpers.WriteJSONError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	// This implementation is only PoC and for actual use (at least) the following should be addressed:
	// In real world scenario this service should be split to at least two services - endpoints and processor.
	//
	// The endpoints service would accept the request for packaging video and push the request for processing to the
	// queue. Service would also need to provide endpoint for checking the progress.
	//
	// The processor service would consume the queue and asynchronously process the requests for packaging videos. Retry
	// in case of failure could be performed by rescheduling the message.
	//
	// The next problem is the segmented video is stored locally, in production environment it should push the segmented
	// video files to the CDN.

	packagedVideo, err := h.packager.PackageVideo(req.FileName, "mp4", uint32(req.SegmentLength)) // hardcode mp4 as no other video formats are supported

	if err != nil {
		// It should use different HTTP statuses depending on type of error, but there is no business error hierarchy yet.
		h.requestsMeter.WithLabelValues(h.endpoint, r.Method, fmt.Sprintf("%d", http.StatusInternalServerError)).Inc()
		helpers.WriteJSONError(w, http.StatusInternalServerError, "error packaging video")
		return
	}

	resp := api.HttpResponse[api.PackageVideoResponse]{
		Status: "ok",
		Data:   &api.PackageVideoResponse{FileName: req.FileName, MpdURL: packagedVideo.PackagedVideo.Data.BaseURL},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)

	h.requestsMeter.WithLabelValues(h.endpoint, r.Method, fmt.Sprintf("%d", http.StatusOK)).Inc()

	return
}
