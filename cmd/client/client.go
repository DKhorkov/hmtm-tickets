package main

import (
	"context"
	"fmt"

	"github.com/DKhorkov/hmtm-tickets/api/protobuf/generated/go/tickets"
	"github.com/DKhorkov/libs/requestid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	requestID := requestid.New()

	ticketID, err := client.CreateTicket(context.Background(), &tickets.CreateTicketIn{
		RequestID:   requestID,
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

	// ticket, err := client.GetTicket(context.Background(), &tickets.GetTicketIn{
	//	RequestID: requestID,
	//	ID:        14,
	// })
	// fmt.Println("ticket by ID:", ticket, "err:", err)

	allTickets, err := client.GetTickets(context.Background(), &tickets.GetTicketsIn{RequestID: requestID})
	fmt.Println("allTickets:", allTickets, "err:", err)

	// userTickets, err := client.GetUserTickets(context.Background(), &tickets.GetUserTicketsIn{
	//	RequestID: requestID,
	//	UserID:    1},
	//)
	// fmt.Println("userTickets:", userTickets, "err:", err)
	//
	// respondsID, err := client.RespondToTicket(context.Background(), &tickets.RespondToTicketIn{
	//	RequestID: requestID,
	//	TicketID:  14,
	//	UserID:    1,
	// })
	// fmt.Println("respondsID:", respondsID, "err:", err)
	//
	// respond, err := client.GetRespond(context.Background(), &tickets.GetRespondIn{
	//	RequestID: requestID,
	//	ID:        1,
	// })
	// fmt.Println("respond:", respond, "err:", err)
	//
	// userResponds, err := client.GetUserResponds(context.Background(), &tickets.GetUserRespondsIn{
	//	RequestID: requestID,
	//	UserID:    1,
	// })
	// fmt.Println("userResponds:", userResponds, "err:", err)
	//
	// ticketResponds, err := client.GetTicketResponds(context.Background(), &tickets.GetTicketRespondsIn{
	//	RequestID: requestID,
	//	TicketID:  14,
	// })
	// fmt.Println("ticketResponds:", ticketResponds, "err:", err)
}
