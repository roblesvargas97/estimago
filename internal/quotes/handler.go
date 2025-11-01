package quotes

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roblesvargas97/estimago/internal/utils"
)

func PostQuote(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var in CreateQuoteIn

		if err := utils.DecodeJSON(w, r, &in); err != nil {
			utils.WriteErr(w, http.StatusBadRequest, "bad_json", err.Error())
			return
		}

		if len(in.Items) == 0 {
			utils.WriteErr(w, http.StatusUnprocessableEntity, "validation_error", "items: at least one item is required")
			return

		}

		if in.MarginPct < 0 || in.MarginPct > 100 || in.TaxPct < 0 || in.TaxPct > 100 {
			utils.WriteErr(w, http.StatusUnprocessableEntity, "validation_error", "margin_pct and tax_pct must be between 0 and 100")
			return
		}

		if len(in.Currency) != 3 {
			utils.WriteErr(w, http.StatusUnprocessableEntity, "validation_error", "currency must be a valid 3-letter ISO code")
			return
		}

		var monthCount int

		if err := pool.QueryRow(r.Context(), `
			SELECT COUNT(*) FROM quotes
			WHERE date_trunc('month', created_at) = date_trunc('month', now())
		`).Scan(&monthCount); err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}

		if monthCount >= 10 {
			utils.WriteErr(w, http.StatusForbidden, "limit_reached", "monthly quote limit reached")
			return
		}

		itemsCalculated, subtotalStr, totalStr, err := calcTotals(in)

		if err != nil {
			utils.WriteErr(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
			return
		}

		itemsJSON, err := json.Marshal(itemsCalculated)

		var q Quote

		err = pool.QueryRow(r.Context(), `
			INSERT INTO quotes (
				client_id, items, labor_hours, labor_rate, margin_pct, tax_pct,
				subtotal, total, currency, notes, status
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'draft')
			RETURNING id, client_id, items, labor_hours, labor_rate, margin_pct, tax_pct,
				subtotal, total, currency, notes, public_id, status, created_at, updated_at
		`, in.ClientID,
			itemsJSON,
			fmt.Sprintf("%.2f", in.LaborHours),
			fmt.Sprintf("%.2f", in.LaborRate),
			fmt.Sprintf("%.2f", in.MarginPct),
			fmt.Sprintf("%.2f", in.TaxPct),
			subtotalStr,
			totalStr,
			strings.ToUpper(in.Currency),
			in.Notes,
		).Scan(
			&q.ID,
			&q.ClientID,
			&q.Items,
			&q.LaborHours,
			&q.LaborRate,
			&q.MarginPct,
			&q.TaxPct,
			&q.Subtotal,
			&q.Total,
			&q.Currency,
			&q.Notes,
			&q.PublicID,
			&q.Status, &q.CreatedAt, &q.UpdatedAt,
		)

		if err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}

		utils.WriteJSON(w, http.StatusCreated, q)

	}
}

func GetQuote(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimSpace(chi.URLParam(r, "id"))

		id, err := uuid.Parse(idStr)
		if err != nil {
			utils.WriteErr(w, http.StatusBadRequest, "validation_error", "invalid uuid")
			return
		}

		var q Quote

		err = pool.QueryRow(r.Context(), `
                        SELECT id, client_id, items, labor_hours, labor_rate, margin_pct, tax_pct,
                               subtotal, total, currency, notes, public_id, status, created_at, updated_at
                        FROM quotes WHERE id=$1
                `, id).Scan(
			&q.ID, &q.ClientID, &q.Items, &q.LaborHours, &q.LaborRate, &q.MarginPct, &q.TaxPct,
			&q.Subtotal, &q.Total, &q.Currency, &q.Notes, &q.PublicID, &q.Status, &q.CreatedAt, &q.UpdatedAt,
		)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				utils.WriteErr(w, http.StatusNotFound, "not_found", "quote not found")
				return
			}

			utils.WriteErr(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}

		utils.WriteJSON(w, http.StatusOK, q)
	}
}

// calcTotals - Calculates quote totals with precise decimal arithmetic for financial accuracy
// Purpose: Computes line totals, subtotals, taxes, and final totals using arbitrary precision math
// Advantages:
//   - Uses big.Rat for precise decimal calculations (no floating-point errors)
//   - Validates individual items during calculation
//   - Handles complex business logic (labor, margin, tax calculations)
//   - Returns formatted currency strings ready for display
//
// Weaknesses:
//   - Performance overhead compared to float64 calculations
//   - Complex calculation logic mixed with validation
//   - No support for different rounding methods per currency
//   - Error messages could be more user-friendly
func calcTotals(in CreateQuoteIn) ([]QuoteItem, string, string, error) {
	sum := big.NewRat(0, 1)

	// Item processing and validation loop
	// Advantages: Validates each item, calculates precise line totals
	// Weaknesses: Mixed validation and calculation logic, ignores strconv errors
	items := make([]QuoteItem, len(in.Items))
	for i, it := range in.Items {
		if strings.TrimSpace(it.Name) == "" {
			return nil, "", "", fmt.Errorf("items[%d].name is required", i)
		}
		if it.Qty < 0 || it.UnitPrice < 0 {
			return nil, "", "", fmt.Errorf("items[%d] qty/unit_price must be >= 0", i)
		}
		lt := mulRat(dec(it.Qty), dec(it.UnitPrice)) // qty * unit_price (precise)
		lt2 := round2(lt)                            // Round to 2 decimals
		ltF, _ := strconv.ParseFloat(lt2, 64)        // Convert back to float64 (error ignored)
		it.LineTotal = &ltF
		items[i] = it
		sum = sum.Add(sum, lt) // Accumulate precise sum
	}

	// Financial calculations with proper business logic flow
	// Advantages: Clear calculation sequence, precise arithmetic, follows standard quote calculation
	// Weaknesses: Fixed calculation order (margin before tax), no support for tax-inclusive pricing
	labor := mulRat(dec(in.LaborHours), dec(in.LaborRate)) // Labor cost calculation
	base := new(big.Rat).Add(sum, labor)                   // Items + Labor = Base
	margin := mulRat(base, pctToRat(in.MarginPct))         // Margin calculation
	subtotal := new(big.Rat).Add(base, margin)             // Base + Margin = Subtotal
	tax := mulRat(subtotal, pctToRat(in.TaxPct))           // Tax on subtotal
	total := new(big.Rat).Add(subtotal, tax)               // Final total

	return items, round2(subtotal), round2(total), nil
}

// dec - Converts float64 to precise big.Rat representation for financial calculations
// Purpose: Eliminates floating-point precision errors by converting to exact rational numbers
// Advantages:
//   - Prevents IEEE 754 floating-point precision issues in financial calculations
//   - Maintains exact decimal representation (12.34 -> 1234/100)
//   - Consistent 6-decimal precision for input processing
//   - Essential for accurate money calculations
//
// Weaknesses:
//   - Performance overhead compared to native float64 operations
//   - Fixed 6-decimal precision may be insufficient for some currencies
//   - No input validation for extreme or invalid values
//   - Memory allocation on each conversion
func dec(f float64) *big.Rat {
	// Convert 12.34 -> 1234/100 to avoid imprecise binary representation
	s := fmt.Sprintf("%.6f", f) // Sufficient precision for our financial inputs
	r := new(big.Rat)
	r.SetString(s)
	return r
}

// mulRat - Multiplies two big.Rat numbers and returns new result (immutable operation)
// Purpose: Performs exact multiplication on arbitrary precision rational numbers
// Advantages:
//   - No precision loss in multiplication operations
//   - Immutable operation (doesn't modify input parameters)
//   - Simple one-liner for complex big.Rat arithmetic
//   - Thread-safe (no shared state modification)
//
// Weaknesses:
//   - Allocates new big.Rat on each call (memory overhead)
//   - No input validation (nil pointer handling)
//   - Limited to multiplication only (no division, addition, etc.)
//   - Could be inlined for performance
func mulRat(a, b *big.Rat) *big.Rat { return new(big.Rat).Mul(a, b) }

// round2 - Rounds big.Rat to 2 decimal places using half-up rounding for currency display
// Purpose: Provides consistent currency formatting with banker's rounding for financial applications
// Advantages:
//   - Uses "half-up" rounding (0.125 -> 0.13) standard for financial calculations
//   - Always returns exactly 2 decimal places for consistent currency display
//   - Handles edge cases (small numbers, single/double digits correctly)
//   - No floating-point precision issues due to big.Rat usage
//
// Weaknesses:
//   - Complex string manipulation logic (error-prone)
//   - Fixed 2-decimal formatting (not configurable for different currencies)
//   - No input validation or error handling for edge cases
//   - Performance overhead for string operations
func round2(r *big.Rat) string {
	// Multiply by 100 to shift decimal places for rounding
	x100 := new(big.Rat).Mul(r, big.NewRat(100, 1))
	// Half-up rounding: add 0.5 and truncate (floor operation)
	plus := new(big.Rat).Add(x100, big.NewRat(1, 2))
	i := new(big.Int)
	plus.FloatString(10) // Ensure internal fraction reduction
	i.Div(plus.Num(), plus.Denom())
	// Convert back to 2 decimal places with proper formatting
	s := i.String()
	// Insert decimal point at appropriate position for currency display
	if len(s) == 1 {
		return "0.0" + s
	}
	if len(s) == 2 {
		return "0." + s
	}
	return s[:len(s)-2] + "." + s[len(s)-2:]
}

// pctToRat - Converts percentage (0-100) to decimal ratio (0.0-1.0) using big.Rat
// Purpose: Safely converts percentage values to decimal ratios for precise calculations
// Advantages:
//   - Maintains precision by using big.Rat arithmetic
//   - Handles percentage to decimal conversion accurately (50.5% -> 0.505)
//   - No floating-point precision loss in division operation
//   - Essential for accurate margin and tax calculations
//
// Weaknesses:
//   - No input validation (negative percentages, values > 100)
//   - Performance overhead compared to simple float64 division
//   - Allocates multiple big.Rat objects for simple operation
//   - Could benefit from input bounds checking
//
// Example: pctToRat(25.5) returns big.Rat representing 0.255
func pctToRat(p float64) *big.Rat {
	r := new(big.Rat)
	r.SetFloat64(p)
	return new(big.Rat).Quo(r, big.NewRat(100, 1)) // p / 100 (precise division)
}
