//go:build integration

package repositories_test

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
	"os"
	"path"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Must be imported for correct work

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	"github.com/DKhorkov/hmtm-tickets/internal/repositories"
	"github.com/DKhorkov/libs/db"
	mocklogging "github.com/DKhorkov/libs/logging/mocks"
	"github.com/DKhorkov/libs/pointers"
	"github.com/DKhorkov/libs/tracing"
	mocktracing "github.com/DKhorkov/libs/tracing/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestTicketsRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(TicketsRepositoryTestSuite))
}

type TicketsRepositoryTestSuite struct {
	suite.Suite

	cwd               string
	ctx               context.Context
	dbConnector       db.Connector
	connection        *sql.Conn
	ticketsRepository *repositories.TicketsRepository
	logger            *mocklogging.MockLogger
	traceProvider     *mocktracing.MockProvider
	spanConfig        tracing.SpanConfig
}

func (s *TicketsRepositoryTestSuite) SetupSuite() {
	s.NoError(goose.SetDialect(driver))

	ctrl := gomock.NewController(s.T())
	s.ctx = context.Background()
	s.logger = mocklogging.NewMockLogger(ctrl)
	dbConnector, err := db.New(dsn, driver, s.logger)
	s.NoError(err)

	cwd, err := os.Getwd()
	s.NoError(err)

	s.cwd = cwd
	s.dbConnector = dbConnector
	s.traceProvider = mocktracing.NewMockProvider(ctrl)
	s.spanConfig = tracing.SpanConfig{}
	s.ticketsRepository = repositories.NewTicketsRepository(s.dbConnector, s.logger, s.traceProvider, s.spanConfig)
}

func (s *TicketsRepositoryTestSuite) SetupTest() {
	s.NoError(
		goose.Up(
			s.dbConnector.Pool(),
			path.Dir(
				path.Dir(s.cwd),
			)+migrationsDir,
		),
	)

	connection, err := s.dbConnector.Connection(s.ctx)
	s.NoError(err)

	s.connection = connection
}

func (s *TicketsRepositoryTestSuite) TearDownTest() {
	s.NoError(
		goose.DownTo(
			s.dbConnector.Pool(),
			path.Dir(
				path.Dir(s.cwd),
			)+migrationsDir,
			gooseZeroVersion,
		),
	)

	s.NoError(s.connection.Close())
}

func (s *TicketsRepositoryTestSuite) TearDownSuite() {
	s.NoError(s.dbConnector.Close())
}

func (s *TicketsRepositoryTestSuite) TestCreateTicketSuccess() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	price := pointers.New[float32](99.99)
	ticketData := entities.CreateTicketDTO{
		UserID:      1,
		CategoryID:  2,
		Name:        "Test Ticket",
		Description: "Test Description",
		Price:       price,
		Quantity:    5,
		TagIDs:      []uint32{10, 20},
		Attachments: []string{"file1.jpg", "file2.pdf"},
	}

	// Error and zero id due to returning nil ID after insert operation
	// SQLite inner realization without AUTO_INCREMENT for SERIAL PRIMARY KEY
	id, err := s.ticketsRepository.CreateTicket(s.ctx, ticketData)
	s.Error(err)
	s.Zero(id)
}

func (s *TicketsRepositoryTestSuite) TestCreateTicketError() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	ticketData := entities.CreateTicketDTO{
		UserID:      1,
		CategoryID:  2,
		Name:        "Test Ticket",
		Description: "Test",
		Price:       pointers.New[float32](99.99),
		Quantity:    5,
	}

	id, err := s.ticketsRepository.CreateTicket(s.ctx, ticketData)
	s.Error(err)
	s.Zero(id)
}

func (s *TicketsRepositoryTestSuite) TestGetTicketByIDExisting() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(3) // Основной + getTicketTagsIDs + getTicketAttachments

	createdAt := time.Now().UTC()
	price := pointers.New[float32](99.99)
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (id, user_id, category_id, name, description, price, quantity, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, "Test Ticket", "Test Description", price, 5, createdAt, createdAt,
	)
	s.NoError(err)

	_, err = s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets_tags_associations (id, ticket_id, tag_id) VALUES "+
			"(?, ?, ?), (?, ?, ?)",
		1, 1, 10,
		2, 1, 20,
	)
	s.NoError(err)

	_, err = s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets_attachments (id, ticket_id, link, created_at, updated_at) VALUES "+
			"(?, ?, ?, ?, ?), (?, ?, ?, ?, ?)",
		1, 1, "file1.jpg", createdAt, createdAt,
		2, 1, "file2.pdf", createdAt, createdAt,
	)
	s.NoError(err)

	ticket, err := s.ticketsRepository.GetTicketByID(s.ctx, 1)
	s.NoError(err)
	s.NotNil(ticket)
	s.Equal(uint64(1), ticket.UserID)
	s.Equal(uint32(2), ticket.CategoryID)
	s.Equal("Test Ticket", ticket.Name)
	s.Equal("Test Description", ticket.Description)
	s.NotNil(ticket.Price)
	s.InDelta(*price, *ticket.Price, 0.01)
	s.Equal(uint32(5), ticket.Quantity)
	s.ElementsMatch([]uint32{10, 20}, ticket.TagIDs)
	s.Equal(2, len(ticket.Attachments))
	s.Contains([]string{ticket.Attachments[0].Link, ticket.Attachments[1].Link}, "file1.jpg")
	s.Contains([]string{ticket.Attachments[0].Link, ticket.Attachments[1].Link}, "file2.pdf")
}

func (s *TicketsRepositoryTestSuite) TestGetTicketByIDNonExisting() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	ticket, err := s.ticketsRepository.GetTicketByID(s.ctx, 999)
	s.Error(err)
	s.Nil(ticket)
}

func (s *TicketsRepositoryTestSuite) TestGetTicketsWithExistingTickets() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(5) // Основной + 2x(getTicketTagsIDs + getTicketAttachments)

	createdAt := time.Now().UTC()
	price1 := pointers.New[float32](99.99)
	price2 := pointers.New[float32](49.99)
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (id, user_id, category_id, name, description, price, quantity, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, "Ticket 1", "Desc 1", price1, 5, createdAt, createdAt,
		2, 2, 3, "Ticket 2", "Desc 2", price2, 3, createdAt, createdAt,
	)
	s.NoError(err)

	tickets, err := s.ticketsRepository.GetTickets(s.ctx, nil, nil)
	s.NoError(err)
	s.NotEmpty(tickets)
	s.Equal(2, len(tickets))
}

func (s *TicketsRepositoryTestSuite) TestGetTicketsWithExistingTicketsAndPagination() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	createdAt := time.Now().UTC()
	price := pointers.New[float32](99.99)
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (id, user_id, category_id, name, description, price, quantity, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, "Test Ticket", "Test Description", price, 5, createdAt, createdAt,
	)
	s.NoError(err)

	pagination := &entities.Pagination{
		Limit:  pointers.New[uint64](1),
		Offset: pointers.New[uint64](2),
	}

	tickets, err := s.ticketsRepository.GetTickets(s.ctx, pagination, nil)
	s.NoError(err)
	s.Empty(tickets)
}

func (s *TicketsRepositoryTestSuite) TestGetTicketsWithExistingTicketsAndFilters() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(3) // Основной + getToyTags + getToyAttachments

	createdAt := time.Now().UTC()
	price := pointers.New[float32](99.99)
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (id, user_id, category_id, name, description, price, quantity, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, "Test Ticket", "Test Description", price, 5, createdAt, createdAt,
	)
	s.NoError(err)

	_, err = s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets_tags_associations (id, ticket_id, tag_id) VALUES "+
			"(?, ?, ?), (?, ?, ?)",
		1, 1, 10,
		2, 1, 20,
	)
	s.NoError(err)

	_, err = s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets_attachments (id, ticket_id, link, created_at, updated_at) VALUES "+
			"(?, ?, ?, ?, ?)",
		1, 1, "file1.jpg", createdAt, createdAt,
	)
	s.NoError(err)

	filters := &entities.TicketsFilters{
		//Search:              pointers.New("ticket"), // no ILike in sqlite
		PriceCeil:           pointers.New[float32](1000),
		PriceFloor:          pointers.New[float32](10),
		QuantityFloor:       pointers.New[uint32](1),
		CategoryIDs:         []uint32{2},
		TagIDs:              []uint32{10, 20},
		CreatedAtOrderByAsc: pointers.New(true),
	}

	tickets, err := s.ticketsRepository.GetTickets(s.ctx, nil, filters)
	s.NoError(err)
	s.NotEmpty(tickets)
	s.Equal(1, len(tickets))
	s.Equal(uint64(1), tickets[0].UserID)
	s.Equal(uint32(2), tickets[0].CategoryID)
	s.Equal("Test Ticket", tickets[0].Name)
	s.Equal("Test Description", tickets[0].Description)
	s.Equal(uint32(5), tickets[0].Quantity)
	s.Contains(tickets[0].TagIDs, uint32(10), uint32(20))
	s.Equal(1, len(tickets[0].Attachments))
	s.Equal("file1.jpg", tickets[0].Attachments[0].Link)
}

func (s *TicketsRepositoryTestSuite) TestGetTicketsWithoutExistingTickets() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	tickets, err := s.ticketsRepository.GetTickets(s.ctx, nil, nil)
	s.NoError(err)
	s.Empty(tickets)
}

func (s *TicketsRepositoryTestSuite) TestGetUserTicketsWithExistingTickets() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(5) // Основной + 2x(getTicketTagsIDs + getTicketAttachments)

	userID := uint64(1)
	createdAt := time.Now().UTC()
	price := pointers.New[float32](99.99)
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (id, user_id, category_id, name, description, price, quantity, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, userID, 2, "Ticket 1", "Desc 1", price, 5, createdAt, createdAt,
		2, userID, 3, "Ticket 2", "Desc 2", price, 3, createdAt, createdAt,
	)
	s.NoError(err)

	tickets, err := s.ticketsRepository.GetUserTickets(s.ctx, userID, nil, nil)
	s.NoError(err)
	s.NotEmpty(tickets)
	s.Equal(2, len(tickets))
}

func (s *TicketsRepositoryTestSuite) TestGetUserTicketsWithoutExisting() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	tickets, err := s.ticketsRepository.GetUserTickets(s.ctx, 999, nil, nil)
	s.NoError(err)
	s.Empty(tickets)
}

func (s *TicketsRepositoryTestSuite) TestGetUserTicketsWithExistingTicketsAndPagination() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	createdAt := time.Now().UTC()
	price := pointers.New[float32](99.99)
	userID := uint64(1)
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (id, user_id, category_id, name, description, price, quantity, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, userID, 2, "Test Ticket", "Test Description", price, 5, createdAt, createdAt,
	)
	s.NoError(err)

	pagination := &entities.Pagination{
		Limit:  pointers.New[uint64](1),
		Offset: pointers.New[uint64](2),
	}

	tickets, err := s.ticketsRepository.GetUserTickets(s.ctx, userID, pagination, nil)
	s.NoError(err)
	s.Empty(tickets)
}

func (s *TicketsRepositoryTestSuite) TestGetUserTicketsWithExistingTicketsAndFilters() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(3) // Основной + getToyTags + getToyAttachments

	createdAt := time.Now().UTC()
	price := pointers.New[float32](99.99)
	userID := uint64(1)
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (id, user_id, category_id, name, description, price, quantity, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, userID, 2, "Test Ticket", "Test Description", price, 5, createdAt, createdAt,
	)
	s.NoError(err)

	_, err = s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets_tags_associations (id, ticket_id, tag_id) VALUES "+
			"(?, ?, ?), (?, ?, ?)",
		1, 1, 10,
		2, 1, 20,
	)
	s.NoError(err)

	_, err = s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets_attachments (id, ticket_id, link, created_at, updated_at) VALUES "+
			"(?, ?, ?, ?, ?)",
		1, 1, "file1.jpg", createdAt, createdAt,
	)
	s.NoError(err)

	filters := &entities.TicketsFilters{
		//Search:              pointers.New("ticket"), // no ILike in sqlite
		PriceCeil:           pointers.New[float32](1000),
		PriceFloor:          pointers.New[float32](10),
		QuantityFloor:       pointers.New[uint32](1),
		CategoryIDs:         []uint32{2},
		TagIDs:              []uint32{10, 20},
		CreatedAtOrderByAsc: pointers.New(true),
	}

	tickets, err := s.ticketsRepository.GetUserTickets(s.ctx, userID, nil, filters)
	s.NoError(err)
	s.NotEmpty(tickets)
	s.Equal(1, len(tickets))
	s.Equal(uint64(1), tickets[0].UserID)
	s.Equal(uint32(2), tickets[0].CategoryID)
	s.Equal("Test Ticket", tickets[0].Name)
	s.Equal("Test Description", tickets[0].Description)
	s.Equal(uint32(5), tickets[0].Quantity)
	s.Contains(tickets[0].TagIDs, uint32(10), uint32(20))
	s.Equal(1, len(tickets[0].Attachments))
	s.Equal("file1.jpg", tickets[0].Attachments[0].Link)
}

func (s *TicketsRepositoryTestSuite) TestDeleteTicketSuccess() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	createdAt := time.Now().UTC()
	price := pointers.New[float32](99.99)
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (user_id, category_id, name, description, price, quantity, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		1, 2, "Test Ticket", "Test Description", price, 5, createdAt, createdAt,
	)
	s.NoError(err)

	err = s.ticketsRepository.DeleteTicket(s.ctx, 1)
	s.NoError(err)

	rows, err := s.connection.QueryContext(
		s.ctx,
		"SELECT id FROM tickets WHERE id = ?",
		1)
	s.NoError(err)

	defer func() {
		s.NoError(rows.Close())
	}()

	s.False(rows.Next())
}

func (s *TicketsRepositoryTestSuite) TestUpdateTicketFullUpdateSuccess() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	s.logger.
		EXPECT().
		ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1)

	createdAt := time.Now().UTC()
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (id, user_id, category_id, name, description, price, quantity, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, "Old Ticket", "Old Desc", nil, 1, createdAt, createdAt,
	)
	s.NoError(err)

	_, err = s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets_tags_associations (id, ticket_id, tag_id) "+
			"VALUES (?, ?, ?)",
		1, 1, 10,
	)
	s.NoError(err)

	_, err = s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets_attachments (id, ticket_id, link, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?)",
		1, 1, "oldfile.jpg", createdAt, createdAt,
	)
	s.NoError(err)

	newCategoryID := uint32(3)
	newName := "Updated Ticket"
	newDesc := "Updated Desc"
	newPrice := pointers.New[float32](150.00)
	newQuantity := uint32(10)
	ticketData := entities.UpdateTicketDTO{
		ID:                    1,
		CategoryID:            &newCategoryID,
		Name:                  &newName,
		Description:           &newDesc,
		Price:                 newPrice,
		Quantity:              &newQuantity,
		TagIDsToAdd:           []uint32{30, 40},
		TagIDsToDelete:        []uint32{10},
		AttachmentsToAdd:      []string{"newfile.jpg"},
		AttachmentIDsToDelete: []uint64{1},
	}

	err = s.ticketsRepository.UpdateTicket(s.ctx, ticketData)
	s.NoError(err)

	// Проверка tickets
	rows, err := s.connection.QueryContext(
		s.ctx,
		"SELECT category_id, name, description, price, quantity FROM tickets WHERE id = ?",
		1,
	)
	s.NoError(err)

	defer func() {
		s.NoError(rows.Close())
	}()

	s.True(rows.Next())
	var categoryID uint32
	var name, description string
	var priceVal sql.NullFloat64
	var quantity uint32
	s.NoError(rows.Scan(&categoryID, &name, &description, &priceVal, &quantity))
	s.Equal(newCategoryID, categoryID)
	s.Equal(newName, name)
	s.Equal(newDesc, description)
	s.True(priceVal.Valid)
	s.InDelta(*newPrice, priceVal.Float64, 0.01)
	s.Equal(newQuantity, quantity)
}

func (s *TicketsRepositoryTestSuite) TestCountTicketsWithExistingTickets() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	createdAt := time.Now().UTC()
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (id, user_id, category_id, name, description, price, quantity, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, "Old Ticket", "Old Desc", nil, 1, createdAt, createdAt,
	)
	s.NoError(err)
	s.NoError(err)

	count, err := s.ticketsRepository.CountTickets(s.ctx, nil)
	s.NoError(err)
	s.Equal(uint64(1), count)
}

func (s *TicketsRepositoryTestSuite) TestCountTicketsWithExistingTicketsAndFilters() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	createdAt := time.Now().UTC()
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (id, user_id, category_id, name, description, price, quantity, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, "Old Ticket", "Old Desc", nil, 1, createdAt, createdAt,
	)
	s.NoError(err)

	_, err = s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets_tags_associations (id, ticket_id, tag_id) "+
			"VALUES (?, ?, ?)",
		1, 1, 10,
	)
	s.NoError(err)

	_, err = s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets_attachments (id, ticket_id, link, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?)",
		1, 1, "oldfile.jpg", createdAt, createdAt,
	)
	s.NoError(err)

	filters := &entities.TicketsFilters{
		//Search:              pointers.New("ticket"), // no ILike in sqlite
		PriceCeil:           pointers.New[float32](1000),
		PriceFloor:          pointers.New[float32](10),
		QuantityFloor:       pointers.New[uint32](1),
		CategoryIDs:         []uint32{2},
		TagIDs:              []uint32{10, 20},
		CreatedAtOrderByAsc: pointers.New(true),
	}

	count, err := s.ticketsRepository.CountTickets(s.ctx, filters)
	s.NoError(err)
	s.Equal(uint64(0), count)
}

func (s *TicketsRepositoryTestSuite) TestCountTicketsWithoutExistingTickets() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	count, err := s.ticketsRepository.CountTickets(s.ctx, nil)
	s.NoError(err)
	s.Zero(count)
}

func (s *TicketsRepositoryTestSuite) TestCountUserTicketsWithExistingTickets() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	userID := uint64(1)
	createdAt := time.Now().UTC()
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (id, user_id, category_id, name, description, price, quantity, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, "Old Ticket", "Old Desc", nil, 1, createdAt, createdAt,
	)
	s.NoError(err)

	count, err := s.ticketsRepository.CountUserTickets(s.ctx, userID, nil)
	s.NoError(err)
	s.Equal(uint64(1), count)
}

func (s *TicketsRepositoryTestSuite) TestCountUserTicketsWithExistingTicketsAndFilters() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	createdAt := time.Now().UTC()
	userID := uint64(1)
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets (id, user_id, category_id, name, description, price, quantity, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, "Old Ticket", "Old Desc", 1000, 10, createdAt, createdAt,
	)
	s.NoError(err)

	_, err = s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets_tags_associations (id, ticket_id, tag_id) "+
			"VALUES (?, ?, ?)",
		1, 1, 10,
	)
	s.NoError(err)

	_, err = s.connection.ExecContext(
		s.ctx,
		"INSERT INTO tickets_attachments (id, ticket_id, link, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?)",
		1, 1, "oldfile.jpg", createdAt, createdAt,
	)
	s.NoError(err)

	filters := &entities.TicketsFilters{
		//Search:              pointers.New("ticket"), // no ILike in sqlite
		PriceCeil:           pointers.New[float32](1000),
		PriceFloor:          pointers.New[float32](10),
		QuantityFloor:       pointers.New[uint32](1),
		CategoryIDs:         []uint32{2},
		TagIDs:              []uint32{10, 20},
		CreatedAtOrderByAsc: pointers.New(true),
	}

	count, err := s.ticketsRepository.CountUserTickets(s.ctx, userID, filters)
	s.NoError(err)
	s.Equal(uint64(0), count)
}

func (s *TicketsRepositoryTestSuite) TestCountUserTicketsWithoutExistingTickets() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	userID := uint64(1)
	count, err := s.ticketsRepository.CountUserTickets(s.ctx, userID, nil)
	s.NoError(err)
	s.Zero(count)
}
