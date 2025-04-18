package tickets

import (
	"github.com/DKhorkov/libs/pointers"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DKhorkov/hmtm-tickets/api/protobuf/generated/go/tickets"
	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	"github.com/stretchr/testify/require"
)

func TestMapTicketToOut(t *testing.T) {
	testCases := []struct {
		name     string
		ticket   entities.Ticket
		expected *tickets.GetTicketOut
	}{
		{
			name: "full ticket with attachments",
			ticket: entities.Ticket{
				ID:          1,
				UserID:      2,
				CategoryID:  3,
				Name:        "Test Ticket",
				Description: "Test Description",
				Price:       pointers.New[float32](99),
				Quantity:    5,
				TagIDs:      []uint32{1, 2, 3},
				Attachments: []entities.Attachment{
					{
						ID:        1,
						TicketID:  1,
						Link:      "attachment1.jpg",
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:        2,
						TicketID:  1,
						Link:      "attachment2.jpg",
						CreatedAt: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC),
					},
				},
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			expected: &tickets.GetTicketOut{
				ID:          1,
				UserID:      2,
				CategoryID:  3,
				Name:        "Test Ticket",
				Description: "Test Description",
				Price:       pointers.New[float32](99),
				Quantity:    5,
				TagIDs:      []uint32{1, 2, 3},
				Attachments: []*tickets.Attachment{
					{
						ID:        1,
						TicketID:  1,
						Link:      "attachment1.jpg",
						CreatedAt: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
						UpdatedAt: timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
					},
					{
						ID:        2,
						TicketID:  1,
						Link:      "attachment2.jpg",
						CreatedAt: timestamppb.New(time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC)),
						UpdatedAt: timestamppb.New(time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC)),
					},
				},
				CreatedAt: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
				UpdatedAt: timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "ticket without attachments and price",
			ticket: entities.Ticket{
				ID:          2,
				UserID:      3,
				CategoryID:  4,
				Name:        "Simple Ticket",
				Description: "Simple Description",
				Price:       nil,
				Quantity:    1,
				TagIDs:      []uint32{},
				Attachments: []entities.Attachment{},
				CreatedAt:   time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2023, 2, 2, 0, 0, 0, 0, time.UTC),
			},
			expected: &tickets.GetTicketOut{
				ID:          2,
				UserID:      3,
				CategoryID:  4,
				Name:        "Simple Ticket",
				Description: "Simple Description",
				Price:       nil,
				Quantity:    1,
				TagIDs:      []uint32{},
				Attachments: []*tickets.Attachment{},
				CreatedAt:   timestamppb.New(time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)),
				UpdatedAt:   timestamppb.New(time.Date(2023, 2, 2, 0, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "minimal ticket",
			ticket: entities.Ticket{
				ID:        3,
				UserID:    4,
				CreatedAt: time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 3, 2, 0, 0, 0, 0, time.UTC),
			},
			expected: &tickets.GetTicketOut{
				ID:          3,
				UserID:      4,
				CategoryID:  0,
				Name:        "",
				Description: "",
				Price:       nil,
				Quantity:    0,
				TagIDs:      nil, // Пустой слайс в entities.Ticket преобразуется в nil в Protobuf
				Attachments: []*tickets.Attachment{},
				CreatedAt:   timestamppb.New(time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC)),
				UpdatedAt:   timestamppb.New(time.Date(2023, 3, 2, 0, 0, 0, 0, time.UTC)),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapTicketToOut(tc.ticket)

			// Проверка полей верхнего уровня
			require.Equal(t, tc.expected.ID, result.ID)
			require.Equal(t, tc.expected.UserID, result.UserID)
			require.Equal(t, tc.expected.CategoryID, result.CategoryID)
			require.Equal(t, tc.expected.Name, result.Name)
			require.Equal(t, tc.expected.Description, result.Description)
			require.Equal(t, tc.expected.Price, result.Price)
			require.Equal(t, tc.expected.Quantity, result.Quantity)
			require.Equal(t, tc.expected.TagIDs, result.TagIDs)

			// Проверка вложений
			require.Equal(t, len(tc.expected.Attachments), len(result.Attachments))
			for i, expectedAttachment := range tc.expected.Attachments {
				actualAttachment := result.Attachments[i]
				require.Equal(t, expectedAttachment.ID, actualAttachment.ID)
				require.Equal(t, expectedAttachment.TicketID, actualAttachment.TicketID)
				require.Equal(t, expectedAttachment.Link, actualAttachment.Link)
				require.Equal(t, expectedAttachment.CreatedAt.AsTime(), actualAttachment.CreatedAt.AsTime())
				require.Equal(t, expectedAttachment.UpdatedAt.AsTime(), actualAttachment.UpdatedAt.AsTime())
			}

			// Проверка временных меток
			require.Equal(t, tc.expected.CreatedAt.AsTime(), result.CreatedAt.AsTime())
			require.Equal(t, tc.expected.UpdatedAt.AsTime(), result.UpdatedAt.AsTime())
		})
	}
}
