package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/runtimeninja/blushbew/services/api/internal/auth"
	"github.com/runtimeninja/blushbew/services/api/internal/blog"
	"github.com/runtimeninja/blushbew/services/api/internal/diagnosis"
)

type Deps struct {
	Logger         *slog.Logger
	Pool           *pgxpool.Pool
	AllowedOrigins []string
	Auth           *auth.Service
	Blog           *blog.Repo
}

func NewRouter(d Deps) *chi.Mux {
	r := chi.NewRouter()

	// middleware chain
	r.Use(func(next http.Handler) http.Handler { return RequestID(next) })
	r.Use(func(next http.Handler) http.Handler { return Recover(d.Logger, next) })
	r.Use(func(next http.Handler) http.Handler { return CORS(d.AllowedOrigins, next) })
	r.Use(func(next http.Handler) http.Handler { return AccessLog(d.Logger, next) })

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	r.Route("/v1", func(v1 chi.Router) {
		// Public blog
		v1.Get("/blog", func(w http.ResponseWriter, r *http.Request) {
			limit := 20
			if s := r.URL.Query().Get("limit"); s != "" {
				if n, err := strconv.Atoi(s); err == nil {
					limit = n
				}
			}
			posts, err := d.Blog.ListPublished(r.Context(), limit)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "blog_list_failed"})
				return
			}
			// We don't want to send full content in list; keep excerpt
			type item struct {
				Slug        string     `json:"slug"`
				Title       string     `json:"title"`
				Excerpt     string     `json:"excerpt"`
				Category    string     `json:"category"`
				PublishedAt *time.Time `json:"published_at,omitempty"`
			}
			out := []item{}
			for _, p := range posts {
				out = append(out, item{
					Slug: p.Slug, Title: p.Title, Excerpt: p.Excerpt, Category: p.Category, PublishedAt: p.PublishedAt,
				})
			}
			writeJSON(w, http.StatusOK, map[string]any{"posts": out})
		})

		v1.Get("/blog/{slug}", func(w http.ResponseWriter, r *http.Request) {
			slug := chi.URLParam(r, "slug")
			p, err := d.Blog.GetPublishedBySlug(r.Context(), slug)
			if err != nil {
				writeJSON(w, http.StatusNotFound, map[string]any{"error": "not_found"})
				return
			}
			writeJSON(w, http.StatusOK, p)
		})

		// Tool: diagnosis
		v1.Post("/diagnosis", func(w http.ResponseWriter, r *http.Request) {
			var in diagnosis.Input
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
				return
			}
			if len(in.ProblemText) < 5 {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "problem_text_too_short"})
				return
			}

			out := diagnosis.Fallback(in)

			// Store request for analytics + future tuning
			b, _ := json.Marshal(out)
			_, _ = d.Pool.Exec(r.Context(), `
				insert into diagnosis_requests(problem_text, skin_type, climate, event_type, result_json)
				values($1,$2,$3,$4,$5)
			`, in.ProblemText, in.SkinType, in.Climate, in.EventType, b)

			writeJSON(w, http.StatusOK, out)
		})

		// Admin auth
		v1.Post("/admin/login", func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Email    string `json:"email"`
				Password string `json:"password"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
				return
			}
			token, err := d.Auth.Login(r.Context(), body.Email, body.Password)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid_credentials"})
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "bb_admin_session",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				// Secure: true (prod এ enable করবো)
			})
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		})

		// Admin protected routes
		v1.Route("/admin", func(ad chi.Router) {
			ad.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					c, err := r.Cookie("bb_admin_session")
					if err != nil || c.Value == "" {
						writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
						return
					}
					ok, err := d.Auth.ValidateSession(r.Context(), c.Value)
					if err != nil || !ok {
						writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
						return
					}
					next.ServeHTTP(w, r)
				})
			})

			ad.Get("/posts", func(w http.ResponseWriter, r *http.Request) {
				posts, err := d.Blog.AdminList(r.Context(), 50)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "admin_list_failed"})
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"posts": posts})
			})

			ad.Post("/posts", func(w http.ResponseWriter, r *http.Request) {
				var p blog.Post
				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
					return
				}
				if p.Slug == "" || p.Title == "" || p.ContentMD == "" {
					writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing_fields"})
					return
				}
				if p.Status == "" {
					p.Status = "draft"
				}
				id, err := d.Blog.AdminCreate(r.Context(), p)
				if err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]any{"error": "create_failed"})
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"id": id})
			})

			ad.Post("/posts/{id}/publish", func(w http.ResponseWriter, r *http.Request) {
				idStr := chi.URLParam(r, "id")
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_id"})
					return
				}
				if err := d.Blog.AdminPublish(r.Context(), id); err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "publish_failed"})
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			})
		})
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
