package responds

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DKhorkov/hmtm-tickets/api/protobuf/generated/go/tickets"
	"github.com/DKhorkov/hmtm-tickets/internal/entities"
)

func mapRespondOut(respond entities.Respond) *tickets.GetRespondOut {
	return &tickets.GetRespondOut{
		ID:        respond.ID,
		TicketID:  respond.TicketID,
		MasterID:  respond.MasterID,
		Price:     respond.Price,
		Comment:   respond.Comment,
		CreatedAt: timestamppb.New(respond.CreatedAt),
		UpdatedAt: timestamppb.New(respond.UpdatedAt),
	}
}
