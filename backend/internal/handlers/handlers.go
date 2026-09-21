package handlers

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"time-sync/backend/internal/models"
)

type Handler struct {
	Pool      *pgxpool.Pool
	PublicURL string
}

func dedupeInt64(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func New(pool *pgxpool.Pool, publicURL string) *Handler {
	return &Handler{Pool: pool, PublicURL: publicURL}
}

// validOptionIDs reports whether every id in optionIDs belongs to eventID.
func validOptionIDs(ctx context.Context, tx pgx.Tx, eventID string, optionIDs []int64) (bool, error) {
	var validCount int
	if err := tx.QueryRow(ctx,
		`SELECT count(*) FROM event_options WHERE event_id = $1 AND id = ANY($2)`,
		eventID, optionIDs,
	).Scan(&validCount); err != nil {
		return false, err
	}
	return validCount == len(optionIDs), nil
}

// eventFinalized reports whether eventID has a finalized_option_id set,
// i.e. voting is closed.
func eventFinalized(ctx context.Context, tx pgx.Tx, eventID string) (bool, error) {
	var finalized bool
	err := tx.QueryRow(ctx,
		`SELECT finalized_option_id IS NOT NULL FROM events WHERE id = $1`,
		eventID,
	).Scan(&finalized)
	if err != nil {
		return false, err
	}
	return finalized, nil
}

// insertAvailabilities writes one row per optionID for participantID.
func insertAvailabilities(ctx context.Context, tx pgx.Tx, participantID int64, optionIDs []int64) error {
	batch := &pgx.Batch{}
	for _, optID := range optionIDs {
		batch.Queue(
			`INSERT INTO availabilities (participant_id, event_option_id) VALUES ($1, $2)`,
			participantID, optID,
		)
	}
	br := tx.SendBatch(ctx, batch)
	for range optionIDs {
		if _, err := br.Exec(); err != nil {
			br.Close()
			return err
		}
	}
	return br.Close()
}

func (h *Handler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	var resp models.StatsResponse
	if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM events`).Scan(&resp.TotalEvents); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM participants`).Scan(&resp.TotalParticipants); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateEvent(c *gin.Context) {
	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, opt := range req.Options {
		if !opt.EndDatetime.After(opt.StartDatetime) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "each option's end_datetime must be after start_datetime"})
			return
		}
	}

	ctx := c.Request.Context()
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer tx.Rollback(ctx)

	id := uuid.New().String()
	ownerToken := uuid.New().String()
	_, err = tx.Exec(ctx,
		`INSERT INTO events (id, title, description, timezone, owner_token) VALUES ($1, $2, $3, $4, $5)`,
		id, req.Title, req.Description, req.Timezone, ownerToken,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	batch := &pgx.Batch{}
	for _, opt := range req.Options {
		batch.Queue(
			`INSERT INTO event_options (event_id, start_datetime, end_datetime) VALUES ($1, $2, $3)`,
			id, opt.StartDatetime, opt.EndDatetime,
		)
	}
	br := tx.SendBatch(ctx, batch)
	for range req.Options {
		if _, err := br.Exec(); err != nil {
			br.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event options"})
			return
		}
	}
	if err := br.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event options"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save event"})
		return
	}

	c.JSON(http.StatusCreated, models.CreateEventResponse{
		ID:         id,
		URL:        h.PublicURL + "/event/" + id,
		OwnerToken: ownerToken,
	})
}

func (h *Handler) GetEvent(c *gin.Context) {
	eventID := c.Param("id")
	if _, err := uuid.Parse(eventID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	ctx := c.Request.Context()

	var resp models.EventDetailResponse
	var description *string
	err := h.Pool.QueryRow(ctx,
		`SELECT id, title, description, timezone, created_at, finalized_option_id FROM events WHERE id = $1`,
		eventID,
	).Scan(&resp.ID, &resp.Title, &description, &resp.Timezone, &resp.CreatedAt, &resp.FinalizedOptionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if description != nil {
		resp.Description = *description
	}

	optRows, err := h.Pool.Query(ctx,
		`SELECT id, start_datetime, end_datetime FROM event_options WHERE event_id = $1 ORDER BY start_datetime`,
		eventID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	resp.Options = []models.EventOption{}
	for optRows.Next() {
		var o models.EventOption
		if err := optRows.Scan(&o.ID, &o.StartDatetime, &o.EndDatetime); err != nil {
			optRows.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		resp.Options = append(resp.Options, o)
	}
	optRows.Close()
	if err := optRows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	partRows, err := h.Pool.Query(ctx,
		`SELECT p.id, p.name, COALESCE(array_agg(a.event_option_id) FILTER (WHERE a.event_option_id IS NOT NULL), '{}')
		 FROM participants p
		 LEFT JOIN availabilities a ON a.participant_id = p.id
		 WHERE p.event_id = $1
		 GROUP BY p.id, p.name
		 ORDER BY p.id`,
		eventID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	resp.Participants = []models.ParticipantSummary{}
	for partRows.Next() {
		var p models.ParticipantSummary
		if err := partRows.Scan(&p.ID, &p.Name, &p.AvailableOptionIDs); err != nil {
			partRows.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		resp.Participants = append(resp.Participants, p)
	}
	partRows.Close()
	if err := partRows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) AddParticipant(c *gin.Context) {
	eventID := c.Param("id")
	if _, err := uuid.Parse(eventID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req models.AddParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.AvailableOptionIDs = dedupeInt64(req.AvailableOptionIDs)

	ctx := c.Request.Context()
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer tx.Rollback(ctx)

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)`, eventID).Scan(&exists); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	finalized, err := eventFinalized(ctx, tx, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if finalized {
		c.JSON(http.StatusConflict, gin.H{"error": "this event's time has been finalized; voting is closed"})
		return
	}

	ok, err := validOptionIDs(ctx, tx, eventID, req.AvailableOptionIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "one or more available_option_ids do not belong to this event"})
		return
	}

	editToken := uuid.New().String()
	var participantID int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO participants (event_id, name, edit_token) VALUES ($1, $2, $3) RETURNING id`,
		eventID, req.Name, editToken,
	).Scan(&participantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create participant"})
		return
	}

	if err := insertAvailabilities(ctx, tx, participantID, req.AvailableOptionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save availabilities"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save participant"})
		return
	}

	c.JSON(http.StatusCreated, models.AddParticipantResponse{ID: participantID, EditToken: editToken})
}

func (h *Handler) UpdateParticipant(c *gin.Context) {
	eventID := c.Param("id")
	if _, err := uuid.Parse(eventID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}
	participantID, err := strconv.ParseInt(c.Param("participantId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid participant id"})
		return
	}

	var req models.UpdateParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.AvailableOptionIDs = dedupeInt64(req.AvailableOptionIDs)

	ctx := c.Request.Context()
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer tx.Rollback(ctx)

	var storedToken string
	err = tx.QueryRow(ctx,
		`SELECT edit_token FROM participants WHERE id = $1 AND event_id = $2`,
		participantID, eventID,
	).Scan(&storedToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "participant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if storedToken != req.EditToken {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid edit token"})
		return
	}

	finalized, err := eventFinalized(ctx, tx, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if finalized {
		c.JSON(http.StatusConflict, gin.H{"error": "this event's time has been finalized; voting is closed"})
		return
	}

	ok, err := validOptionIDs(ctx, tx, eventID, req.AvailableOptionIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "one or more available_option_ids do not belong to this event"})
		return
	}

	if _, err := tx.Exec(ctx, `UPDATE participants SET name = $1 WHERE id = $2`, req.Name, participantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update participant"})
		return
	}
	if _, err := tx.Exec(ctx, `DELETE FROM availabilities WHERE participant_id = $1`, participantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update availabilities"})
		return
	}
	if err := insertAvailabilities(ctx, tx, participantID, req.AvailableOptionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save availabilities"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save participant"})
		return
	}

	c.JSON(http.StatusOK, models.AddParticipantResponse{ID: participantID})
}

// FinalizeEvent sets (or, with a null option_id, clears) the event's chosen
// time. Only the organizer may do this: the request must carry the
// owner_token issued once at event creation.
func (h *Handler) FinalizeEvent(c *gin.Context) {
	eventID := c.Param("id")
	if _, err := uuid.Parse(eventID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req models.FinalizeEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer tx.Rollback(ctx)

	var storedOwnerToken *string
	err = tx.QueryRow(ctx, `SELECT owner_token::text FROM events WHERE id = $1`, eventID).Scan(&storedOwnerToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if storedOwnerToken == nil || subtle.ConstantTimeCompare([]byte(*storedOwnerToken), []byte(req.OwnerToken)) != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the event organizer can finalize"})
		return
	}

	if req.OptionID != nil {
		ok, err := validOptionIDs(ctx, tx, eventID, []int64{*req.OptionID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "option_id does not belong to this event"})
			return
		}
	}

	tag, err := tx.Exec(ctx, `UPDATE events SET finalized_option_id = $1 WHERE id = $2`, req.OptionID, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update event"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save event"})
		return
	}

	c.Status(http.StatusNoContent)
}
