package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/K-H-Tech/apm/internal/models"
	"github.com/google/uuid"
)

var (
	// ErrPriorityScoreNotFound is returned when a priority score is not found
	ErrPriorityScoreNotFound = errors.New("priority score not found")
)

// PriorityRepository implements contract.PriorityRepository
type PriorityRepository struct {
	db *sql.DB
}

// NewPriorityRepository creates a new priority repository
func NewPriorityRepository(db *sql.DB) *PriorityRepository {
	return &PriorityRepository{db: db}
}

// Create creates a new priority score
func (r *PriorityRepository) Create(ctx context.Context, score *models.PriorityScore) error {
	query := `
		INSERT INTO priority_scores (
			id, prd_id, framework, reach, impact, confidence, effort,
			moscow_category, custom_criteria, final_score, calculated_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	now := time.Now()
	if score.ID == uuid.Nil {
		score.ID = uuid.New()
	}
	score.CreatedAt = now
	score.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		score.ID,
		score.PRDID,
		score.Framework,
		score.Reach,
		score.Impact,
		score.Confidence,
		score.Effort,
		score.MoSCoWCategory,
		score.CustomCriteria,
		score.FinalScore,
		score.CalculatedBy,
		score.CreatedAt,
		score.UpdatedAt,
	)

	return err
}

// GetByID retrieves a priority score by ID
func (r *PriorityRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PriorityScore, error) {
	query := `
		SELECT id, prd_id, framework, reach, impact, confidence, effort,
			   moscow_category, custom_criteria, final_score, calculated_by, created_at, updated_at
		FROM priority_scores
		WHERE id = $1
	`

	var score models.PriorityScore
	var reach, impact, confidence, effort sql.NullInt64
	var moscowCategory, customCriteria sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&score.ID,
		&score.PRDID,
		&score.Framework,
		&reach,
		&impact,
		&confidence,
		&effort,
		&moscowCategory,
		&customCriteria,
		&score.FinalScore,
		&score.CalculatedBy,
		&score.CreatedAt,
		&score.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPriorityScoreNotFound
	}
	if err != nil {
		return nil, err
	}

	if reach.Valid {
		r := int(reach.Int64)
		score.Reach = &r
	}
	if impact.Valid {
		i := int(impact.Int64)
		score.Impact = &i
	}
	if confidence.Valid {
		c := int(confidence.Int64)
		score.Confidence = &c
	}
	if effort.Valid {
		e := int(effort.Int64)
		score.Effort = &e
	}
	if moscowCategory.Valid {
		cat := models.MoSCoWCategory(moscowCategory.String)
		score.MoSCoWCategory = &cat
	}
	score.CustomCriteria = customCriteria.String

	return &score, nil
}

// GetByPRDID retrieves all priority scores for a PRD
func (r *PriorityRepository) GetByPRDID(ctx context.Context, prdID uuid.UUID) ([]*models.PriorityScore, error) {
	query := `
		SELECT id, prd_id, framework, reach, impact, confidence, effort,
			   moscow_category, custom_criteria, final_score, calculated_by, created_at, updated_at
		FROM priority_scores
		WHERE prd_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, prdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := make([]*models.PriorityScore, 0)
	for rows.Next() {
		var score models.PriorityScore
		var reach, impact, confidence, effort sql.NullInt64
		var moscowCategory, customCriteria sql.NullString

		if err := rows.Scan(
			&score.ID,
			&score.PRDID,
			&score.Framework,
			&reach,
			&impact,
			&confidence,
			&effort,
			&moscowCategory,
			&customCriteria,
			&score.FinalScore,
			&score.CalculatedBy,
			&score.CreatedAt,
			&score.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if reach.Valid {
			r := int(reach.Int64)
			score.Reach = &r
		}
		if impact.Valid {
			i := int(impact.Int64)
			score.Impact = &i
		}
		if confidence.Valid {
			c := int(confidence.Int64)
			score.Confidence = &c
		}
		if effort.Valid {
			e := int(effort.Int64)
			score.Effort = &e
		}
		if moscowCategory.Valid {
			cat := models.MoSCoWCategory(moscowCategory.String)
			score.MoSCoWCategory = &cat
		}
		score.CustomCriteria = customCriteria.String

		scores = append(scores, &score)
	}

	return scores, rows.Err()
}

// GetByPRDIDAndFramework retrieves a specific framework's score for a PRD
func (r *PriorityRepository) GetByPRDIDAndFramework(ctx context.Context, prdID uuid.UUID, framework models.PriorityFramework) (*models.PriorityScore, error) {
	query := `
		SELECT id, prd_id, framework, reach, impact, confidence, effort,
			   moscow_category, custom_criteria, final_score, calculated_by, created_at, updated_at
		FROM priority_scores
		WHERE prd_id = $1 AND framework = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var score models.PriorityScore
	var reach, impact, confidence, effort sql.NullInt64
	var moscowCategory, customCriteria sql.NullString

	err := r.db.QueryRowContext(ctx, query, prdID, framework).Scan(
		&score.ID,
		&score.PRDID,
		&score.Framework,
		&reach,
		&impact,
		&confidence,
		&effort,
		&moscowCategory,
		&customCriteria,
		&score.FinalScore,
		&score.CalculatedBy,
		&score.CreatedAt,
		&score.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPriorityScoreNotFound
	}
	if err != nil {
		return nil, err
	}

	if reach.Valid {
		r := int(reach.Int64)
		score.Reach = &r
	}
	if impact.Valid {
		i := int(impact.Int64)
		score.Impact = &i
	}
	if confidence.Valid {
		c := int(confidence.Int64)
		score.Confidence = &c
	}
	if effort.Valid {
		e := int(effort.Int64)
		score.Effort = &e
	}
	if moscowCategory.Valid {
		cat := models.MoSCoWCategory(moscowCategory.String)
		score.MoSCoWCategory = &cat
	}
	score.CustomCriteria = customCriteria.String

	return &score, nil
}

// Update updates an existing priority score
func (r *PriorityRepository) Update(ctx context.Context, score *models.PriorityScore) error {
	query := `
		UPDATE priority_scores SET
			reach = $2,
			impact = $3,
			confidence = $4,
			effort = $5,
			moscow_category = $6,
			custom_criteria = $7,
			final_score = $8,
			updated_at = $9
		WHERE id = $1
	`

	score.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		score.ID,
		score.Reach,
		score.Impact,
		score.Confidence,
		score.Effort,
		score.MoSCoWCategory,
		score.CustomCriteria,
		score.FinalScore,
		score.UpdatedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPriorityScoreNotFound
	}

	return nil
}

// Delete deletes a priority score by ID
func (r *PriorityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM priority_scores WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPriorityScoreNotFound
	}

	return nil
}

// DeleteByPRDID deletes all priority scores for a PRD
func (r *PriorityRepository) DeleteByPRDID(ctx context.Context, prdID uuid.UUID) error {
	query := `DELETE FROM priority_scores WHERE prd_id = $1`
	_, err := r.db.ExecContext(ctx, query, prdID)
	return err
}

// GetRankings retrieves PRDs ranked by score for a specific framework
func (r *PriorityRepository) GetRankings(ctx context.Context, orgID uuid.UUID, framework models.PriorityFramework, limit, offset int) ([]PRDRanking, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(DISTINCT ps.prd_id)
		FROM priority_scores ps
		JOIN prds p ON p.id = ps.prd_id
		WHERE p.organization_id = $1 AND ps.framework = $2
	`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, orgID, framework).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get rankings with latest score per PRD
	query := `
		WITH latest_scores AS (
			SELECT DISTINCT ON (ps.prd_id)
				ps.id, ps.prd_id, ps.framework, ps.final_score, ps.created_at,
				p.title
			FROM priority_scores ps
			JOIN prds p ON p.id = ps.prd_id
			WHERE p.organization_id = $1 AND ps.framework = $2
			ORDER BY ps.prd_id, ps.created_at DESC
		)
		SELECT id, prd_id, framework, final_score, created_at, title
		FROM latest_scores
		ORDER BY final_score DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, framework, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	rankings := make([]PRDRanking, 0)
	rank := offset + 1
	for rows.Next() {
		var ranking PRDRanking
		if err := rows.Scan(
			&ranking.ScoreID,
			&ranking.PRDID,
			&ranking.Framework,
			&ranking.Score,
			&ranking.ScoredAt,
			&ranking.PRDTitle,
		); err != nil {
			return nil, 0, err
		}
		ranking.Rank = rank
		rank++
		rankings = append(rankings, ranking)
	}

	return rankings, total, rows.Err()
}

// PRDRanking represents a PRD's ranking by priority score
type PRDRanking struct {
	Rank      int                     `json:"rank"`
	PRDID     uuid.UUID               `json:"prd_id"`
	PRDTitle  string                  `json:"prd_title"`
	ScoreID   uuid.UUID               `json:"score_id"`
	Framework models.PriorityFramework `json:"framework"`
	Score     float64                 `json:"score"`
	ScoredAt  time.Time               `json:"scored_at"`
}

// GetMoSCoWDistribution retrieves the MoSCoW category distribution for an organization
func (r *PriorityRepository) GetMoSCoWDistribution(ctx context.Context, orgID uuid.UUID) (*MoSCoWDistribution, error) {
	query := `
		SELECT
			moscow_category,
			COUNT(*) as count
		FROM priority_scores ps
		JOIN prds p ON p.id = ps.prd_id
		WHERE p.organization_id = $1
		AND ps.framework = 'moscow'
		AND ps.moscow_category IS NOT NULL
		GROUP BY moscow_category
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dist := &MoSCoWDistribution{}
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, err
		}

		switch models.MoSCoWCategory(category) {
		case models.MoSCoWMust:
			dist.MustCount = count
		case models.MoSCoWShould:
			dist.ShouldCount = count
		case models.MoSCoWCould:
			dist.CouldCount = count
		case models.MoSCoWWont:
			dist.WontCount = count
		}
	}

	dist.Total = dist.MustCount + dist.ShouldCount + dist.CouldCount + dist.WontCount
	if dist.Total > 0 {
		dist.MustPercent = float64(dist.MustCount) / float64(dist.Total) * 100
		dist.ShouldPercent = float64(dist.ShouldCount) / float64(dist.Total) * 100
		dist.CouldPercent = float64(dist.CouldCount) / float64(dist.Total) * 100
		dist.WontPercent = float64(dist.WontCount) / float64(dist.Total) * 100
	}

	return dist, rows.Err()
}

// MoSCoWDistribution represents the distribution of MoSCoW categories
type MoSCoWDistribution struct {
	MustCount     int     `json:"must_count"`
	ShouldCount   int     `json:"should_count"`
	CouldCount    int     `json:"could_count"`
	WontCount     int     `json:"wont_count"`
	Total         int     `json:"total"`
	MustPercent   float64 `json:"must_percent"`
	ShouldPercent float64 `json:"should_percent"`
	CouldPercent  float64 `json:"could_percent"`
	WontPercent   float64 `json:"wont_percent"`
}

// GetScoreHistory retrieves the score history for a PRD
func (r *PriorityRepository) GetScoreHistory(ctx context.Context, prdID uuid.UUID, framework models.PriorityFramework) ([]*models.PriorityScore, error) {
	query := `
		SELECT id, prd_id, framework, reach, impact, confidence, effort,
			   moscow_category, custom_criteria, final_score, calculated_by, created_at, updated_at
		FROM priority_scores
		WHERE prd_id = $1 AND framework = $2
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, prdID, framework)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := make([]*models.PriorityScore, 0)
	for rows.Next() {
		var score models.PriorityScore
		var reach, impact, confidence, effort sql.NullInt64
		var moscowCategory, customCriteria sql.NullString

		if err := rows.Scan(
			&score.ID,
			&score.PRDID,
			&score.Framework,
			&reach,
			&impact,
			&confidence,
			&effort,
			&moscowCategory,
			&customCriteria,
			&score.FinalScore,
			&score.CalculatedBy,
			&score.CreatedAt,
			&score.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if reach.Valid {
			r := int(reach.Int64)
			score.Reach = &r
		}
		if impact.Valid {
			i := int(impact.Int64)
			score.Impact = &i
		}
		if confidence.Valid {
			c := int(confidence.Int64)
			score.Confidence = &c
		}
		if effort.Valid {
			e := int(effort.Int64)
			score.Effort = &e
		}
		if moscowCategory.Valid {
			cat := models.MoSCoWCategory(moscowCategory.String)
			score.MoSCoWCategory = &cat
		}
		score.CustomCriteria = customCriteria.String

		scores = append(scores, &score)
	}

	return scores, rows.Err()
}
