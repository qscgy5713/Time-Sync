package handlers

import (
	"errors"
	"net/http"

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
	_, err = tx.Exec(ctx,
		`INSERT INTO events (id, title, description, timezone) VALUES ($1, $2, $3, $4)`,
		id, req.Title, req.Description, req.Timezone,
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
		ID:  id,
		URL: h.PublicURL + "/event/" + id,
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
		`SELECT id, title, description, timezone, created_at FROM events WHERE id = $1`,
		eventID,
	).Scan(&resp.ID, &resp.Title, &description, &resp.Timezone, &resp.CreatedAt)
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

	var validCount int
	if err := tx.QueryRow(ctx,
		`SELECT count(*) FROM event_options WHERE event_id = $1 AND id = ANY($2)`,
		eventID, req.AvailableOptionIDs,
	).Scan(&validCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if validCount != len(req.AvailableOptionIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "one or more available_option_ids do not belong to this event"})
		return
	}

	var participantID int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO participants (event_id, name) VALUES ($1, $2) RETURNING id`,
		eventID, req.Name,
	).Scan(&participantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create participant"})
		return
	}

	batch := &pgx.Batch{}
	for _, optID := range req.AvailableOptionIDs {
		batch.Queue(
			`INSERT INTO availabilities (participant_id, event_option_id) VALUES ($1, $2)`,
			participantID, optID,
		)
	}
	br := tx.SendBatch(ctx, batch)
	for range req.AvailableOptionIDs {
		if _, err := br.Exec(); err != nil {
			br.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save availabilities"})
			return
		}
	}
	if err := br.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save availabilities"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save participant"})
		return
	}

	c.JSON(http.StatusCreated, models.AddParticipantResponse{ID: participantID})
}
