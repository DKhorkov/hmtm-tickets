package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/DKhorkov/hmtm-tickets/api/protobuf/generated/go/tickets"
	"github.com/DKhorkov/libs/requestid"
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
		UserID:      1,
		CategoryID:  1,
		TagIDs:      []uint32{1},
		Name:        "test ticket 2",
		Description: "test description",
		Price:       20,
		Quantity:    10,
		Attachments: []string{"someref", "anotherref"},
	})
	fmt.Println("ticketID:", ticketID, "err:", err)

	ticket, err := client.GetTicket(ctx, &tickets.GetTicketIn{
		ID: 1,
	})
	fmt.Println("ticket by ID:", ticket, "err:", err)

	allTickets, err := client.GetTickets(ctx, &emptypb.Empty{})
	fmt.Println("allTickets:", allTickets, "err:", err)

	userTickets, err := client.GetUserTickets(ctx, &tickets.GetUserTicketsIn{
		UserID: 1},
	)
	fmt.Println("userTickets:", userTickets, "err:", err)

	respondsID, err := client.RespondToTicket(ctx, &tickets.RespondToTicketIn{
		TicketID: 1,
		UserID:   1,
	})
	fmt.Println("respondsID:", respondsID, "err:", err)

	respond, err := client.GetRespond(ctx, &tickets.GetRespondIn{
		ID: 1,
	})
	fmt.Println("respond:", respond, "err:", err)

	userResponds, err := client.GetUserResponds(ctx, &tickets.GetUserRespondsIn{
		UserID: 1,
	})
	fmt.Println("userResponds:", userResponds, "err:", err)

	ticketResponds, err := client.GetTicketResponds(ctx, &tickets.GetTicketRespondsIn{
		TicketID: 1,
	})
	fmt.Println("ticketResponds:", ticketResponds, "err:", err)
}
