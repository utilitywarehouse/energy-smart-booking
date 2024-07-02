package indexer

import (
	"context"
	"net/http"

	"cloud.google.com/go/bigquery"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/googleapi"
)

type BigQuery interface {
	Begin()
	Commit(context.Context) error
	Queue(bigquery.ValueSaver)
}

type bq struct {
	inserter *bigquery.Inserter
	batch    *batch
	sizeHint int
}

func New(ctx context.Context, table *bigquery.Table, t interface{}, sizeHint int) BigQuery {
	schema := mustSchema(t)
	if err := ensureTable(ctx, table, schema); err != nil {
		logrus.WithError(err).Panicf("unable to create table %s", table.TableID)
	}

	return &bq{
		inserter: table.Inserter(),
		sizeHint: sizeHint,
	}
}

func (b *bq) Begin() {
	b.batch = newBatch(b.sizeHint)
}

func (b *bq) Queue(item bigquery.ValueSaver) {
	b.batch.queue(item)
}

func (b *bq) Commit(ctx context.Context) error {
	return b.inserter.Put(ctx, b.batch.items)
}

func ensureTable(ctx context.Context, table *bigquery.Table, schema bigquery.Schema) error {
	_, err := table.Metadata(ctx)
	if err != nil {
		gErr, ok := err.(*googleapi.Error)
		if !ok {
			return err
		}

		if gErr.Code != http.StatusNotFound {
			return err
		}

		if err = table.Create(ctx, &bigquery.TableMetadata{
			Name:   table.TableID,
			Schema: schema,
		}); err != nil {
			return err
		}
	}

	return nil
}

type batch struct {
	items []bigquery.ValueSaver
}

func newBatch(sizeHint int) *batch {
	return &batch{
		items: make([]bigquery.ValueSaver, 0, sizeHint),
	}
}

func (b *batch) queue(item bigquery.ValueSaver) {
	b.items = append(b.items, item)
}

func mustSchema(t interface{}) bigquery.Schema {
	schema, err := bigquery.InferSchema(t)
	if err != nil {
		logrus.WithError(err).Panicf("unable to parse schema for %+v", t)
	}
	return schema
}
