package tickets

import (
	"github.com/DKhorkov/hmtm-tickets/api/protobuf/generated/go/tickets"
	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func prepareTicketOut(ticket *entities.Ticket) *tickets.GetTicketOut {
	attachments := make([]*tickets.Attachment, len(ticket.Attachments))
	for j, attachment := range ticket.Attachments {
		attachments[j] = &tickets.Attachment{
			ID:        attachment.ID,
			TicketID:  attachment.TicketID,
			Link:      attachment.Link,
			CreatedAt: timestamppb.New(attachment.CreatedAt),
			UpdatedAt: timestamppb.New(attachment.UpdatedAt),
		}
	}

	return &tickets.GetTicketOut{
		ID:          ticket.ID,
		UserID:      ticket.UserID,
		Name:        ticket.Name,
		Description: ticket.Description,
		Price:       ticket.Price,
		Quantity:    ticket.Quantity,
		CategoryID:  ticket.CategoryID,
		TagIDs:      ticket.TagIDs,
		Attachments: attachments,
		CreatedAt:   timestamppb.New(ticket.CreatedAt),
		UpdatedAt:   timestamppb.New(ticket.UpdatedAt),
	}
}
