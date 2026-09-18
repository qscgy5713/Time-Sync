package models

import "time"

type EventOptionInput struct {
	StartDatetime time.Time `json:"start_datetime" binding:"required"`
	EndDatetime   time.Time `json:"end_datetime" binding:"required"`
}

type CreateEventRequest struct {
	Title       string              `json:"title" binding:"required,max=100"`
	Description string              `json:"description"`
	Timezone    string              `json:"timezone" binding:"required"`
	Options     []EventOptionInput  `json:"options" binding:"required,min=1,max=2000,dive"`
}

type CreateEventResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type EventOption struct {
	ID            int64     `json:"id"`
	StartDatetime time.Time `json:"start_datetime"`
	EndDatetime   time.Time `json:"end_datetime"`
}

type ParticipantSummary struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	AvailableOptionIDs []int64 `json:"available_option_ids"`
}

type EventDetailResponse struct {
	ID           string               `json:"id"`
	Title        string               `json:"title"`
	Description  string               `json:"description"`
	Timezone     string               `json:"timezone"`
	CreatedAt    time.Time            `json:"created_at"`
	Options      []EventOption        `json:"options"`
	Participants []ParticipantSummary `json:"participants"`
}

type AddParticipantRequest struct {
	Name               string  `json:"name" binding:"required,max=50"`
	AvailableOptionIDs []int64 `json:"available_option_ids" binding:"required,min=1,max=2000"`
}

type AddParticipantResponse struct {
	ID        int64  `json:"id"`
	EditToken string `json:"edit_token"`
}

type UpdateParticipantRequest struct {
	Name               string  `json:"name" binding:"required,max=50"`
	AvailableOptionIDs []int64 `json:"available_option_ids" binding:"required,min=1,max=2000"`
	EditToken          string  `json:"edit_token" binding:"required,uuid"`
}
