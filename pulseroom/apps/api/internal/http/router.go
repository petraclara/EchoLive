package http

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pulseroom/api/internal/config"
	"github.com/pulseroom/api/internal/http/handlers"
	authmw "github.com/pulseroom/api/internal/http/middleware"
	"github.com/pulseroom/api/internal/repository"
	"github.com/pulseroom/api/internal/ws"
)

type Server struct {
	Router *chi.Mux
}

func NewServer(cfg config.Config, store *repository.Store, hub *ws.Hub) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	webURL := config.PublicWebURL()
	apiURL := config.PublicAPIURL()
	corsOrigins := []string{cfg.CORSOrigin, webURL}
	if extra := os.Getenv("CORS_ORIGINS"); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			if o = strings.TrimSpace(o); o != "" {
				corsOrigins = append(corsOrigins, o)
			}
		}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	authH := &handlers.AuthHandler{Store: store, JWTSecret: cfg.JWTSecret}
	eventH := &handlers.EventHandler{Store: store, WebAppURL: webURL}
	annH := &handlers.AnnouncementHandler{Store: store, Hub: hub}
	joinH := &handlers.JoinHandler{Store: store, APIURL: apiURL, WebAppURL: webURL}
	resH := &handlers.ResourceHandler{Store: store, Hub: hub}
	wsH := &handlers.WSHandler{
		Store: store, Hub: hub, JWTSecret: cfg.JWTSecret,
		AllowedOrigins: corsOrigins,
	}

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/v1", func(r chi.Router) {
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)

		r.Post("/join", joinH.Join)
		r.Get("/events/{eventID}/public", joinH.Public)

		r.Group(func(r chi.Router) {
			r.Use(authmw.OrganizerAuth(cfg.JWTSecret))
			r.Get("/auth/me", authH.Me)

			r.Get("/events", eventH.List)
			r.Post("/events", eventH.Create)
			r.Get("/events/{eventID}", eventH.Get)
			r.Patch("/events/{eventID}", eventH.Update)
			r.Delete("/events/{eventID}", eventH.Delete)
			r.Get("/events/{eventID}/qr", eventH.QR)

			r.Get("/events/{eventID}/announcements", annH.List)
			r.Post("/events/{eventID}/announcements", annH.Create)
			r.Post("/events/{eventID}/announcements/{announcementID}/pin", annH.Pin)
			r.Delete("/events/{eventID}/announcements/{announcementID}", annH.Delete)

			r.Get("/events/{eventID}/resources", resH.List)
			r.Post("/events/{eventID}/resources", resH.Create)
			r.Delete("/events/{eventID}/resources/{resourceID}", resH.Delete)

			r.Get("/events/{eventID}/agenda", resH.ListAgenda)
			r.Post("/events/{eventID}/agenda", resH.CreateAgenda)
			r.Delete("/events/{eventID}/agenda/{itemID}", resH.DeleteAgenda)
		})
	})

	r.Get("/ws/events/{eventID}", wsH.Serve)

	return &Server{Router: r}
}
