package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"net/http"
	"vodpackager/logging"
	"vodpackager/metrics"
	"vodpackager/mp4"
	"vodpackager/packager"
	"vodpackager/segmenter"
	"vodpackager/status"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	// ---------- Logging ----------
	logging.InitializeLogging()

	// ---------- Metrics ----------
	appMetrics := metrics.InitializeMetrics()

	router := chi.NewRouter()

	// ---------- Standard chi middlewares ----------
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)    // logs each request to stdout
	router.Use(middleware.Recoverer) // fallback recover (we also have our own)
	// TODO exception handling
	//router.Use(recoveryMiddleware)   // custom JSON‑error recovery

	// ---------- Application logic ----------
	mp4Segmenter := mp4.MP4Segmenter{}
	dashPackager := packager.New("./.res", "./.run", "http://localhost", []segmenter.Segmenter{&mp4Segmenter})

	// ---------- Application handlers ----------
	packagerHandler := packager.CreateVideoPackagerHandler(dashPackager, appMetrics)

	// ---------- Application routes ----------
	router.Post("/package", packagerHandler.Handle) // TODO requires authentication
	router.Get("/status", status.StatusHandler)
	router.Get("/statusEx", status.StatusExHandler) // TODO requires authentication and/or nginx rule for IP address whitelisting

	// ---------- Prometheus endpoint ----------
	router.Handle("/metrics", promhttp.Handler()) // TODO requires nginx rule for IP address whitelisting, authentication wouldn't be practical

	// ---------- Server start ----------
	addr := ":8080"
	log.Info().Msgf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}
