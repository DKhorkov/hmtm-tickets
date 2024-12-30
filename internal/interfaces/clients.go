package interfaces

import (
	"github.com/DKhorkov/hmtm-toys/api/protobuf/generated/go/toys"
)

type ToysGrpcClient interface {
	toys.CategoriesServiceClient
	toys.TagsServiceClient
	toys.MastersServiceClient
}
