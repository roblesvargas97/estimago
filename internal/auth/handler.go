package auth

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/roblesvargas97/estimago/internal/utils"
)

// RegisterRoutes wires the auth HTTP handlers into the provided router.
// Example usage:
//
//	r.Route("/api/v1/auth", func(sub chi.Router) {
//	        auth.RegisterRoutes(sub, pool, cfg)
//	})
func RegisterRoutes(r chi.Router, pool *pgxpool.Pool, cfg Config) {
	r.Post("/signup", SignupHandler(pool))
	r.Post("/login", LoginHandler(pool, cfg))
}

// SignupHandler handles user registration.
func SignupHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in SignupIn
		if err := utils.DecodeJSON(w, r, &in); err != nil {
			utils.WriteErr(w, http.StatusBadRequest, "bad_json", err.Error())
			return
		}

		in.Name = strings.TrimSpace(in.Name)
		in.Email = strings.ToLower(strings.TrimSpace(in.Email))
		in.Password = strings.TrimSpace(in.Password)

		if in.Email == "" || in.Password == "" {
			utils.WriteErr(w, http.StatusUnprocessableEntity, "validation_error", "email and password are required")
			return
		}

		hashed, err := HashPassword(in.Password)
		if err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, "hash_error", err.Error())
			return
		}

		user, err := InsertUser(r.Context(), pool, in.Name, in.Email, hashed)
		if err != nil {
			if utils.IsUniqueViolationErr(err) {
				utils.WriteErr(w, http.StatusConflict, "conflict", "email already registered")
				return
			}
			utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}

		out := SignupOut{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			PlanID:    user.PlanID,
			CreatedAt: user.CreatedAt,
		}

		utils.WriteJSON(w, http.StatusCreated, out)
	}
}

// LoginHandler handles user login and JWT issuance.
func LoginHandler(pool *pgxpool.Pool, cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in LoginIn
		if err := utils.DecodeJSON(w, r, &in); err != nil {
			utils.WriteErr(w, http.StatusBadRequest, "bad_json", err.Error())
			return
		}

		in.Email = strings.ToLower(strings.TrimSpace(in.Email))
		in.Password = strings.TrimSpace(in.Password)

		if in.Email == "" || in.Password == "" {
			utils.WriteErr(w, http.StatusUnprocessableEntity, "validation_error", "email and password are required")
			return
		}

		user, err := GetUserByEmail(r.Context(), pool, in.Email)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErr(w, http.StatusUnauthorized, "invalid_credentials", "email or password incorrect")
				return
			}
			utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}

		if err := CheckPassword(user.PasswordHash, in.Password); err != nil {
			utils.WriteErr(w, http.StatusUnauthorized, "invalid_credentials", "email or password incorrect")
			return
		}

		token, err := MakeJWT(cfg, user)
		if err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, "token_error", err.Error())
			return
		}

		utils.WriteJSON(w, http.StatusOK, LoginOut{
			Token: token,
			User: AuthUser{
				ID:     user.ID,
				Email:  user.Email,
				PlanID: user.PlanID,
			},
		})
	}
}

// MeHandler returns the authenticated user profile.
func MeHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromCtx(r)
		if !ok {
			utils.WriteErr(w, http.StatusUnauthorized, "unauthorized", "invalid token")
			return
		}

		user, err := GetUserByID(r.Context(), pool, userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErr(w, http.StatusNotFound, "not_found", "user not found")
				return
			}
			utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}

		utils.WriteJSON(w, http.StatusOK, SignupOut{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			PlanID:    user.PlanID,
			CreatedAt: user.CreatedAt,
		})
	}
}
