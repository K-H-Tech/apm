package handlers

import (
	"net/http"
	"strconv"

	"github.com/K-H-Tech/apm/internal/models"
	"github.com/K-H-Tech/apm/internal/service"
	"github.com/K-H-Tech/apm/pkg/priority"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PriorityHandler handles priority-related HTTP requests
type PriorityHandler struct {
	priorityService *service.PriorityService
}

// NewPriorityHandler creates a new priority handler
func NewPriorityHandler(priorityService *service.PriorityService) *PriorityHandler {
	return &PriorityHandler{
		priorityService: priorityService,
	}
}

// RegisterRoutes registers priority routes
func (h *PriorityHandler) RegisterRoutes(r *gin.RouterGroup) {
	priorities := r.Group("/priorities")
	{
		// Calculate scores
		priorities.POST("/rice", h.CalculateRICE)
		priorities.POST("/ice", h.CalculateICE)
		priorities.POST("/moscow", h.SetMoSCoW)
		priorities.POST("/weighted", h.CalculateWeighted)

		// Rankings
		priorities.GET("/rankings", h.GetRankings)

		// Distribution
		priorities.GET("/moscow/distribution", h.GetMoSCoWDistribution)

		// Guidance
		priorities.GET("/guidance/rice", h.GetRICEGuidance)
		priorities.GET("/guidance/ice", h.GetICEGuidance)
		priorities.GET("/guidance/moscow", h.GetMoSCoWGuidance)
		priorities.GET("/guidance/weighted/templates", h.GetWeightedTemplates)

		// PRD-specific
		priorities.GET("/prd/:prd_id", h.GetByPRD)
		priorities.GET("/prd/:prd_id/history/:framework", h.GetScoreHistory)
		priorities.DELETE("/:id", h.Delete)

		// Suggestion
		priorities.POST("/moscow/suggest", h.SuggestMoSCoW)
	}
}

// CalculateRICERequest represents the request for RICE calculation
type CalculateRICERequest struct {
	PRDID      string `json:"prd_id" binding:"required"`
	Reach      int    `json:"reach" binding:"required,min=1"`
	Impact     int    `json:"impact" binding:"required,oneof=25 50 100 200 300"`
	Confidence int    `json:"confidence" binding:"required,min=50,max=100"`
	Effort     int    `json:"effort" binding:"required,min=1"`
}

// CalculateRICE calculates a RICE score
// @Summary Calculate RICE score
// @Tags Priorities
// @Accept json
// @Produce json
// @Param request body CalculateRICERequest true "RICE parameters"
// @Success 200 {object} models.PriorityScore
// @Router /api/v1/priorities/rice [post]
func (h *PriorityHandler) CalculateRICE(c *gin.Context) {
	var req CalculateRICERequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prdID, err := uuid.Parse(req.PRDID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prd_id"})
		return
	}

	userID := getUserID(c)

	score, err := h.priorityService.CalculateRICE(c.Request.Context(), service.CalculateRICEInput{
		PRDID:        prdID,
		Reach:        req.Reach,
		Impact:       req.Impact,
		Confidence:   req.Confidence,
		Effort:       req.Effort,
		CalculatedBy: userID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Include interpretation
	interpretation := h.priorityService.GetRICEInterpretation(score.FinalScore)

	c.JSON(http.StatusOK, gin.H{
		"score":          score,
		"interpretation": interpretation,
	})
}

// CalculateICERequest represents the request for ICE calculation
type CalculateICERequest struct {
	PRDID      string `json:"prd_id" binding:"required"`
	Impact     int    `json:"impact" binding:"required,min=1,max=10"`
	Confidence int    `json:"confidence" binding:"required,min=1,max=10"`
	Ease       int    `json:"ease" binding:"required,min=1,max=10"`
}

// CalculateICE calculates an ICE score
// @Summary Calculate ICE score
// @Tags Priorities
// @Accept json
// @Produce json
// @Param request body CalculateICERequest true "ICE parameters"
// @Success 200 {object} models.PriorityScore
// @Router /api/v1/priorities/ice [post]
func (h *PriorityHandler) CalculateICE(c *gin.Context) {
	var req CalculateICERequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prdID, err := uuid.Parse(req.PRDID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prd_id"})
		return
	}

	userID := getUserID(c)

	score, err := h.priorityService.CalculateICE(c.Request.Context(), service.CalculateICEInput{
		PRDID:        prdID,
		Impact:       req.Impact,
		Confidence:   req.Confidence,
		Ease:         req.Ease,
		CalculatedBy: userID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Include interpretation
	interpretation := h.priorityService.GetICEInterpretation(score.FinalScore)

	c.JSON(http.StatusOK, gin.H{
		"score":          score,
		"interpretation": interpretation,
	})
}

// SetMoSCoWRequest represents the request for MoSCoW categorization
type SetMoSCoWRequest struct {
	PRDID    string `json:"prd_id" binding:"required"`
	Category string `json:"category" binding:"required,oneof=must should could wont"`
}

// SetMoSCoW sets the MoSCoW category
// @Summary Set MoSCoW category
// @Tags Priorities
// @Accept json
// @Produce json
// @Param request body SetMoSCoWRequest true "MoSCoW category"
// @Success 200 {object} models.PriorityScore
// @Router /api/v1/priorities/moscow [post]
func (h *PriorityHandler) SetMoSCoW(c *gin.Context) {
	var req SetMoSCoWRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prdID, err := uuid.Parse(req.PRDID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prd_id"})
		return
	}

	userID := getUserID(c)

	score, err := h.priorityService.SetMoSCoW(c.Request.Context(), service.SetMoSCoWInput{
		PRDID:        prdID,
		Category:     models.MoSCoWCategory(req.Category),
		CalculatedBy: userID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"score": score})
}

// CalculateWeightedRequest represents the request for weighted scoring
type CalculateWeightedRequest struct {
	PRDID    string                      `json:"prd_id" binding:"required"`
	Criteria []priority.WeightedCriterion `json:"criteria" binding:"required,min=1"`
}

// CalculateWeighted calculates a weighted score
// @Summary Calculate weighted score
// @Tags Priorities
// @Accept json
// @Produce json
// @Param request body CalculateWeightedRequest true "Weighted criteria"
// @Success 200 {object} models.PriorityScore
// @Router /api/v1/priorities/weighted [post]
func (h *PriorityHandler) CalculateWeighted(c *gin.Context) {
	var req CalculateWeightedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prdID, err := uuid.Parse(req.PRDID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prd_id"})
		return
	}

	userID := getUserID(c)

	score, err := h.priorityService.CalculateWeighted(c.Request.Context(), service.CalculateWeightedInput{
		PRDID:        prdID,
		Criteria:     req.Criteria,
		CalculatedBy: userID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Include interpretation
	interpretation := h.priorityService.GetWeightedInterpretation(score.FinalScore)

	c.JSON(http.StatusOK, gin.H{
		"score":          score,
		"interpretation": interpretation,
	})
}

// GetRankings retrieves ranked PRDs
// @Summary Get PRD rankings
// @Tags Priorities
// @Produce json
// @Param framework query string true "Priority framework" Enums(rice, ice, moscow, weighted)
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/priorities/rankings [get]
func (h *PriorityHandler) GetRankings(c *gin.Context) {
	framework := c.Query("framework")
	if framework == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "framework is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	orgID := getOrganizationID(c)

	rankings, total, err := h.priorityService.GetRankings(c.Request.Context(), orgID, models.PriorityFramework(framework), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rankings": rankings,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetMoSCoWDistribution retrieves the MoSCoW distribution
// @Summary Get MoSCoW distribution
// @Tags Priorities
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/priorities/moscow/distribution [get]
func (h *PriorityHandler) GetMoSCoWDistribution(c *gin.Context) {
	orgID := getOrganizationID(c)

	dist, err := h.priorityService.GetMoSCoWDistribution(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get validation warnings
	warnings := h.priorityService.ValidateMoSCoWDistribution(dist)

	c.JSON(http.StatusOK, gin.H{
		"distribution": dist,
		"warnings":     warnings,
	})
}

// GetByPRD retrieves all priority scores for a PRD
// @Summary Get priority scores for a PRD
// @Tags Priorities
// @Produce json
// @Param prd_id path string true "PRD ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/priorities/prd/{prd_id} [get]
func (h *PriorityHandler) GetByPRD(c *gin.Context) {
	prdID, err := uuid.Parse(c.Param("prd_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prd_id"})
		return
	}

	scores, err := h.priorityService.GetByPRD(c.Request.Context(), prdID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
}

// GetScoreHistory retrieves score history for a PRD
// @Summary Get score history for a PRD
// @Tags Priorities
// @Produce json
// @Param prd_id path string true "PRD ID"
// @Param framework path string true "Framework"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/priorities/prd/{prd_id}/history/{framework} [get]
func (h *PriorityHandler) GetScoreHistory(c *gin.Context) {
	prdID, err := uuid.Parse(c.Param("prd_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prd_id"})
		return
	}

	framework := c.Param("framework")

	history, err := h.priorityService.GetScoreHistory(c.Request.Context(), prdID, models.PriorityFramework(framework))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

// Delete deletes a priority score
// @Summary Delete a priority score
// @Tags Priorities
// @Param id path string true "Score ID"
// @Success 204
// @Router /api/v1/priorities/{id} [delete]
func (h *PriorityHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.priorityService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// SuggestMoSCoWRequest represents the request for MoSCoW suggestion
type SuggestMoSCoWRequest struct {
	IsCritical    bool `json:"is_critical"`
	HasWorkaround bool `json:"has_workaround"`
	IsRegulatory  bool `json:"is_regulatory"`
	IsLegal       bool `json:"is_legal"`
	BusinessValue int  `json:"business_value" binding:"min=1,max=10"`
	UserRequested bool `json:"user_requested"`
}

// SuggestMoSCoW suggests a MoSCoW category
// @Summary Suggest MoSCoW category
// @Tags Priorities
// @Accept json
// @Produce json
// @Param request body SuggestMoSCoWRequest true "Criteria"
// @Success 200 {object} map[string]string
// @Router /api/v1/priorities/moscow/suggest [post]
func (h *PriorityHandler) SuggestMoSCoW(c *gin.Context) {
	var req SuggestMoSCoWRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := h.priorityService.SuggestMoSCoW(service.SuggestMoSCoWInput{
		IsCritical:    req.IsCritical,
		HasWorkaround: req.HasWorkaround,
		IsRegulatory:  req.IsRegulatory,
		IsLegal:       req.IsLegal,
		BusinessValue: req.BusinessValue,
		UserRequested: req.UserRequested,
	})

	c.JSON(http.StatusOK, gin.H{
		"suggested_category": category,
	})
}

// GetRICEGuidance returns RICE scoring guidance
// @Summary Get RICE guidance
// @Tags Priorities
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/priorities/guidance/rice [get]
func (h *PriorityHandler) GetRICEGuidance(c *gin.Context) {
	guidance := h.priorityService.GetRICEGuidance()
	c.JSON(http.StatusOK, guidance)
}

// GetICEGuidance returns ICE scoring guidance
// @Summary Get ICE guidance
// @Tags Priorities
// @Produce json
// @Success 200 {object} []priority.ICEComponentGuidance
// @Router /api/v1/priorities/guidance/ice [get]
func (h *PriorityHandler) GetICEGuidance(c *gin.Context) {
	guidance := h.priorityService.GetICEGuidance()
	c.JSON(http.StatusOK, guidance)
}

// GetMoSCoWGuidance returns MoSCoW category guidance
// @Summary Get MoSCoW guidance
// @Tags Priorities
// @Produce json
// @Success 200 {object} []priority.MoSCoWCategoryInfo
// @Router /api/v1/priorities/guidance/moscow [get]
func (h *PriorityHandler) GetMoSCoWGuidance(c *gin.Context) {
	info := h.priorityService.GetMoSCoWInfo()
	c.JSON(http.StatusOK, info)
}

// GetWeightedTemplates returns weighted scoring templates
// @Summary Get weighted scoring templates
// @Tags Priorities
// @Produce json
// @Success 200 {object} []priority.CommonCriteriaTemplate
// @Router /api/v1/priorities/guidance/weighted/templates [get]
func (h *PriorityHandler) GetWeightedTemplates(c *gin.Context) {
	templates := h.priorityService.GetWeightedTemplates()
	c.JSON(http.StatusOK, templates)
}
