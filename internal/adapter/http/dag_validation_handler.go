package http

import (
	"davidterranova/jurigen/internal/usecase"
	"davidterranova/jurigen/pkg/xhttp"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// ValidateRequest represents the request payload for DAG validation
//
// @Description DAG validation request containing the DAG structure to validate
type ValidateRequest struct {
	DAG DAGPresenter `json:"dag" validate:"required"`
}

// ValidationResultPresenter represents the validation result for API responses
//
// @Description Comprehensive DAG validation results including errors, warnings, and statistics
type ValidationResultPresenter struct {
	IsValid    bool                          `json:"is_valid" example:"true"`
	Errors     []ValidationErrorPresenter    `json:"errors,omitempty"`
	Warnings   []ValidationWarningPresenter  `json:"warnings,omitempty"`
	Statistics ValidationStatisticsPresenter `json:"statistics"`
}

// ValidationErrorPresenter represents a validation error
//
// @Description Detailed validation error with context information
type ValidationErrorPresenter struct {
	Code     string `json:"code" example:"DAG_HAS_CYCLES"`
	Message  string `json:"message" example:"DAG contains 1 cycle(s). A valid DAG must be acyclic"`
	NodeID   string `json:"node_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	AnswerID string `json:"answer_id,omitempty" example:"fc28c4b6-d185-cf56-a7e4-dead499ff1e8"`
	Severity string `json:"severity" example:"error"`
}

// ValidationWarningPresenter represents a validation warning
//
// @Description Non-critical validation issue that doesn't prevent DAG usage
type ValidationWarningPresenter struct {
	Code     string `json:"code" example:"DEEP_NESTING"`
	Message  string `json:"message" example:"DAG has deep nesting which may impact performance"`
	NodeID   string `json:"node_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	AnswerID string `json:"answer_id,omitempty" example:"fc28c4b6-d185-cf56-a7e4-dead499ff1e8"`
}

// ValidationStatisticsPresenter provides DAG structure statistics
//
// @Description Statistical information about the DAG structure and validation results
type ValidationStatisticsPresenter struct {
	TotalNodes   int      `json:"total_nodes" example:"5"`
	RootNodes    int      `json:"root_nodes" example:"1"`
	LeafNodes    int      `json:"leaf_nodes" example:"2"`
	TotalAnswers int      `json:"total_answers" example:"12"`
	MaxDepth     int      `json:"max_depth" example:"3"`
	HasCycles    bool     `json:"has_cycles" example:"false"`
	RootNodeIDs  []string `json:"root_node_ids,omitempty"`
	LeafNodeIDs  []string `json:"leaf_node_ids,omitempty"`
	CyclePaths   []string `json:"cycle_paths,omitempty"`
}

// ValidateDAG validates a DAG structure without saving it
//
// @Summary Validate Legal Case DAG
// @Description Validate a DAG structure to ensure it meets all requirements (single root node, acyclic, valid relationships)
// @Tags DAGs
// @Accept json
// @Produce json
// @Param dag body ValidateRequest true "DAG structure to validate"
// @Success 200 {object} ValidationResultPresenter "DAG validation completed (may contain errors)"
// @Failure 400 {object} xhttp.ErrorResponse "Invalid request body"
// @Failure 500 {object} xhttp.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /dags/validate [post]
func (h *dagHandler) ValidateDAG(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse the request body
	var validateRequest ValidateRequest
	err := json.NewDecoder(r.Body).Decode(&validateRequest)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode validation request body")
		xhttp.WriteError(ctx, w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Convert presenter to DAG
	dagToValidate := h.presenterToDAG(validateRequest.DAG)

	// Validate the DAG using the validator service
	validator := usecase.NewDAGValidator()
	validationResult := validator.ValidateDAG(dagToValidate)

	// Convert validation result to presenter format
	resultPresenter := h.validationResultToPresenter(validationResult)

	// Return validation results (always 200 OK, even if DAG is invalid)
	xhttp.WriteObject(ctx, w, http.StatusOK, resultPresenter)
}

// validationResultToPresenter converts usecase ValidationResult to ValidationResultPresenter
func (h *dagHandler) validationResultToPresenter(result usecase.ValidationResult) ValidationResultPresenter {
	presenter := ValidationResultPresenter{
		IsValid: result.IsValid,
		Statistics: ValidationStatisticsPresenter{
			TotalNodes:   result.Statistics.TotalNodes,
			RootNodes:    result.Statistics.RootNodes,
			LeafNodes:    result.Statistics.LeafNodes,
			TotalAnswers: result.Statistics.TotalAnswers,
			MaxDepth:     result.Statistics.MaxDepth,
			HasCycles:    result.Statistics.HasCycles,
		},
	}

	// Copy statistics directly as they're already strings
	presenter.Statistics.RootNodeIDs = result.Statistics.RootNodeIDs
	presenter.Statistics.LeafNodeIDs = result.Statistics.LeafNodeIDs
	presenter.Statistics.CyclePaths = result.Statistics.CyclePaths

	// Convert errors
	presenter.Errors = make([]ValidationErrorPresenter, len(result.Errors))
	for i, err := range result.Errors {
		presenter.Errors[i] = ValidationErrorPresenter{
			Code:     err.Code,
			Message:  err.Message,
			NodeID:   err.NodeID,
			AnswerID: err.AnswerID,
			Severity: err.Severity,
		}
	}

	// Convert warnings
	presenter.Warnings = make([]ValidationWarningPresenter, len(result.Warnings))
	for i, warning := range result.Warnings {
		presenter.Warnings[i] = ValidationWarningPresenter{
			Code:     warning.Code,
			Message:  warning.Message,
			NodeID:   warning.NodeID,
			AnswerID: warning.AnswerID,
		}
	}

	return presenter
}
