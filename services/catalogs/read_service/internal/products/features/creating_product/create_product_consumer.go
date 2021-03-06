package creating_product

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/mehdihadeli/store-golang-microservice-sample/pkg/mediatr"
	"github.com/mehdihadeli/store-golang-microservice-sample/pkg/tracing"
	"github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/read_service/internal/products/contracts/grpc/kafka_messages"
	"github.com/mehdihadeli/store-golang-microservice-sample/services/catalogs/read_service/internal/products/delivery"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	"time"
)

type createProductConsumer struct {
	*delivery.ProductConsumersBase
}

func NewCreateProductConsumer(productConsumerBase *delivery.ProductConsumersBase) *createProductConsumer {
	return &createProductConsumer{productConsumerBase}
}

const (
	retryAttempts = 3
	retryDelay    = 300 * time.Millisecond
)

var (
	retryOptions = []retry.Option{retry.Attempts(retryAttempts), retry.Delay(retryDelay), retry.DelayType(retry.BackOffDelay)}
)

func (c *createProductConsumer) Consume(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	c.Metrics.CreateProductKafkaMessages.Inc()

	ctx, span := tracing.StartKafkaConsumerTracerSpan(ctx, m.Headers, "createProductConsumer.Consume")
	defer span.Finish()

	msg := &kafka_messages.ProductCreated{}
	if err := proto.Unmarshal(m.Value, msg); err != nil {
		c.Log.WarnMsg("proto.Unmarshal", err)
		tracing.TraceErr(span, err)
		c.CommitErrMessage(ctx, r, m)

		return
	}

	p := msg.GetProduct()
	command := NewCreateProduct(p.GetProductID(), p.GetName(), p.GetDescription(), p.GetPrice(), p.GetCreatedAt().AsTime())
	if err := c.Validator.StructCtx(ctx, command); err != nil {
		tracing.TraceErr(span, err)
		c.Log.WarnMsg("validate", err)
		c.CommitErrMessage(ctx, r, m)
		return
	}

	if err := retry.Do(func() error {
		_, err := mediatr.Send[*CreateProductResponseDto, *CreateProduct](ctx, command)
		if err != nil {
			tracing.TraceErr(span, err)
			return err
		}

		return nil
	}, append(retryOptions, retry.Context(ctx))...); err != nil {
		c.Log.WarnMsg("CreateProduct.Handle", err)
		tracing.TraceErr(span, err)
		c.CommitErrMessage(ctx, r, m)
		return
	}

	c.CommitMessage(ctx, r, m)
}
