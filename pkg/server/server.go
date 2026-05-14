package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/EthanArc/go_final_project/pkg/api"
)

const (
	webDir      = "./web" //static frontend assets directory
	DefaultAddr = ":7540" //default port
)

type Server struct {
	Logger *slog.Logger // pointer to custom logger
	Server *http.Server //Goland standart server
}

// Starting custom server inctance
func StartServer(logger *slog.Logger) *Server {

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(webDir))) //regestring file server

	api.Init(mux) //regestring backend API endpoints
	server := &http.Server{
		Addr:              DefaultAddr,
		Handler:           mux,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       15 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	return &Server{
		Logger: logger,
		Server: server,
	}
}
