package searching_product

import (
	"context"
	"github.com/mehdihadeli/store-golang-microservice-sample/pkg/logger"
	"github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/write_service/config"
	"github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/write_service/internal/products/contracts/repositories"
	"github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/write_service/internal/products/mappers"
	"github.com/opentracing/opentracing-go"
)

type SearchProductsHandler struct {
	log    logger.Logger
	cfg    *config.Config
	pgRepo repositories.ProductRepository
}

func NewSearchProductsHandler(log logger.Logger, cfg *config.Config, pgRepo repositories.ProductRepository) *SearchProductsHandler {
	return &SearchProductsHandler{log: log, cfg: cfg, pgRepo: pgRepo}
}

func (c *SearchProductsHandler) Handle(ctx context.Context, query SearchProducts) (*SearchProductsResponseDto, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "SearchProductsHandler.Handle")
	defer span.Finish()

	products, err := c.pgRepo.SearchProducts(ctx, query.SearchText, query.ListQuery)
	if err != nil {
		return nil, err
	}

	listResultDto := mappers.ListResultToListResultDto(products, mappers.ProductsToProductsDto)

	return &SearchProductsResponseDto{Products: listResultDto}, nil
}