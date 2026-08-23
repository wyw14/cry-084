package httptransport

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	hazardapp "github.com/local/cry-084/internal/application/hazard"
	inspectionapp "github.com/local/cry-084/internal/application/inspection"
	"github.com/local/cry-084/internal/domain/inspection"
	"github.com/local/cry-084/internal/domain/shared"
)

type Handler struct {
	validate    *validator.Validate
	inspections *inspectionapp.Service
	hazards     *hazardapp.Service
}

func NewHandler(inspections *inspectionapp.Service, hazards *hazardapp.Service) *Handler {
	return &Handler{validate: validator.New(), inspections: inspections, hazards: hazards}
}

type claimRequest struct {
	Version uint64 `json:"version" validate:"required"`
}

func (h *Handler) ClaimTask(c *gin.Context) {
	var req claimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, shared.ErrValidation)
		return
	}
	actor, _ := c.Get("actor")
	task, err := h.inspections.Claim(c.Request.Context(), inspectionapp.ClaimCommand{TaskID: shared.ID(c.Param("id")), Actor: actor.(shared.Actor), Expected: shared.Version(req.Version)})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

type beginRequest struct {
	Version uint64 `json:"version" validate:"required"`
}

func (h *Handler) BeginTask(c *gin.Context) {
	var req beginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, shared.ErrValidation)
		return
	}
	actor, _ := c.Get("actor")
	task, err := h.inspections.Begin(c.Request.Context(), inspectionapp.BeginCommand{TaskID: shared.ID(c.Param("id")), Actor: actor.(shared.Actor), Expected: shared.Version(req.Version)})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

type evidenceRequest struct {
	ID      string `json:"id" validate:"required"`
	FileID  string `json:"file_id" validate:"required"`
	Digest  string `json:"digest" validate:"required"`
	Caption string `json:"caption"`
}
type resultRequest struct {
	Version   uint64                `json:"version" validate:"required"`
	AssetID   string                `json:"asset_id" validate:"required"`
	QRSecret  string                `json:"qr_secret" validate:"required"`
	ScannedAt time.Time             `json:"scanned_at" validate:"required"`
	Kind      inspection.ResultKind `json:"kind" validate:"required"`
	Answers   map[string]any        `json:"answers"`
	Evidence  []evidenceRequest     `json:"evidence"`
}

func (h *Handler) SubmitResult(c *gin.Context) {
	var req resultRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validate.Struct(req) != nil {
		writeError(c, shared.ErrValidation)
		return
	}
	actor, _ := c.Get("actor")
	evidence := make([]inspection.Evidence, 0, len(req.Evidence))
	for _, item := range req.Evidence {
		evidence = append(evidence, inspection.Evidence{ID: shared.ID(item.ID), FileID: shared.ID(item.FileID), Digest: item.Digest, Caption: item.Caption, At: req.ScannedAt})
	}
	result, err := h.inspections.Submit(c.Request.Context(), inspectionapp.SubmitCommand{TaskID: shared.ID(c.Param("id")), Actor: actor.(shared.Actor), AssetID: shared.ID(req.AssetID), QRSecret: req.QRSecret, ScannedAt: req.ScannedAt, Kind: req.Kind, Answers: req.Answers, Evidence: evidence, Expected: shared.Version(req.Version)})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

type assignRequest struct {
	TeamID     string `json:"team_id" validate:"required"`
	AssigneeID string `json:"assignee_id" validate:"required"`
	Version    uint64 `json:"version" validate:"required"`
}

func (h *Handler) AssignHazard(c *gin.Context) {
	var req assignRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validate.Struct(req) != nil {
		writeError(c, shared.ErrValidation)
		return
	}
	actor, _ := c.Get("actor")
	record, err := h.hazards.Assign(c.Request.Context(), actor.(shared.Actor), shared.ID(c.Param("id")), shared.ID(req.TeamID), shared.ID(req.AssigneeID), shared.Version(req.Version))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, record)
}

type rectifyRequest struct {
	Evidence []string `json:"evidence" validate:"required,min=1"`
	Reason   string   `json:"reason" validate:"required"`
	Version  uint64   `json:"version" validate:"required"`
}

func (h *Handler) RectifyHazard(c *gin.Context) {
	var req rectifyRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validate.Struct(req) != nil {
		writeError(c, shared.ErrValidation)
		return
	}
	ids := make([]shared.ID, len(req.Evidence))
	for i, id := range req.Evidence {
		ids[i] = shared.ID(id)
	}
	actor, _ := c.Get("actor")
	record, err := h.hazards.Rectify(c.Request.Context(), actor.(shared.Actor), shared.ID(c.Param("id")), ids, req.Reason, shared.Version(req.Version))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, record)
}

type reviewRequest struct {
	Accepted bool   `json:"accepted"`
	Reason   string `json:"reason" validate:"required"`
	Version  uint64 `json:"version" validate:"required"`
}

func (h *Handler) ReviewHazard(c *gin.Context) {
	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validate.Struct(req) != nil {
		writeError(c, shared.ErrValidation)
		return
	}
	actor, _ := c.Get("actor")
	record, err := h.hazards.ReviewAndRestore(c.Request.Context(), actor.(shared.Actor), shared.ID(c.Param("id")), req.Accepted, req.Reason, shared.Version(req.Version))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, record)
}
