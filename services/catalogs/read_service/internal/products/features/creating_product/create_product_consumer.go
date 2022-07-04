package creating_product

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/mehdihadeli/store-golang-microservice-sample/pkg/tracing"
	"github.com/mehdihadeli/store-golang-microservice-sample/pkg/utils"
	"github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/read_service/internal/products/configurations"
	"github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/read_service/internal/products/contracts/grpc/kafka_messages"
	"github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/read_service/internal/products/features/creating_product/dtos"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	"time"
)

type createProductConsumer struct {
	*configurations.ProductKafkaConsumersConfigurations
}

func NewCreateProductConsumer(config *configurations.ProductKafkaConsumersConfigurations) *createProductConsumer {
	return &createProductConsumer{config}
}

const (
	retryAttempts = 3
	retryDelay    = 300 * time.Millisecond
)

var (
	retryOptions = []retry.Option{retry.Attempts(retryAttempts), retry.Delay(retryDelay), retry.DelayType(retry.BackOffDelay)}
)

func (c *createProductConsumer) Consume(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	c.Infrastructure.Metrics.CreateProductKafkaMessages.Inc()

	ctx, span := tracing.StartKafkaConsumerTracerSpan(ctx, m.Headers, "readerMessageProcessor.processProductCreated")
	defer span.Finish()

	msg := &kafka_messages.ProductCreated{}
	if err := proto.Unmarshal(m.Value, msg); err != nil {
		c.Infrastructure.Log.WarnMsg("proto.Unmarshal", err)
		tracing.TraceErr(span, err)
		c.CommitErrMessage(ctx, r, m)

		return
	}

	p := msg.GetProduct()
	command := NewCreateProduct(p.GetProductID(), p.GetName(), p.GetDescription(), p.GetPrice(), p.GetCreatedAt().AsTime(), p.GetUpdatedAt().AsTime())
	if err := c.Infrastructure.Validator.StructCtx(ctx, command); err != nil {
		tracing.TraceErr(span, err)
		c.Infrastructure.Log.WarnMsg("validate", err)
		c.CommitErrMessage(ctx, r, m)
		return
	}

	if err := retry.Do(func() error {
		result, err := c.Mediator.Send(ctx, command)
		if err != nil {
			tracing.TraceErr(span, err)
			return err
		}

		_, ok := result.(*dtos.CreateProductResponseDto)
		err = utils.CheckType(ok)
		if err != nil {
			tracing.TraceErr(span, err)
			return err
		}

		c.Infrastructure.Log.Infof("(product created) id: {%s}", command.ProductID)

		return nil
	}, append(retryOptions, retry.Context(ctx))...); err != nil {
		c.Infrastructure.Log.WarnMsg("CreateProduct.Handle", err)
		tracing.TraceErr(span, err)
		c.Infrastructure.Metrics.ErrorKafkaMessages.Inc()
		return
	}

	c.CommitMessage(ctx, r, m)
}