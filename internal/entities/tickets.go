package entities

import "time"

type Ticket struct {
	ID          uint64       `json:"id"`
	UserID      uint64       `json:"user_id"`
	CategoryID  uint32       `json:"category_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       *float32     `json:"price,omitempty"`
	Quantity    uint32       `json:"quantity"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	TagIDs      []uint32     `json:"tag_ids,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type CreateTicketDTO struct {
	UserID      uint64   `json:"user_id"`
	CategoryID  uint32   `json:"category_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       *float32 `json:"price,omitempty"`
	Quantity    uint32   `json:"quantity"`
	TagIDs      []uint32 `json:"tag_ids,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

type Attachment struct {
	ID        uint64    `json:"id"`
	TicketID  uint64    `json:"ticket_id"`
	Link      string    `json:"link"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateTicketDTO struct {
	ID                    uint64   `json:"id"`
	CategoryID            *uint32  `json:"category_id,omitempty"`
	Name                  *string  `json:"name,omitempty"`
	Description           *string  `json:"description,omitempty"`
	Price                 *float32 `json:"price,omitempty"`
	Quantity              *uint32  `json:"quantity,omitempty"`
	TagIDsToAdd           []uint32 `json:"tag_ids_to_add,omitempty"`
	TagIDsToDelete        []uint32 `json:"tag_ids_to_delete,omitempty"`
	AttachmentsToAdd      []string `json:"attachments_to_add,omitempty"`
	AttachmentIDsToDelete []uint64 `json:"attachment_ids_to_delete,omitempty"`
}

type RawUpdateTicketDTO struct {
	ID          uint64   `json:"id"`
	CategoryID  *uint32  `json:"category_id,omitempty"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float32 `json:"price,omitempty"`
	Quantity    *uint32  `json:"quantity,omitempty"`
	TagIDs      []uint32 `json:"tag_ids,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}
