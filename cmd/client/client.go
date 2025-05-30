package main

import (
	"context"
	"fmt"

	"github.com/DKhorkov/libs/pointers"
	"github.com/DKhorkov/libs/requestid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/DKhorkov/hmtm-tickets/api/protobuf/generated/go/tickets"
)

type Client struct {
	tickets.TicketsServiceClient
	tickets.RespondsServiceClient
}

func main() {
	clientConnection, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", "0.0.0.0", 8050),
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	if err != nil {
		panic(err)
	}

	client := &Client{
		TicketsServiceClient:  tickets.NewTicketsServiceClient(clientConnection),
		RespondsServiceClient: tickets.NewRespondsServiceClient(clientConnection),
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), requestid.Key, requestid.New())

	ticketID, err := client.CreateTicket(ctx, &tickets.CreateTicketIn{
		UserID:      31,
		CategoryID:  1,
		TagIDs:      []uint32{1},
		Name:        "test ticket 2",
		Description: "test description",
		Quantity:    10,
		Attachments: []string{"someref", "anotherref"},
	})
	fmt.Println("ticketID:", ticketID, "err:", err)

	ticket, err := client.GetTicket(ctx, &tickets.GetTicketIn{
		ID: 2,
	})
	fmt.Println("ticket by ID:", ticket, "err:", err)

	allTickets, err := client.GetTickets(ctx, &tickets.GetTicketsIn{})
	fmt.Println("allTickets:", allTickets, "err:", err)

	userTickets, err := client.GetUserTickets(ctx, &tickets.GetUserTicketsIn{
		UserID: 31,
	},
	)
	fmt.Println("userTickets:", userTickets, "err:", err)

	respondsID, err := client.RespondToTicket(ctx, &tickets.RespondToTicketIn{
		TicketID: 2,
		UserID:   1,
		Price:    112,
		Comment:  pointers.New[string]("test"),
	})
	fmt.Println("respondsID:", respondsID, "err:", err)

	respond, err := client.GetRespond(ctx, &tickets.GetRespondIn{
		ID: respondsID.GetRespondID(),
	})
	fmt.Println("respond:", respond, "err:", err)

	userResponds, err := client.GetUserResponds(ctx, &tickets.GetUserRespondsIn{
		UserID: 1,
	})
	fmt.Println("userResponds:", userResponds, "err:", err)

	ticketResponds, err := client.GetTicketResponds(ctx, &tickets.GetTicketRespondsIn{
		TicketID: 2,
	})
	fmt.Println("ticketResponds:", ticketResponds, "err:", err)

	_, err = client.UpdateRespond(ctx, &tickets.UpdateRespondIn{
		ID:      respond.GetID(),
		Price:   pointers.New[float32](112.50),
		Comment: pointers.New[string]("test228"),
	})
	fmt.Println("err:", err)

	_, err = client.DeleteRespond(ctx, &tickets.DeleteRespondIn{ID: respond.GetID()})
	fmt.Println("err:", err)

	_, err = client.UpdateTicket(ctx, &tickets.UpdateTicketIn{
		ID:          1,
		CategoryID:  pointers.New[uint32](2),
		Name:        pointers.New[string]("update ticket name"),
		Description: pointers.New[string]("update ticket description"),
		Price:       pointers.New[float32](123.45),
		Quantity:    pointers.New[uint32](2),
	})
	fmt.Println("err:", err)

	_, err = client.DeleteTicket(ctx, &tickets.DeleteTicketIn{ID: 3})
	fmt.Println("err:", err)
}
