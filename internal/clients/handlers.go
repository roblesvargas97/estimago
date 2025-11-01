package clients

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/roblesvargas97/estimago/internal/utils"
)

func PostClient(pool *pgxpool.Pool) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var in CreateClientIn
		if err := utils.DecodeJSON(w, r, &in); err != nil {

			utils.WriteErr(w, http.StatusBadRequest, "bad_json", err.Error())
			return

		}

		in.Name = strings.TrimSpace(in.Name)
		if in.Name == "" {
			utils.WriteErr(w, http.StatusUnprocessableEntity, "validation_error", "name is required")
			return
		}

		var n int

		if err := pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM clients`).Scan(&n); err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}

		if n >= 3 {
			utils.WriteErr(w, http.StatusForbidden, "limit_reached", "free tier client limit reached")
			return
		}

		var c Client

		meta := []byte(`{}`)

		if in.Meta != nil {
			meta = *in.Meta
		}

		err := pool.QueryRow(r.Context(), `
		
		INSERT INTO clients (name, email, phone, meta)
		VALUES ($1, $2, $3, $4)
		RETURNING id,name,email,phone,meta,created_at,updated_at
		`, in.Name, in.Email, in.Phone, meta).Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Meta, &c.CreatedAt, &c.UpdatedAt)

		if err != nil && utils.IsUniqueViolationErr(err) {
			utils.WriteErr(w, http.StatusConflict, "conflict", "client with same (name,email) already exists")
			return
		}

		if err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}
		utils.WriteJSON(w, http.StatusCreated, c)

	}

}

func ListClients(pool *pgxpool.Pool) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		q := strings.TrimSpace(r.URL.Query().Get("q"))

		page, _ := strconv.Atoi(utils.DefaultIfEmpty(r.URL.Query().Get("page"), "1"))

		limit, _ := strconv.Atoi(utils.DefaultIfEmpty(r.URL.Query().Get("limit"), "20"))

		if limit <= 0 || limit > 100 {
			limit = 20
		}

		offset := (page - 1) * limit

		var total int

		if q == "" {

			if err := pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM clients`).Scan(&total); err != nil {
				utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
				return
			} else {
				if err := pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM clients WHERE name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%'`, q).Scan(&total); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
					return
				}
			}

		}

		sql := `SELECT id, name, email, phone, meta, created_at, updated_at
		FROM clients
		`
		args := []any{}

		if q != "" {
			sql += ` WHERE name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%' `
			args = append(args, q)
		}

		sql += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
		args = append(args, limit, offset)

		rows, err := pool.Query(r.Context(), sql, args...)

		if err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}
		defer rows.Close()

		outs := []Client{}

		for rows.Next() {
			var c Client
			if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Meta, &c.CreatedAt, &c.UpdatedAt); err != nil {
				utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
				return
			}
			outs = append(outs, c)
		}
		w.Header().Set("X-Total-Count", strconv.Itoa(total))
		utils.WriteJSON(w, http.StatusOK, outs)

	}

}

func GetClient(pool *pgxpool.Pool) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		idStr := strings.TrimSpace(chi.URLParam(r, "id"))

		id, err := uuid.Parse(idStr)
		if err != nil {
			utils.WriteErr(w, http.StatusBadRequest, "validation_error", "invalid uuid")
			return
		}

		var c Client

		err = pool.QueryRow(r.Context(), `SELECT id,name,email,phone,meta,created_at,updated_at FROM clients WHERE id=$1`, id).Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Meta, &c.CreatedAt, &c.UpdatedAt)

		if errors.Is(err, context.DeadlineExceeded) || err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, "not_found", "client not found")
			return
		}

		if err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}

		utils.WriteJSON(w, http.StatusOK, c)

	}

}
