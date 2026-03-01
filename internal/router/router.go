package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/howallet/howallet/internal/config"
	"github.com/howallet/howallet/internal/handler"
	mw "github.com/howallet/howallet/internal/middleware"
)

// New creates and configures the chi router with all routes.
func New(
	cfg *config.Config,
	logger *slog.Logger,
	authH *handler.AuthHandler,
	hhH *handler.HouseholdHandler,
	accH *handler.AccountHandler,
	txnH *handler.TransactionHandler,
	expH *handler.ExportHandler,
	checkMembership mw.MembershipChecker,
) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(mw.Logger(logger))
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.Frontend.URL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Household-ID"},
		ExposedHeaders:   []string{"Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		handler.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Public auth routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
		r.Post("/refresh", authH.Refresh)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(mw.JWTAuth(&cfg.JWT))

		// Auth (logout needs JWT)
		r.Post("/auth/logout", authH.Logout)

		// Households (no X-Household-ID needed)
		r.Route("/api/households", func(r chi.Router) {
			r.Post("/", hhH.Create)
			r.Get("/", hhH.List)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/members", hhH.ListMembers)
				r.Get("/invitations", hhH.ListPendingInvitations)
				r.Post("/invite", hhH.Invite)
				r.Delete("/members/{userId}", hhH.RemoveMember)
			})
		})

		// Accept invitation
		r.Post("/api/invitations/{token}/accept", hhH.AcceptInvitation)

		// Routes that require X-Household-ID (membership enforced)
		r.Group(func(r chi.Router) {
			r.Use(mw.HouseholdCtx(checkMembership))

			// Accounts
			r.Route("/api/accounts", func(r chi.Router) {
				r.Post("/", accH.Create)
				r.Get("/", accH.List)
				r.Get("/{id}", accH.Get)
				r.Put("/{id}", accH.Update)
				r.Delete("/{id}", accH.Delete)
			})

			// Transactions
			r.Route("/api/transactions", func(r chi.Router) {
				r.Post("/", txnH.Create)
				r.Get("/", txnH.List)
				r.Get("/{id}", txnH.Get)
				r.Put("/{id}", txnH.Update)
				r.Delete("/{id}", txnH.Delete)
			})

			// Export
			r.Get("/api/export/csv", expH.ExportCSV)
		})
	})

	return r
}
