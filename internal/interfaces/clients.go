package interfaces

import (
	"github.com/DKhorkov/hmtm-toys/api/protobuf/generated/go/toys"
)

//go:generate mockgen -source=clients.go -destination=../../mocks/clients/toys_client.go -package=mockclients -exclude_interfaces=
type ToysClient interface {
	toys.CategoriesServiceClient
	toys.TagsServiceClient
	toys.MastersServiceClient
}
