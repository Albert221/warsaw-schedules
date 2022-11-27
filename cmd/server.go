package cmd

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"warsaw-schedules.dev/db"
	"warsaw-schedules.dev/web"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the server",
	RunE:  runServer,
}

func runServer(cmd *cobra.Command, args []string) error {
	host := envOrDefault("HOST", "localhost")
	port := envOrDefault("PORT", "8080")

	stopRepo := cmd.Context().Value(stopRepoKey).(db.StopRepository)

	r := newRouter(stopRepo)

	srv := &http.Server{
		Handler:      r,
		Addr:         host + ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Listening on", srv.Addr)
	log.Fatal(srv.ListenAndServe())

	return nil
}

func newRouter(stopRepo db.StopRepository) *mux.Router {
	stopCtrl := web.NewStopController(stopRepo)

	r := mux.NewRouter()
	r.Handle("/stop", stopCtrl.StopsList()).Methods("GET")
	return r
}

func envOrDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
