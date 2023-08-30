package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

var (
	ErrOccupancyNotFound = errors.New("occupancy not found")
	ErrFailedUpdate      = errors.New("no rows were affected by update statement")

	ErrNoOccupancyEligibleFound = errors.New("eligible occupancy not found")
)

type OccupancyStore struct {
	pool  *pgxpool.Pool
	batch *pgx.Batch
}

func NewOccupancy(pool *pgxpool.Pool) *OccupancyStore {
	return &OccupancyStore{pool: pool}
}

func (s *OccupancyStore) Begin() {
	s.batch = &pgx.Batch{}
}

func (s *OccupancyStore) Commit(ctx context.Context) error {
	res := s.pool.SendBatch(ctx, s.batch)

	s.batch = nil
	return res.Close()
}

func (s *OccupancyStore) Insert(occupancy models.Occupancy) {
	q := `
	INSERT INTO occupancy (occupancy_id, site_id, account_id, created_at)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (occupancy_id)
	DO NOTHING;`

	s.batch.Queue(q, occupancy.OccupancyID, occupancy.SiteID, occupancy.AccountID, occupancy.CreatedAt)
}

func (s *OccupancyStore) UpdateSiteID(occupancyID, siteID string) {
	q := `UPDATE occupancy SET site_id = $2 WHERE occupancy_id = $1;`

	s.batch.Queue(q, occupancyID, siteID)
}

func (s *OccupancyStore) GetOccupancyByID(ctx context.Context, occupancyID string) (*models.Occupancy, error) {
	var occ models.Occupancy
	q := `SELECT occupancy_id, site_id, account_id FROM occupancy WHERE occupancy_id = $1;`
	if err := s.pool.QueryRow(ctx, q, occupancyID).
		Scan(&occ.OccupancyID, &occ.SiteID, &occ.AccountID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOccupancyNotFound
		}
		return nil, err
	}

	return &occ, nil
}

func (s *OccupancyStore) GetLiveOccupanciesByAccountID(ctx context.Context, accountID string) ([]models.Occupancy, error) {
	occupancies := make([]models.Occupancy, 0)

	q := `SELECT 
		o.occupancy_id,
		o.site_id,
		o.account_id,
		o.created_at
	FROM 
		occupancy o
	INNER JOIN service s 
		ON o.occupancy_id = s.occupancy_id
	WHERE
		o.account_id = $1
	AND
		s.is_live IS TRUE
	ORDER BY
		o.created_at DESC;`

	rows, err := s.pool.Query(ctx, q, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query get occupancies by account id, %w", err)
	}

	for rows.Next() {
		var occ models.Occupancy

		err := rows.Scan(&occ.OccupancyID, &occ.SiteID, &occ.AccountID, &occ.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row, %w", err)

		}

		occupancies = append(occupancies, occ)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error found in rows, %w", err)
	}

	return occupancies, nil
}

func (s *OccupancyStore) GetSiteExternalReferenceByAccountID(ctx context.Context, accountID string) (*models.Site, *models.OccupancyEligibility, error) {
	var site models.Site
	var occupancyEligibility models.OccupancyEligibility

	q := `
	SELECT
		si.site_id,
		si.postcode,
		si.uprn,
		si.building_name_number,
		si.dependent_thoroughfare,
		si.thoroughfare,
		si.double_dependent_locality,
		si.dependent_locality,
		si.locality,
		si.county,
		si.town,
		si.department,
		si.organisation,
		si.po_box,
		si.delivery_point_suffix,
		si.sub_building_name_number,
		oe.occupancy_id,
		oe.reference
	
		FROM occupancy_eligible oe
		JOIN service s ON oe.occupancy_id = s.occupancy_id
		JOIN occupancy o ON o.occupancy_id = oe.occupancy_id
		JOIN site si ON si.site_id = o.site_id
	
		WHERE o.account_id = $1
		AND s.is_live IS TRUE
		ORDER BY
			o.created_at DESC;`

	err := s.pool.QueryRow(ctx, q, accountID).Scan(&site.SiteID,
		&site.Postcode,
		&site.UPRN,
		&site.BuildingNameNumber,
		&site.DependentThoroughfare,
		&site.Thoroughfare,
		&site.DoubleDependentLocality,
		&site.DependentLocality,
		&site.Locality,
		&site.County,
		&site.Town,
		&site.Department,
		&site.Organisation,
		&site.PoBox,
		&site.DeliveryPointSuffix,
		&site.SubBuildingNameNumber,
		&occupancyEligibility.OccupancyID,
		&occupancyEligibility.Reference,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrNoOccupancyEligibleFound
		}
	}

	return &site, &occupancyEligibility, nil
}
