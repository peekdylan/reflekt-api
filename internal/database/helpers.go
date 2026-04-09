package database

import "database/sql"

// NullString converts a regular Go string into a sql.NullString.
// This is needed because our database columns for mood and ai_analysis
// are nullable — they start empty and get filled in after AI analysis completes.
func NullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}
