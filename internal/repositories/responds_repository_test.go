//go:build integration

package repositories_test

import (
	"context"
	"database/sql"
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
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

const (
	driver = "sqlite3"
	//dsn    = "file::memory:?cache=shared"
	dsn              = "../../test.db"
	migrationsDir    = "/migrations"
	gooseZeroVersion = 0
)

func TestRespondsRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RespondsRepositoryTestSuite))
}

type RespondsRepositoryTestSuite struct {
	suite.Suite

	cwd                string
	ctx                context.Context
	dbConnector        db.Connector
	connection         *sql.Conn
	respondsRepository *repositories.RespondsRepository
	logger             *mocklogging.MockLogger
	traceProvider      *mocktracing.MockProvider
	spanConfig         tracing.SpanConfig
}

func (s *RespondsRepositoryTestSuite) SetupSuite() {
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
	s.respondsRepository = repositories.NewRespondsRepository(s.dbConnector, s.logger, s.traceProvider, s.spanConfig)
}

func (s *RespondsRepositoryTestSuite) SetupTest() {
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

func (s *RespondsRepositoryTestSuite) TearDownTest() {
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

func (s *RespondsRepositoryTestSuite) TearDownSuite() {
	s.NoError(s.dbConnector.Close())
}

func (s *RespondsRepositoryTestSuite) TestRespondToTicketSuccess() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	respondData := entities.RespondToTicketDTO{
		TicketID: 1,
		MasterID: 2,
		Price:    100.50,
		Comment:  pointers.New("Test comment"),
	}

	// Error and zero id due to returning nil ID after insert operation
	// SQLite inner realization without AUTO_INCREMENT for SERIAL PRIMARY KEY
	id, err := s.respondsRepository.RespondToTicket(s.ctx, respondData)
	s.Error(err)
	s.Zero(id)
}

func (s *RespondsRepositoryTestSuite) TestGetRespondByIDExisting() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	createdAt := time.Now().UTC()
	updatedAt := createdAt
	comment := pointers.New("Test respond")
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO responds (id, ticket_id, master_id, price, comment, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, 150.75, *comment, createdAt, updatedAt,
	)
	s.NoError(err)

	respond, err := s.respondsRepository.GetRespondByID(s.ctx, 1)
	s.NoError(err)
	s.NotNil(respond)
	s.Equal(uint64(1), respond.TicketID)
	s.Equal(uint64(2), respond.MasterID)
	s.InDelta(150.75, respond.Price, 0.01)
	s.Equal(*comment, *respond.Comment)
	s.WithinDuration(createdAt, respond.CreatedAt, time.Second)
	s.WithinDuration(updatedAt, respond.UpdatedAt, time.Second)
}

func (s *RespondsRepositoryTestSuite) TestGetRespondByIDNonExisting() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	respond, err := s.respondsRepository.GetRespondByID(s.ctx, 999)
	s.Error(err)
	s.Nil(respond)
}

func (s *RespondsRepositoryTestSuite) TestGetTicketRespondsWithExisting() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	ticketID := uint64(1)
	comment1 := pointers.New("Respond 1")
	comment2 := pointers.New("Respond 2")
	createdAt := time.Now().UTC()
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO responds (id, ticket_id, master_id, price, comment, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?)",
		1, ticketID, 2, 100.00, *comment1, createdAt, createdAt,
		2, ticketID, 3, 200.00, *comment2, createdAt, createdAt,
	)
	s.NoError(err)

	responds, err := s.respondsRepository.GetTicketResponds(s.ctx, ticketID)
	s.NoError(err)
	s.NotEmpty(responds)
	s.Equal(2, len(responds))
}

func (s *RespondsRepositoryTestSuite) TestGetTicketRespondsWithoutExisting() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	responds, err := s.respondsRepository.GetTicketResponds(s.ctx, 999)
	s.NoError(err)
	s.Empty(responds)
}

func (s *RespondsRepositoryTestSuite) TestGetMasterRespondsWithExisting() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	masterID := uint64(2)
	comment1 := pointers.New("Respond 1")
	comment2 := pointers.New("Respond 2")
	createdAt := time.Now().UTC()
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO responds (id, ticket_id, master_id, price, comment, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?)",
		1, 1, masterID, 100.00, *comment1, createdAt, createdAt,
		2, 1, masterID, 200.00, *comment2, createdAt, createdAt,
	)
	s.NoError(err)

	responds, err := s.respondsRepository.GetMasterResponds(s.ctx, masterID)
	s.NoError(err)
	s.NotEmpty(responds)
	s.Equal(2, len(responds))
}

func (s *RespondsRepositoryTestSuite) TestGetMasterRespondsWithoutExisting() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	responds, err := s.respondsRepository.GetMasterResponds(s.ctx, 999)
	s.NoError(err)
	s.Empty(responds)
}

func (s *RespondsRepositoryTestSuite) TestUpdateRespondSuccess() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	createdAt := time.Now().UTC()
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO responds (id, ticket_id, master_id, price, comment, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, 100.00, "Old comment", createdAt, createdAt,
	)
	s.NoError(err)

	newPrice := pointers.New[float32](200.50)
	newComment := pointers.New("Updated comment")
	respondData := entities.UpdateRespondDTO{
		ID:      1,
		Price:   newPrice,
		Comment: newComment,
	}
	err = s.respondsRepository.UpdateRespond(s.ctx, respondData)
	s.NoError(err)

	// Проверка обновления
	rows, err := s.connection.QueryContext(
		s.ctx,
		"SELECT price, comment FROM responds WHERE id = ?",
		1,
	)
	s.NoError(err)

	defer func() {
		s.NoError(rows.Close())
	}()

	s.True(rows.Next())
	var price float32
	var comment sql.NullString
	s.NoError(rows.Scan(&price, &comment))
	s.InDelta(*newPrice, price, 0.01)
	s.True(comment.Valid)
	s.Equal(*newComment, comment.String)
}

func (s *RespondsRepositoryTestSuite) TestUpdateRespondNoPrice() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	createdAt := time.Now().UTC()
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO responds (id, ticket_id, master_id, price, comment, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, 100.00, "Old comment", createdAt, createdAt,
	)
	s.NoError(err)

	newComment := pointers.New("Updated comment")
	respondData := entities.UpdateRespondDTO{
		ID:      1,
		Price:   nil,
		Comment: newComment,
	}
	err = s.respondsRepository.UpdateRespond(s.ctx, respondData)
	s.NoError(err)

	// Проверка, что price не изменился
	rows, err := s.connection.QueryContext(
		s.ctx,
		"SELECT price, comment FROM responds WHERE id = ?",
		1,
	)
	s.NoError(err)

	defer func() {
		s.NoError(rows.Close())
	}()

	s.True(rows.Next())
	var price float32
	var comment sql.NullString
	s.NoError(rows.Scan(&price, &comment))
	s.InDelta(100.00, price, 0.01)
	s.True(comment.Valid)
	s.Equal(*newComment, comment.String)
}

func (s *RespondsRepositoryTestSuite) TestUpdateRespondNullComment() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	createdAt := time.Now().UTC()
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO responds (id, ticket_id, master_id, price, comment, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, 100.00, "Old comment", createdAt, createdAt,
	)
	s.NoError(err)

	newPrice := pointers.New[float32](200.50)
	respondData := entities.UpdateRespondDTO{
		ID:      1,
		Price:   newPrice,
		Comment: nil,
	}
	err = s.respondsRepository.UpdateRespond(s.ctx, respondData)
	s.NoError(err)

	// Проверка, что comment стал NULL
	rows, err := s.connection.QueryContext(
		s.ctx,
		"SELECT price, comment FROM responds WHERE id = ?",
		1,
	)
	s.NoError(err)

	defer func() {
		s.NoError(rows.Close())
	}()

	s.True(rows.Next())
	var price float32
	var comment sql.NullString
	s.NoError(rows.Scan(&price, &comment))
	s.InDelta(*newPrice, price, 0.01)
	s.False(comment.Valid)
}

func (s *RespondsRepositoryTestSuite) TestDeleteRespondSuccess() {
	s.traceProvider.
		EXPECT().
		Span(gomock.Any(), gomock.Any()).
		Return(context.Background(), mocktracing.NewMockSpan()).
		Times(1)

	createdAt := time.Now().UTC()
	_, err := s.connection.ExecContext(
		s.ctx,
		"INSERT INTO responds (id, ticket_id, master_id, price, comment, created_at, updated_at) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?)",
		1, 1, 2, 100.00, "Old comment", createdAt, createdAt,
	)
	s.NoError(err)

	err = s.respondsRepository.DeleteRespond(s.ctx, 1)
	s.NoError(err)

	// Проверка, что запись удалена
	rows, err := s.connection.QueryContext(
		s.ctx,
		"SELECT id FROM responds WHERE id = ?",
		1,
	)
	s.NoError(err)

	defer func() {
		s.NoError(rows.Close())
	}()

	s.False(rows.Next())
}
