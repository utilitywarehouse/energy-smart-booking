package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
)

var ErrOccupancyNotFound = errors.New("occupancy not found")

type OccupancyStore struct {
	pool *pgxpool.Pool
}

type Occupancy struct {
	ID        string
	SiteID    string
	AccountID string
}

func NewOccupancy(pool *pgxpool.Pool) *OccupancyStore {
	return &OccupancyStore{pool: pool}
}

func (s *OccupancyStore) Add(ctx context.Context, id, siteID, accountID string, at time.Time) error {
	q := `
	INSERT INTO occupancies (id, site_id, account_id, created_at)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id)
	DO NOTHING;`
	_, err := s.pool.Exec(ctx, q, id, siteID, accountID, at)

	return err
}

func (s *OccupancyStore) AddSite(ctx context.Context, occupancyID, siteID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE occupancies SET site_id = $2 WHERE id = $1;`, occupancyID, siteID)

	return err
}

func (s *OccupancyStore) Get(ctx context.Context, id string) (Occupancy, error) {
	var occ Occupancy
	if err := s.pool.QueryRow(ctx, `SELECT id, site_id, account_id FROM occupancies WHERE id = $1`, id).
		Scan(&occ.ID, &occ.SiteID, &occ.AccountID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Occupancy{}, ErrOccupancyNotFound
		}
		return Occupancy{}, err
	}

	return occ, nil
}

func (s *OccupancyStore) GetIDsByAccount(ctx context.Context, accountID string) ([]string, error) {
	q := `SELECT id FROM occupancies WHERE account_id = $1 ORDER BY created_at DESC;`

	return s.queryOccupanciesByIdentifier(ctx, q, accountID)
}

func (s *OccupancyStore) GetIDsBySite(ctx context.Context, siteID string) ([]string, error) {
	q := `SELECT id FROM occupancies WHERE site_id = $1;`

	return s.queryOccupanciesByIdentifier(ctx, q, siteID)
}

// GetIDsByPostcode gets occupancies by postcode.
func (s *OccupancyStore) GetIDsByPostcode(ctx context.Context, postCode string) ([]string, error) {
	q := `SELECT id FROM occupancies WHERE site_id IN (SELECT id FROM sites WHERE post_code = $1);`

	return s.queryOccupanciesByIdentifier(ctx, q, postCode)
}

func (s *OccupancyStore) GetIDsByMPXN(ctx context.Context, mpxn string) ([]string, error) {
	q := `SELECT occupancy_id FROM services WHERE mpxn = $1 AND is_live IS TRUE;`

	return s.queryOccupanciesByIdentifier(ctx, q, mpxn)
}

func (s *OccupancyStore) GetLiveOccupancies(ctx context.Context) ([]string, error) {
	var ids = make([]string, 0)

	q := `SELECT distinct(occupancy_id) FROM services WHERE is_live IS TRUE limit 2500;`

	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var occID string
		err = rows.Scan(&occID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, occID)
	}

	return ids, rows.Err()
}

func (s *OccupancyStore) LoadOccupancy(ctx context.Context, occupancyID string) (domain.Occupancy, error) {

	q := `
	SELECT o.account_id,
	       e.reasons, sup.reasons, c.reasons,
	       a.id, a.psr_codes, a.opt_out,
	       s.id, s.post_code,
	       p.post_code, p.wan_coverage
	FROM occupancies o 
	LEFT JOIN eligibility e ON o.id = e.occupancy_id
	LEFT JOIN suppliability sup ON o.id = sup.occupancy_id
	LEFT JOIN campaignability c ON o.id = c.occupancy_id
	LEFT JOIN accounts a ON o.account_id = a.id
	LEFT JOIN sites s on o.site_id = s.id 
	LEFT JOIN postcodes p on s.post_code = p.post_code
	WHERE o.id = $1;`

	rows, err := s.pool.Query(ctx, q, occupancyID)
	if err != nil {
		return domain.Occupancy{}, err
	}
	defer rows.Close()

	occupancy := domain.Occupancy{
		ID: occupancyID,
	}

	for rows.Next() {
		var (
			occupancyAccountID                        string
			accountID, siteID, sitePostCode, postCode sql.NullString
			psrCodes                                  []string
			wanCoverage, optOut                       sql.NullBool
			evaluationResult                          domain.OccupancyEvaluation
		)

		err = rows.Scan(
			&occupancyAccountID,
			&evaluationResult.Eligibility,
			&evaluationResult.Suppliability,
			&evaluationResult.Campaignability,
			&accountID,
			&psrCodes,
			&optOut,
			&siteID,
			&sitePostCode,
			&postCode,
			&wanCoverage,
		)
		if err != nil {
			return domain.Occupancy{}, err
		}

		evaluationResult.OccupancyID = occupancyID
		occupancy.EvaluationResult = evaluationResult

		if postCode.Valid {
			occupancy.Site = &domain.Site{
				ID:          siteID.String,
				Postcode:    sitePostCode.String,
				WanCoverage: wanCoverage.Bool,
			}
		}
		occupancy.Account = domain.Account{
			ID: occupancyAccountID,
		}
		if accountID.Valid {
			occupancy.Account.PSRCodes = psrCodes
			if optOut.Valid {
				occupancy.Account.OptOut = optOut.Bool
			}
		}
	}

	if err = rows.Err(); err != nil {
		return domain.Occupancy{}, rows.Err()
	}

	return occupancy, nil
}

func (s *OccupancyStore) queryOccupanciesByIdentifier(ctx context.Context, query, id string) ([]string, error) {
	var ids = make([]string, 0)

	rows, err := s.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var occID string
		err = rows.Scan(&occID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, occID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}
