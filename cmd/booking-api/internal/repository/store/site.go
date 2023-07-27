package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

var ErrSiteNotFound = errors.New("site not found")

type SiteSerializer interface {
	SerializeSiteAddress(ctx context.Context, siteAddress models.SiteAddress) ([]byte, error)
	UnserializeSiteAddress(ctx context.Context, blob []byte) (models.SiteAddress, error)
}

type SiteStore struct {
	pool       *pgxpool.Pool
	batch      *pgx.Batch
	serializer SiteSerializer
}

func NewSite(pool *pgxpool.Pool, serializer SiteSerializer) *SiteStore {
	return &SiteStore{
		pool:       pool,
		serializer: serializer,
	}
}

func (s *SiteStore) Begin() {
	s.batch = &pgx.Batch{}
}

func (s *SiteStore) Commit(ctx context.Context) error {
	res := s.pool.SendBatch(ctx, s.batch)
	s.batch = nil
	return res.Close()
}

func (s *SiteStore) Upsert(ctx context.Context, site models.Site) error {

	q := `
	INSERT INTO site (
		site_id,
		address)
	VALUES ($1, $2)
	ON CONFLICT (site_id)
	DO UPDATE 
	SET address = $2,
		updated_at = now();
	`

	addressBlob, err := s.serializer.SerializeSiteAddress(ctx, site.SiteAddress)
	if err != nil {
		return fmt.Errorf("failed to serialize site address, %w", err)
	}

	s.batch.Queue(q, site.SiteID, addressBlob)

	return nil
}

func (s *SiteStore) GetSiteBySiteID(ctx context.Context, siteID string) (*models.Site, error) {
	var querySiteID string
	var blob []byte
	q := `SELECT 
			site_id,
			address
	 FROM site 
	 WHERE
	 	site_id = $1;`
	if err := s.pool.QueryRow(ctx, q, siteID).
		Scan(&querySiteID, &blob); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSiteNotFound
		}
		return nil, err
	}

	siteAddress, err := s.serializer.UnserializeSiteAddress(ctx, blob)
	if err != nil {
		return nil, fmt.Errorf("failed to unserialize site address, %w", err)
	}

	return &models.Site{
		SiteID:      querySiteID,
		SiteAddress: siteAddress,
	}, nil
}
