package configurations

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/read_service/internal/products/delivery"
	getting_product_by_id "github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/read_service/internal/products/features/get_product_by_id/endpoints/v1"
	getting_products "github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/read_service/internal/products/features/getting_products/endpoints/v1"
	searching_products "github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/read_service/internal/products/features/searching_products/endpoints/v1"
)

func (c *productsModuleConfigurator) configEndpoints(ctx context.Context, group *echo.Group) {
	fmt.Print(c)

	productEndpointBase := &delivery.ProductEndpointBase{
		ProductsGroup:                group,
		InfrastructureConfigurations: c.InfrastructureConfigurations,
	}

	// GetProducts
	getProductsEndpoint := getting_products.NewGetProductsEndpoint(productEndpointBase)
	getProductsEndpoint.MapRoute()

	// SearchProducts
	searchProductsEndpoint := searching_products.NewSearchProductsEndpoint(productEndpointBase)
	searchProductsEndpoint.MapRoute()

	// GetProductById
	getProductByIdEndpoint := getting_product_by_id.NewGetProductByIdEndpoint(productEndpointBase)
	getProductByIdEndpoint.MapRoute()
}
