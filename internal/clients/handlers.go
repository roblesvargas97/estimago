package clients

import (
	"net/http"
	"strings"

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
