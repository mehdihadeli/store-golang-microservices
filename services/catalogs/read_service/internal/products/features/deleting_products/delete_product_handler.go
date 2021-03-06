package deleting_products

import (
	"context"
	"github.com/mehdihadeli/store-golang-microservice-sample/pkg/logger"
	"github.com/mehdihadeli/store-golang-microservice-sample/pkg/mediatr"
	"github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/read_service/config"
	"github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/read_service/internal/products/contracts"
	"github.com/opentracing/opentracing-go"
)

type DeleteProductHandler struct {
	log             logger.Logger
	cfg             *config.Config
	mongoRepository contracts.ProductRepository
	redisRepository contracts.ProductCacheRepository
}

func NewDeleteProductHandler(log logger.Logger, cfg *config.Config, repository contracts.ProductRepository, redisRepository contracts.ProductCacheRepository) *DeleteProductHandler {
	return &DeleteProductHandler{log: log, cfg: cfg, mongoRepository: repository, redisRepository: redisRepository}
}

func (c *DeleteProductHandler) Handle(ctx context.Context, command *DeleteProduct) (*mediatr.Unit, error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "DeleteProductHandler.Handle")
	defer span.Finish()

	if err := c.mongoRepository.DeleteProductByID(ctx, command.ProductID); err != nil {
		return nil, err
	}

	c.log.Infof("(product deleted) id: {%s}", command.ProductID)

	c.redisRepository.DelProduct(ctx, command.ProductID.String())

	return &mediatr.Unit{}, nil
}
