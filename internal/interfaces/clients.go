package interfaces

import (
	"github.com/DKhorkov/hmtm-toys/api/protobuf/generated/go/toys"
)

type ToysClient interface {
	toys.CategoriesServiceClient
	toys.TagsServiceClient
	toys.MastersServiceClient
}
