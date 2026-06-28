package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/veloper/goard/internal"
)

//go:embed web/*
var webFS embed.FS

func main() {
	cfg := loadConfig()

	store, err := internal.NewStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer store.Close()

	if (cfg.AdminUsername == "") != (cfg.AdminPAT == "") {
		log.Fatalf("GOARD_ADMIN_USERNAME and GOARD_ADMIN_PAT must be set together")
	}
	if cfg.AdminUsername == "" {
		log.Fatal("GOARD_ADMIN_USERNAME and GOARD_ADMIN_PAT are required")
	}
	if err := store.EnsureAdmin(cfg.AdminUsername, cfg.AdminPAT); err != nil {
		log.Fatalf("admin: %v", err)
	}

	hub := internal.NewHub()
	go hub.Run()

	handler := internal.NewHandler(store, hub)
	handler.InitMCP(store)
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/info", handler.Info)
	mux.HandleFunc("GET /api/me", handler.Me)
	mux.HandleFunc("GET /api/users", handler.ListUsers)
	mux.HandleFunc("GET /api/users/{id}", handler.GetUser)
	mux.HandleFunc("POST /api/users", handler.CreateUser)
	mux.HandleFunc("PATCH /api/users/{id}", handler.UpdateUser)
	mux.HandleFunc("DELETE /api/users/{id}", handler.DeleteUser)
	mux.HandleFunc("GET /api/users/{id}/pat", handler.GetUserPAT)
	mux.HandleFunc("PUT /api/users/{id}/pat", handler.SetUserPAT)

	mux.HandleFunc("GET /api/projects", handler.ListProjects)
	mux.HandleFunc("POST /api/projects", handler.CreateProject)
	mux.HandleFunc("GET /api/projects/{id}", handler.GetProject)
	mux.HandleFunc("PATCH /api/projects/{id}", handler.UpdateProject)
	mux.HandleFunc("DELETE /api/projects/{id}", handler.DeleteProject)

	mux.HandleFunc("GET /api/projects/{id}/issues", handler.ListIssues)
	mux.HandleFunc("POST /api/projects/{id}/issues", handler.CreateIssue)
	mux.HandleFunc("GET /api/issues/{id}", handler.GetIssue)
	mux.HandleFunc("PATCH /api/issues/{id}", handler.UpdateIssue)
	mux.HandleFunc("PUT /api/issues/{id}/state", handler.UpdateIssueState)
	mux.HandleFunc("DELETE /api/issues/{id}", handler.DeleteIssue)

	mux.HandleFunc("GET /api/issues/{id}/comments", handler.ListComments)
	mux.HandleFunc("POST /api/issues/{id}/comments", handler.CreateComment)
	mux.HandleFunc("GET /api/ws", handler.ServeWs)
	mux.HandleFunc("POST /mcp", handler.ServeMCP)

	// Web UI — serve embedded files
	webSub, _ := fs.Sub(webFS, "web")
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, webSub, "login.html")
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, webSub, "index.html")
	})
	mux.HandleFunc("GET /projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, webSub, "project.html")
	})
	mux.HandleFunc("GET /issues/{id}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, webSub, "issue.html")
	})

	// Middleware stack
	wrapped := internal.PATMiddleware(store)(mux)

	log.Printf("goard listening on %s", cfg.Addr())
	if err := http.ListenAndServe(cfg.Addr(), wrapped); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func loadConfig() internal.Config {
	cfg := internal.Config{
		DBPath: "goard.db",
		Port:   "8300",
	}
	if v := os.Getenv("GOARD_DB_PATH"); v != "" {
		cfg.DBPath = v
	}
	cfg.Host = os.Getenv("GOARD_HOST")
	if v := os.Getenv("GOARD_PORT"); v != "" {
		cfg.Port = v
	}
	cfg.AdminUsername = os.Getenv("GOARD_ADMIN_USERNAME")
	cfg.AdminPAT = os.Getenv("GOARD_ADMIN_PAT")
	return cfg
}
