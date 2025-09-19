package usecase_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/internal/order/infrastructure/repositories"
	"github.com/jailtonjunior94/order/pkg/database"
	"github.com/jailtonjunior94/order/pkg/database/uow"

	"github.com/stretchr/testify/suite"
)

type CreateSuite struct {
	suite.Suite

	db              *sql.DB
	ctx             context.Context
	uow             uow.UnitOfWork
	orderRepository interfaces.OrderRepository
	cockroach       *database.CockroachDBContainer
}

func TestCreateSuite(t *testing.T) {
	suite.Run(t, new(CreateSuite))
}

func (s *CreateSuite) SetupSuite() {
	s.ctx = context.Background()
	s.cockroach = database.SetupCockroachDB(s.ctx, s.T(), "order", "testuser", "testpassword")

	s.db, _ = sql.Open("postgres", s.cockroach.ConnectionString(s.T()))
	database.RunMigration(s.T(), s.db, "order", "../../migrations")

	s.uow = uow.NewUnitOfWork(s.db)
	s.orderRepository = repositories.NewOrderRepository(s.uow.DBTX(), nil)
}

func (s *CreateSuite) TearDownSuite() {
	s.cockroach.Terminate(s.ctx, s.T())
}

func (s *CreateSuite) TestExecute() {
	fmt.Println("Implementar os testes")
}
