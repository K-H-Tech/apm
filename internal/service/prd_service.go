package service

import (
	"context"
	"errors"

	"github.com/K-H-Tech/apm/internal/models"
	"github.com/K-H-Tech/apm/internal/repository"
	"github.com/google/uuid"
)

var (
	// ErrPRDNotFound is returned when a PRD is not found
	ErrPRDNotFound = errors.New("PRD not found")

	// ErrUnauthorized is returned when a user is not authorized
	ErrUnauthorized = errors.New("unauthorized to perform this action")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
)

// PRDAIService defines AI operations needed by PRDService
type PRDAIService interface {
	GeneratePRD(ctx context.Context, input GeneratePRDInput) (*GeneratedPRDResult, error)
	RefinePRD(ctx context.Context, content *models.PRDContent, feedback string) (*models.PRDContent, error)
	GenerateUserStories(ctx context.Context, content *models.PRDContent) ([]models.UserStory, error)
	GenerateAcceptanceCriteria(ctx context.Context, story *models.UserStory) ([]string, error)
}

// GeneratedPRDResult contains the generated PRD with its title
type GeneratedPRDResult struct {
	Title   string
	Content *models.PRDContent
}

// PRDService implements contract.PRDService
type PRDService struct {
	prdRepo      *repository.PRDRepository
	templateRepo *repository.TemplateRepository
	aiService    PRDAIService
}

// NewPRDService creates a new PRD service
func NewPRDService(
	prdRepo *repository.PRDRepository,
	templateRepo *repository.TemplateRepository,
	aiService PRDAIService,
) *PRDService {
	return &PRDService{
		prdRepo:      prdRepo,
		templateRepo: templateRepo,
		aiService:    aiService,
	}
}

// CreatePRDInput holds input for creating a PRD
type CreatePRDInput struct {
	OrganizationID uuid.UUID
	OwnerID        uuid.UUID
	Title          string
	TemplateID     *uuid.UUID
	Content        *models.PRDContent
}

// Create creates a new PRD
func (s *PRDService) Create(ctx context.Context, input CreatePRDInput) (*models.PRD, error) {
	if input.Title == "" {
		return nil, ErrInvalidInput
	}

	prd := &models.PRD{
		ID:             uuid.New(),
		OrganizationID: input.OrganizationID,
		OwnerID:        input.OwnerID,
		Title:          input.Title,
		Status:         models.PRDStatusDraft,
		TemplateID:     input.TemplateID,
	}

	// Use provided content or initialize empty
	if input.Content != nil {
		prd.Content = *input.Content
	}

	if err := s.prdRepo.Create(ctx, prd); err != nil {
		return nil, err
	}

	return prd, nil
}

// GetByID retrieves a PRD by ID
func (s *PRDService) GetByID(ctx context.Context, id uuid.UUID) (*models.PRD, error) {
	return s.prdRepo.GetByID(ctx, id)
}

// Update updates an existing PRD
func (s *PRDService) Update(ctx context.Context, prd *models.PRD) error {
	existing, err := s.prdRepo.GetByID(ctx, prd.ID)
	if err != nil {
		return err
	}

	// Preserve immutable fields
	prd.OrganizationID = existing.OrganizationID
	prd.OwnerID = existing.OwnerID
	prd.CreatedAt = existing.CreatedAt

	return s.prdRepo.Update(ctx, prd)
}

// Delete deletes a PRD
func (s *PRDService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.prdRepo.Delete(ctx, id)
}

// ListByOrganization lists PRDs for an organization
func (s *PRDService) ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.PRD, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.prdRepo.ListByOrganization(ctx, orgID, limit, offset)
}

// ListByOwner lists PRDs owned by a user
func (s *PRDService) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*models.PRD, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.prdRepo.ListByOwner(ctx, ownerID, limit, offset)
}

// ListByStatus lists PRDs by status
func (s *PRDService) ListByStatus(ctx context.Context, orgID uuid.UUID, status models.PRDStatus, limit, offset int) ([]*models.PRD, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.prdRepo.ListByStatus(ctx, orgID, status, limit, offset)
}

// Search searches PRDs by title or content
func (s *PRDService) Search(ctx context.Context, orgID uuid.UUID, query string, limit, offset int) ([]*models.PRD, int, error) {
	if query == "" {
		return nil, 0, ErrInvalidInput
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.prdRepo.Search(ctx, orgID, query, limit, offset)
}

// UpdateStatus updates the status of a PRD
func (s *PRDService) UpdateStatus(ctx context.Context, id uuid.UUID, status models.PRDStatus) error {
	prd, err := s.prdRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	prd.Status = status
	return s.prdRepo.Update(ctx, prd)
}

// CreateVersion creates a new version of a PRD
func (s *PRDService) CreateVersion(ctx context.Context, prdID uuid.UUID, userID uuid.UUID, summary string) (*models.PRDVersion, error) {
	prd, err := s.prdRepo.GetByID(ctx, prdID)
	if err != nil {
		return nil, err
	}

	// Get latest version number
	latestVersion, err := s.prdRepo.GetLatestVersionNumber(ctx, prdID)
	if err != nil {
		return nil, err
	}

	userIDPtr := &userID
	version := &models.PRDVersion{
		ID:            uuid.New(),
		PRDID:         prdID,
		VersionNumber: latestVersion + 1,
		Content:       prd.Content,
		ChangedBy:     userIDPtr,
		ChangeSummary: summary,
	}

	if err := s.prdRepo.CreateVersion(ctx, version); err != nil {
		return nil, err
	}

	return version, nil
}

// GetVersion retrieves a specific version of a PRD
func (s *PRDService) GetVersion(ctx context.Context, prdID uuid.UUID, versionNumber int) (*models.PRDVersion, error) {
	return s.prdRepo.GetVersion(ctx, prdID, versionNumber)
}

// ListVersions lists all versions of a PRD
func (s *PRDService) ListVersions(ctx context.Context, prdID uuid.UUID) ([]*models.PRDVersion, error) {
	return s.prdRepo.ListVersions(ctx, prdID)
}

// RestoreVersion restores a PRD to a specific version
func (s *PRDService) RestoreVersion(ctx context.Context, prdID uuid.UUID, versionNumber int, userID uuid.UUID) error {
	version, err := s.prdRepo.GetVersion(ctx, prdID, versionNumber)
	if err != nil {
		return err
	}

	prd, err := s.prdRepo.GetByID(ctx, prdID)
	if err != nil {
		return err
	}

	// Create a new version with current state before restoring
	_, err = s.CreateVersion(ctx, prdID, userID, "Auto-saved before restore")
	if err != nil {
		return err
	}

	// Restore content from the specified version
	prd.Content = version.Content
	return s.prdRepo.Update(ctx, prd)
}

// GenerateFromNotesInput holds input for AI generation
type GenerateFromNotesInput struct {
	OrganizationID uuid.UUID
	OwnerID        uuid.UUID
	Notes          string
	MeetingType    string
	TemplateID     *uuid.UUID
}

// GenerateFromNotes generates a PRD from meeting notes using AI
func (s *PRDService) GenerateFromNotes(ctx context.Context, input GenerateFromNotesInput) (*models.PRD, error) {
	if input.Notes == "" {
		return nil, ErrInvalidInput
	}

	if s.aiService == nil {
		return nil, errors.New("AI service not configured")
	}

	// Get template if specified
	var template *models.PRDTemplate
	if input.TemplateID != nil {
		t, err := s.templateRepo.GetByID(ctx, *input.TemplateID)
		if err != nil {
			return nil, err
		}
		template = t
	}

	// Generate PRD content using AI
	result, err := s.aiService.GeneratePRD(ctx, GeneratePRDInput{
		Notes:       input.Notes,
		MeetingType: input.MeetingType,
		Template:    template,
	})
	if err != nil {
		return nil, err
	}

	// Create the PRD
	return s.Create(ctx, CreatePRDInput{
		OrganizationID: input.OrganizationID,
		OwnerID:        input.OwnerID,
		Title:          result.Title,
		TemplateID:     input.TemplateID,
		Content:        result.Content,
	})
}

// RefinePRD refines an existing PRD using AI
func (s *PRDService) RefinePRD(ctx context.Context, prdID uuid.UUID, feedback string) (*models.PRD, error) {
	prd, err := s.prdRepo.GetByID(ctx, prdID)
	if err != nil {
		return nil, err
	}

	if s.aiService == nil {
		return nil, errors.New("AI service not configured")
	}

	// Refine content using AI
	refinedContent, err := s.aiService.RefinePRD(ctx, &prd.Content, feedback)
	if err != nil {
		return nil, err
	}

	prd.Content = *refinedContent
	if err := s.prdRepo.Update(ctx, prd); err != nil {
		return nil, err
	}

	return prd, nil
}

// GenerateUserStories generates user stories for a PRD using AI
func (s *PRDService) GenerateUserStories(ctx context.Context, prdID uuid.UUID) ([]models.UserStory, error) {
	prd, err := s.prdRepo.GetByID(ctx, prdID)
	if err != nil {
		return nil, err
	}

	if s.aiService == nil {
		return nil, errors.New("AI service not configured")
	}

	stories, err := s.aiService.GenerateUserStories(ctx, &prd.Content)
	if err != nil {
		return nil, err
	}

	// Update PRD with generated stories
	prd.Content.UserStories = stories
	if err := s.prdRepo.Update(ctx, prd); err != nil {
		return nil, err
	}

	return stories, nil
}

// GenerateAcceptanceCriteria generates acceptance criteria for a user story using AI
func (s *PRDService) GenerateAcceptanceCriteria(ctx context.Context, story *models.UserStory) ([]string, error) {
	if s.aiService == nil {
		return nil, errors.New("AI service not configured")
	}

	return s.aiService.GenerateAcceptanceCriteria(ctx, story)
}

// LinkToJira links a PRD to a Jira epic
func (s *PRDService) LinkToJira(ctx context.Context, prdID uuid.UUID, epicKey string) error {
	prd, err := s.prdRepo.GetByID(ctx, prdID)
	if err != nil {
		return err
	}

	prd.JiraEpicKey = epicKey
	return s.prdRepo.Update(ctx, prd)
}

// LinkToConfluence links a PRD to a Confluence page
func (s *PRDService) LinkToConfluence(ctx context.Context, prdID uuid.UUID, pageID string) error {
	prd, err := s.prdRepo.GetByID(ctx, prdID)
	if err != nil {
		return err
	}

	prd.ConfluencePageID = pageID
	return s.prdRepo.Update(ctx, prd)
}
