package convert

import "github.com/jackc/pgx/v5/pgtype"

// FormatNullableDate formats a nullable pgtype.Date scanned from a date column using
// layout, returning nil if the column was NULL.
func FormatNullableDate(d pgtype.Date, layout string) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.Format(layout)
	return &s
}
