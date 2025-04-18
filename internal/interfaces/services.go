package interfaces

//go:generate mockgen -source=services.go -destination=../../mocks/services/tickets_service.go -package=mockservices -exclude_interfaces=RespondsService,ToysService
type TicketsService interface {
	TicketsRepository
}

//go:generate mockgen -source=services.go -destination=../../mocks/services/responds_service.go -package=mockservices -exclude_interfaces=TicketsService,ToysService
type RespondsService interface {
	RespondsRepository
}

//go:generate mockgen -source=services.go -destination=../../mocks/services/toys_service.go -package=mockservices -exclude_interfaces=RespondsService,TicketsService
type ToysService interface {
	ToysRepository
}
