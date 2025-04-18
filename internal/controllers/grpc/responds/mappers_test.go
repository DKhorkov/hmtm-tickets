package responds

import (
	"github.com/DKhorkov/libs/pointers"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DKhorkov/hmtm-tickets/api/protobuf/generated/go/tickets"
	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	"github.com/stretchr/testify/require"
)

func TestMapRespondOut(t *testing.T) {
	testCases := []struct {
		name     string
		respond  entities.Respond
		expected *tickets.GetRespondOut
	}{
		{
			name: "full respond with comment",
			respond: entities.Respond{
				ID:        1,
				TicketID:  2,
				MasterID:  3,
				Price:     99.99,
				Comment:   pointers.New("Test"),
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			expected: &tickets.GetRespondOut{
				ID:        1,
				TicketID:  2,
				MasterID:  3,
				Price:     99.99,
				Comment:   pointers.New("Test"),
				CreatedAt: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
				UpdatedAt: timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "respond without comment",
			respond: entities.Respond{
				ID:        2,
				TicketID:  3,
				MasterID:  4,
				Price:     49.99,
				Comment:   nil,
				CreatedAt: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 2, 2, 0, 0, 0, 0, time.UTC),
			},
			expected: &tickets.GetRespondOut{
				ID:        2,
				TicketID:  3,
				MasterID:  4,
				Price:     49.99,
				Comment:   nil,
				CreatedAt: timestamppb.New(time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)),
				UpdatedAt: timestamppb.New(time.Date(2023, 2, 2, 0, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "minimal respond",
			respond: entities.Respond{
				ID:        3,
				TicketID:  4,
				MasterID:  5,
				Price:     0,
				Comment:   nil,
				CreatedAt: time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 3, 2, 0, 0, 0, 0, time.UTC),
			},
			expected: &tickets.GetRespondOut{
				ID:        3,
				TicketID:  4,
				MasterID:  5,
				Price:     0,
				Comment:   nil,
				CreatedAt: timestamppb.New(time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC)),
				UpdatedAt: timestamppb.New(time.Date(2023, 3, 2, 0, 0, 0, 0, time.UTC)),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapRespondOut(tc.respond)

			// Проверка полей
			require.Equal(t, tc.expected.ID, result.ID)
			require.Equal(t, tc.expected.TicketID, result.TicketID)
			require.Equal(t, tc.expected.MasterID, result.MasterID)
			require.Equal(t, tc.expected.Price, result.Price)
			require.Equal(t, tc.expected.Comment, result.Comment)

			// Проверка временных меток
			require.Equal(t, tc.expected.CreatedAt.AsTime(), result.CreatedAt.AsTime())
			require.Equal(t, tc.expected.UpdatedAt.AsTime(), result.UpdatedAt.AsTime())
		})
	}
}
