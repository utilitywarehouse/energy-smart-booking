package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

var ErrSiteNotFound = errors.New("site not found")

type SiteStore struct {
	pool  *pgxpool.Pool
	batch *pgx.Batch
}

func NewSite(pool *pgxpool.Pool) *SiteStore {
	return &SiteStore{pool: pool}
}

func (s *SiteStore) Begin() {
	s.batch = &pgx.Batch{}
}

func (s *SiteStore) Commit(ctx context.Context) error {
	res := s.pool.SendBatch(ctx, s.batch)
	s.batch = nil
	return res.Close()
}

func (s *SiteStore) Upsert(site models.Site) {

	q := `
	INSERT INTO site (site_id,
		postcode,
		uprn,
		building_name_number,
		dependent_thoroughfare,
		thoroughfare,
		double_dependent_locality,
		dependent_locality,
		locality,
		county,
		town,
		department,
		organisation,
		po_box,
		delivery_point_suffix,
		sub_building_name_number)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	ON CONFLICT (site_id)
	DO UPDATE 
	SET postcode = $2,
		uprn = $3,
		building_name_number = $4,
		dependent_thoroughfare = $5,
		thoroughfare = $6,
		double_dependent_locality = $7,
		dependent_locality = $8,
		locality = $9,
		county = $10,
		town = $11,
		department = $12,
		organisation = $13,
		po_box = $14,
		delivery_point_suffix = $15,
		sub_building_name_number = $16,
		updated_at = now();
	`

	s.batch.Queue(q,
		site.SiteID,
		site.Postcode,
		site.UPRN,
		site.BuildingNameNumber,
		site.DependentThoroughfare,
		site.Thoroughfare,
		site.DoubleDependentLocality,
		site.DependentLocality,
		site.Locality,
		site.County,
		site.Town,
		site.Department,
		site.Organisation,
		site.PoBox,
		site.DeliveryPointSuffix,
		site.SubBuildingNameNumber)
}

func (s *SiteStore) GetSiteBySiteID(ctx context.Context, siteID string) (*models.Site, error) {
	var site models.Site
	q := `SELECT 
			site_id,
			postcode,
			uprn,
			building_name_number,
			dependent_thoroughfare,
			thoroughfare,
			double_dependent_locality,
			dependent_locality,
			locality,
			county,
			town,
			department,
			organisation,
			po_box,
			delivery_point_suffix,
			sub_building_name_number
	 FROM site 
	 WHERE
	 	site_id = $1;`
	if err := s.pool.QueryRow(ctx, q, siteID).
		Scan(&site.SiteID,
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
			&site.SubBuildingNameNumber); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSiteNotFound
		}
		return nil, err
	}

	return &site, nil
}

func (s *SiteStore) GetSiteByOccupancyID(ctx context.Context, occupancyID string) (*models.Site, error) {
	var site models.Site
	q := `
	SELECT 
		s.site_id,
		s.postcode,
		s.uprn,
		s.building_name_number,
		s.dependent_thoroughfare,
		s.thoroughfare,
		s.double_dependent_locality,
		s.dependent_locality,
		s.locality,
		s.county,
		s.town,
		s.department,
		s.organisation,
		s.po_box,
		s.delivery_point_suffix,
		s.sub_building_name_number
	FROM site AS s 
	JOIN occupancy AS o ON o.site_id = s.site_id
	WHERE o.occupancy_id = $1;`
	if err := s.pool.QueryRow(ctx, q, occupancyID).
		Scan(&site.SiteID,
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
			&site.SubBuildingNameNumber); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSiteNotFound
		}
		return nil, err
	}

	return &site, nil
}
