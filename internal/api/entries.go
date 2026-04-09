package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	// Internal packages
	"github.com/peekdylan/reflekt-api/internal/ai"
	"github.com/peekdylan/reflekt-api/internal/database"

	// Third-party packages
	"github.com/google/uuid"
)

// createEntryRequest defines the expected JSON body for creating a new entry.
type createEntryRequest struct {
	Title string   `json:"title"`
	Body  string   `json:"body"`
	Tags  []string `json:"tags"`
}

// entryResponse is the structure we return to the client for a journal entry.
// We format fields like dates and UUIDs as clean strings for the frontend.
type entryResponse struct {
	ID         string   `json:"id"`
	UserID     string   `json:"user_id"`
	Title      string   `json:"title"`
	Body       string   `json:"body"`
	Mood       string   `json:"mood"`
	AIAnalysis string   `json:"ai_analysis"`
	Tags       []string `json:"tags"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

// formatEntry converts a database.Entry into an entryResponse for the API.
// Keeping this conversion in one place makes it easy to change later.
func formatEntry(entry database.Entry) entryResponse {
	// Handle nullable fields safely — these are empty until AI analysis completes
	mood := ""
	if entry.Mood.Valid {
		mood = entry.Mood.String
	}

	aiAnalysis := ""
	if entry.AiAnalysis.Valid {
		aiAnalysis = entry.AiAnalysis.String
	}

	return entryResponse{
		ID:         entry.ID.String(),
		UserID:     entry.UserID.String(),
		Title:      entry.Title,
		Body:       entry.Body,
		Mood:       mood,
		AIAnalysis: aiAnalysis,
		Tags:       entry.Tags,
		CreatedAt:  entry.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  entry.UpdatedAt.Format(time.RFC3339),
	}
}

// HandlerCreateEntry creates a new journal entry for the authenticated user.
// After saving the entry, it triggers AI analysis in the background so the
// user gets an instant response without waiting for the AI to finish.
func (cfg *APIConfig) HandlerCreateEntry(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user's ID from the request context (set by middleware)
	userID, ok := getUserIDFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Decode the request body
	var req createEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Title == "" || req.Body == "" {
		respondWithError(w, http.StatusBadRequest, "Title and body are required")
		return
	}

	// Default to empty tags slice if none provided
	if req.Tags == nil {
		req.Tags = []string{}
	}

	// Parse the user ID string into a UUID
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	// Insert the entry into the database
	entry, err := cfg.DB.CreateEntry(r.Context(), database.CreateEntryParams{
		UserID: parsedUserID,
		Title:  req.Title,
		Body:   req.Body,
		Tags:   req.Tags,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create entry")
		return
	}

	// Respond immediately so the user isn't waiting for AI analysis
	respondWithJSON(w, http.StatusCreated, formatEntry(entry))

	// Analyze the entry with Claude AI in the background using a goroutine.
	// The mood and analysis will be saved to the database once Claude responds.
	// We use context.Background() because the request context will be cancelled
	// after we respond, but we still want the goroutine to finish its work.
	go func() {
		result, err := ai.AnalyzeEntry(cfg.AnthropicKey, entry.Title, entry.Body)
		if err != nil {
			log.Printf("AI analysis failed for entry %s: %v", entry.ID, err)
			return
		}

		log.Printf("AI analysis complete for entry %s — mood: %s", entry.ID, result.Mood)

		// Update the entry with the AI-generated mood and analysis
		_, err = cfg.DB.UpdateEntryAIAnalysis(context.Background(), database.UpdateEntryAIAnalysisParams{
			ID:         entry.ID,
			Mood:       database.NullString(result.Mood),
			AiAnalysis: database.NullString(result.Analysis),
		})
		if err != nil {
			log.Printf("Failed to save AI analysis for entry %s: %v", entry.ID, err)
		}
	}()
}

// HandlerGetEntries returns all journal entries for the authenticated user,
// ordered by most recent first.
func (cfg *APIConfig) HandlerGetEntries(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	// Fetch all entries for this user from the database
	entries, err := cfg.DB.GetEntriesByUserID(r.Context(), parsedUserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch entries")
		return
	}

	// Convert each database entry into an API response struct
	response := make([]entryResponse, len(entries))
	for i, entry := range entries {
		response[i] = formatEntry(entry)
	}

	respondWithJSON(w, http.StatusOK, response)
}

// HandlerDeleteEntry deletes a specific journal entry by ID.
// Users can only delete their own entries — the query filters by both entry ID and user ID.
func (cfg *APIConfig) HandlerDeleteEntry(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Extract the entry ID from the URL path (e.g. /v1/entries/some-uuid)
	entryIDStr := r.PathValue("id")
	if entryIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "Entry ID is required")
		return
	}

	parsedEntryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid entry ID")
		return
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	// Delete the entry — the query ensures users can only delete their own entries
	if err := cfg.DB.DeleteEntry(r.Context(), database.DeleteEntryParams{
		ID:     parsedEntryID,
		UserID: parsedUserID,
	}); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete entry")
		return
	}

	// 204 No Content is the standard response for a successful delete
	w.WriteHeader(http.StatusNoContent)
}
