package entities

import "time"

type Ticket struct {
	ID          uint64       `json:"id"`
	UserID      uint64       `json:"userId"`
	CategoryID  uint32       `json:"categoryId"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       *float32     `json:"price,omitempty"`
	Quantity    uint32       `json:"quantity"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	TagIDs      []uint32     `json:"tagIds,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type CreateTicketDTO struct {
	UserID      uint64   `json:"userId"`
	CategoryID  uint32   `json:"categoryId"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       *float32 `json:"price,omitempty"`
	Quantity    uint32   `json:"quantity"`
	TagIDs      []uint32 `json:"tagIds,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

type Attachment struct {
	ID        uint64    `json:"id"`
	TicketID  uint64    `json:"ticketId"`
	Link      string    `json:"link"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UpdateTicketDTO struct {
	ID                    uint64   `json:"id"`
	CategoryID            *uint32  `json:"categoryId,omitempty"`
	Name                  *string  `json:"name,omitempty"`
	Description           *string  `json:"description,omitempty"`
	Price                 *float32 `json:"price,omitempty"`
	Quantity              *uint32  `json:"quantity,omitempty"`
	TagIDsToAdd           []uint32 `json:"tagIdsToAdd,omitempty"`
	TagIDsToDelete        []uint32 `json:"tagIdsToDelete,omitempty"`
	AttachmentsToAdd      []string `json:"attachmentsToAdd,omitempty"`
	AttachmentIDsToDelete []uint64 `json:"attachmentIdsToDelete,omitempty"`
}

type RawUpdateTicketDTO struct {
	ID          uint64   `json:"id"`
	CategoryID  *uint32  `json:"categoryId,omitempty"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float32 `json:"price,omitempty"`
	Quantity    *uint32  `json:"quantity,omitempty"`
	TagIDs      []uint32 `json:"tagIds,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

type TicketsFilters struct {
	Search              *string  `json:"search,omitempty"`
	PriceCeil           *float32 `json:"priceCeil,omitempty"`     // max price
	PriceFloor          *float32 `json:"priceFloor,omitempty"`    // min price
	QuantityFloor       *uint32  `json:"quantityFloor,omitempty"` // min quantity
	CategoryIDs         []uint32 `json:"categoryIds,omitempty"`
	TagIDs              []uint32 `json:"tagIds,omitempty"`
	CreatedAtOrderByAsc *bool    `json:"createdAtOrderByAsc,omitempty"`
}
