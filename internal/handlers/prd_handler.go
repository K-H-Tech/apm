package handlers

import (
	"net/http"
	"strconv"

	"github.com/K-H-Tech/apm/internal/models"
	"github.com/K-H-Tech/apm/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PRDHandler handles PRD-related HTTP requests
type PRDHandler struct {
	prdService *service.PRDService
}

// NewPRDHandler creates a new PRD handler
func NewPRDHandler(prdService *service.PRDService) *PRDHandler {
	return &PRDHandler{
		prdService: prdService,
	}
}

// RegisterRoutes registers PRD routes
func (h *PRDHandler) RegisterRoutes(r *gin.RouterGroup) {
	prds := r.Group("/prds")
	{
		prds.POST("", h.Create)
		prds.GET("", h.List)
		prds.GET("/:id", h.GetByID)
		prds.PUT("/:id", h.Update)
		prds.DELETE("/:id", h.Delete)
		prds.PATCH("/:id/status", h.UpdateStatus)

		// Versions
		prds.GET("/:id/versions", h.ListVersions)
		prds.GET("/:id/versions/:version", h.GetVersion)
		prds.POST("/:id/versions", h.CreateVersion)
		prds.POST("/:id/versions/:version/restore", h.RestoreVersion)

		// AI Operations
		prds.POST("/generate", h.GenerateFromNotes)
		prds.POST("/:id/refine", h.Refine)
		prds.POST("/:id/stories", h.GenerateUserStories)

		// Integrations
		prds.POST("/:id/link/jira", h.LinkToJira)
		prds.POST("/:id/link/confluence", h.LinkToConfluence)
	}
}

// CreatePRDRequest represents the request body for creating a PRD
type CreatePRDRequest struct {
	Title      string              `json:"title" binding:"required"`
	TemplateID *string             `json:"template_id,omitempty"`
	Content    *models.PRDContent  `json:"content,omitempty"`
}

// Create creates a new PRD
// @Summary Create a new PRD
// @Tags PRDs
// @Accept json
// @Produce json
// @Param request body CreatePRDRequest true "PRD to create"
// @Success 201 {object} models.PRD
// @Router /api/v1/prds [post]
func (h *PRDHandler) Create(c *gin.Context) {
	var req CreatePRDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user and organization from context (set by auth middleware)
	userID := getUserID(c)
	orgID := getOrganizationID(c)

	// Validate authentication - don't proceed with nil UUIDs
	if userID == uuid.Nil || orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var templateID *uuid.UUID
	if req.TemplateID != nil {
		tid, err := uuid.Parse(*req.TemplateID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template_id"})
			return
		}
		templateID = &tid
	}

	prd, err := h.prdService.Create(c.Request.Context(), service.CreatePRDInput{
		OrganizationID: orgID,
		OwnerID:        userID,
		Title:          req.Title,
		TemplateID:     templateID,
		Content:        req.Content,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, prd)
}

// GetByID retrieves a PRD by ID
// @Summary Get a PRD by ID
// @Tags PRDs
// @Produce json
// @Param id path string true "PRD ID"
// @Success 200 {object} models.PRD
// @Router /api/v1/prds/{id} [get]
func (h *PRDHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Authorization check
	orgID := getOrganizationID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	prd, err := h.prdService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRD not found"})
		return
	}

	// Verify PRD belongs to user's organization
	if prd.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, prd)
}

// ListPRDsResponse represents the response for listing PRDs
type ListPRDsResponse struct {
	PRDs   []*models.PRD `json:"prds"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// List lists PRDs with pagination
// @Summary List PRDs
// @Tags PRDs
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param status query string false "Filter by status"
// @Param owner query string false "Filter by owner ID"
// @Param q query string false "Search query"
// @Success 200 {object} ListPRDsResponse
// @Router /api/v1/prds [get]
func (h *PRDHandler) List(c *gin.Context) {
	orgID := getOrganizationID(c)

	// Validate authentication
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")
	owner := c.Query("owner")
	query := c.Query("q")

	var prds []*models.PRD
	var total int
	var err error

	ctx := c.Request.Context()

	switch {
	case query != "":
		prds, total, err = h.prdService.Search(ctx, orgID, query, limit, offset)
	case status != "":
		prds, total, err = h.prdService.ListByStatus(ctx, orgID, models.PRDStatus(status), limit, offset)
	case owner != "":
		ownerID, parseErr := uuid.Parse(owner)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner id"})
			return
		}
		prds, total, err = h.prdService.ListByOwner(ctx, ownerID, limit, offset)
	default:
		prds, total, err = h.prdService.ListByOrganization(ctx, orgID, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListPRDsResponse{
		PRDs:   prds,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// UpdatePRDRequest represents the request body for updating a PRD
type UpdatePRDRequest struct {
	Title   string             `json:"title,omitempty"`
	Content *models.PRDContent `json:"content,omitempty"`
}

// Update updates a PRD
// @Summary Update a PRD
// @Tags PRDs
// @Accept json
// @Produce json
// @Param id path string true "PRD ID"
// @Param request body UpdatePRDRequest true "Updates"
// @Success 200 {object} models.PRD
// @Router /api/v1/prds/{id} [put]
func (h *PRDHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Authorization check
	orgID := getOrganizationID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req UpdatePRDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	prd, err := h.prdService.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRD not found"})
		return
	}

	// Verify PRD belongs to user's organization
	if prd.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if req.Title != "" {
		prd.Title = req.Title
	}
	if req.Content != nil {
		prd.Content = *req.Content
	}

	if err := h.prdService.Update(ctx, prd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prd)
}

// Delete deletes a PRD
// @Summary Delete a PRD
// @Tags PRDs
// @Param id path string true "PRD ID"
// @Success 204
// @Router /api/v1/prds/{id} [delete]
func (h *PRDHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Authorization check
	orgID := getOrganizationID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	ctx := c.Request.Context()

	// Verify PRD belongs to user's organization before deleting
	prd, err := h.prdService.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRD not found"})
		return
	}

	if prd.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.prdService.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateStatusRequest represents the request body for updating PRD status
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateStatus updates a PRD's status
// @Summary Update PRD status
// @Tags PRDs
// @Accept json
// @Produce json
// @Param id path string true "PRD ID"
// @Param request body UpdateStatusRequest true "New status"
// @Success 200 {object} models.PRD
// @Router /api/v1/prds/{id}/status [patch]
func (h *PRDHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Authorization check
	orgID := getOrganizationID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Verify PRD belongs to user's organization before updating
	prd, err := h.prdService.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRD not found"})
		return
	}

	if prd.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.prdService.UpdateStatus(ctx, id, models.PRDStatus(req.Status)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	prd, err = h.prdService.GetByID(ctx, id)
	if err != nil {
		// Status was updated but failed to retrieve - still report success
		c.JSON(http.StatusOK, gin.H{"message": "status updated"})
		return
	}
	c.JSON(http.StatusOK, prd)
}

// ListVersions lists all versions of a PRD
func (h *PRDHandler) ListVersions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Authorization check
	orgID := getOrganizationID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	ctx := c.Request.Context()

	// Verify PRD belongs to user's organization
	prd, err := h.prdService.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRD not found"})
		return
	}

	if prd.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	versions, err := h.prdService.ListVersions(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

// GetVersion retrieves a specific version
func (h *PRDHandler) GetVersion(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Authorization check
	orgID := getOrganizationID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	versionNum, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version number"})
		return
	}

	ctx := c.Request.Context()

	// Verify PRD belongs to user's organization
	prd, err := h.prdService.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRD not found"})
		return
	}

	if prd.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	version, err := h.prdService.GetVersion(ctx, id, versionNum)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	c.JSON(http.StatusOK, version)
}

// CreateVersionRequest represents the request for creating a version
type CreateVersionRequest struct {
	Summary string `json:"summary"`
}

// CreateVersion creates a new version
func (h *PRDHandler) CreateVersion(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Summary = "Manual version"
	}

	userID := getUserID(c)
	orgID := getOrganizationID(c)

	// Validate authentication
	if userID == uuid.Nil || orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	ctx := c.Request.Context()

	// Verify PRD belongs to user's organization
	prd, err := h.prdService.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRD not found"})
		return
	}

	if prd.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	version, err := h.prdService.CreateVersion(ctx, id, userID, req.Summary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, version)
}

// RestoreVersion restores a PRD to a specific version
func (h *PRDHandler) RestoreVersion(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	versionNum, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version number"})
		return
	}

	userID := getUserID(c)
	orgID := getOrganizationID(c)

	// Validate authentication
	if userID == uuid.Nil || orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	ctx := c.Request.Context()

	// Verify PRD belongs to user's organization
	prd, err := h.prdService.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRD not found"})
		return
	}

	if prd.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.prdService.RestoreVersion(ctx, id, versionNum, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	prd, err = h.prdService.GetByID(ctx, id)
	if err != nil {
		// Version was restored but failed to retrieve - still report success
		c.JSON(http.StatusOK, gin.H{"message": "version restored"})
		return
	}
	c.JSON(http.StatusOK, prd)
}

// GenerateFromNotesRequest represents the request for AI generation
type GenerateFromNotesRequest struct {
	Notes       string  `json:"notes" binding:"required"`
	MeetingType string  `json:"meeting_type,omitempty"`
	TemplateID  *string `json:"template_id,omitempty"`
}

// GenerateFromNotes generates a PRD from meeting notes
// @Summary Generate PRD from notes
// @Tags PRDs
// @Accept json
// @Produce json
// @Param request body GenerateFromNotesRequest true "Meeting notes"
// @Success 201 {object} models.PRD
// @Router /api/v1/prds/generate [post]
func (h *PRDHandler) GenerateFromNotes(c *gin.Context) {
	var req GenerateFromNotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	orgID := getOrganizationID(c)

	// Validate authentication
	if userID == uuid.Nil || orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var templateID *uuid.UUID
	if req.TemplateID != nil {
		tid, err := uuid.Parse(*req.TemplateID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template_id"})
			return
		}
		templateID = &tid
	}

	prd, err := h.prdService.GenerateFromNotes(c.Request.Context(), service.GenerateFromNotesInput{
		OrganizationID: orgID,
		OwnerID:        userID,
		Notes:          req.Notes,
		MeetingType:    req.MeetingType,
		TemplateID:     templateID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, prd)
}

// RefineRequest represents the request for refining a PRD
type RefineRequest struct {
	Feedback string `json:"feedback" binding:"required"`
}

// Refine refines a PRD based on feedback
// @Summary Refine PRD with AI
// @Tags PRDs
// @Accept json
// @Produce json
// @Param id path string true "PRD ID"
// @Param request body RefineRequest true "Feedback"
// @Success 200 {object} models.PRD
// @Router /api/v1/prds/{id}/refine [post]
func (h *PRDHandler) Refine(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Authorization check
	orgID := getOrganizationID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req RefineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Verify PRD belongs to user's organization
	existingPRD, err := h.prdService.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRD not found"})
		return
	}

	if existingPRD.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	userID := getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	prd, err := h.prdService.RefinePRD(ctx, id, req.Feedback, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prd)
}

// GenerateUserStories generates user stories for a PRD
// @Summary Generate user stories
// @Tags PRDs
// @Produce json
// @Param id path string true "PRD ID"
// @Success 200 {object} map[string][]models.UserStory
// @Router /api/v1/prds/{id}/stories [post]
func (h *PRDHandler) GenerateUserStories(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Authorization check
	orgID := getOrganizationID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	ctx := c.Request.Context()

	// Verify PRD belongs to user's organization
	prd, err := h.prdService.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRD not found"})
		return
	}

	if prd.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	userID := getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	stories, err := h.prdService.GenerateUserStories(ctx, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stories": stories})
}

// LinkToJiraRequest represents the request for linking to Jira
type LinkToJiraRequest struct {
	EpicKey string `json:"epic_key" binding:"required"`
}

// LinkToJira links a PRD to a Jira epic
func (h *PRDHandler) LinkToJira(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Authorization check
	orgID := getOrganizationID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req LinkToJiraRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Verify PRD belongs to user's organization
	prd, err := h.prdService.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRD not found"})
		return
	}

	if prd.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.prdService.LinkToJira(ctx, id, req.EpicKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "linked to Jira"})
}

// LinkToConfluenceRequest represents the request for linking to Confluence
type LinkToConfluenceRequest struct {
	PageID string `json:"page_id" binding:"required"`
}

// LinkToConfluence links a PRD to a Confluence page
func (h *PRDHandler) LinkToConfluence(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Authorization check
	orgID := getOrganizationID(c)
	if orgID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req LinkToConfluenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Verify PRD belongs to user's organization
	prd, err := h.prdService.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRD not found"})
		return
	}

	if prd.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.prdService.LinkToConfluence(ctx, id, req.PageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "linked to Confluence"})
}

// Helper functions to get user and organization from context
func getUserID(c *gin.Context) uuid.UUID {
	if id, exists := c.Get("user_id"); exists {
		if uid, ok := id.(uuid.UUID); ok {
			return uid
		}
	}
	return uuid.Nil
}

func getOrganizationID(c *gin.Context) uuid.UUID {
	if id, exists := c.Get("organization_id"); exists {
		if oid, ok := id.(uuid.UUID); ok {
			return oid
		}
	}
	return uuid.Nil
}
