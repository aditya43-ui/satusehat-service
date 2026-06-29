package usecase

import (
	"errors"
	"net/http"
	"time"

	"service/internal/interfaces/satusehat"
	"service/internal/satusehat/usecase/encounter"
	pkgErrors "service/pkg/errors"
	"service/pkg/response"
	"service/pkg/utils/validator"

	"github.com/gin-gonic/gin"
)

type EncounterHandler struct {
	service encounter.Service
}

func NewEncounterHandler(service encounter.Service) *EncounterHandler {
	return &EncounterHandler{
		service: service,
	}
}

func (h *EncounterHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/satusehat/encounter")
	{
		group.POST("", h.Create)
		// group.POST("/sync/:idxdaftar", h.SyncFromSIMRS)
		group.GET("", h.Search)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.PATCH("/:id", h.Patch)
	}
}

// handleSatuSehatError mengekstrak OperationOutcome agar JSON bisa tampil terstruktur
func handleSatuSehatError(c *gin.Context, err error) {
	var ssErr *satusehat.ErrorOperationOutcome
	if errors.As(err, &ssErr) {
		c.Error(err) // Log error asli
		c.JSON(ssErr.StatusCode, gin.H{
			"status":  "error",
			"message": ssErr.Outcome,
			"error": gin.H{
				"severity":  "error",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}
	appErr := pkgErrors.FromError(err)
	response.ErrorWithLog(c, err, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
}

func (h *EncounterHandler) Create(c *gin.Context) {
	var req encounter.EncounterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		customErr := validator.TranslateError(err)
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", customErr)
		return
	}

	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}

	// Di sini, Anda bisa menyimpan result.RawResponse atau result.ID ke database log jika diperlukan.
	// Contoh:
	// go logService.Create(c.Request.Context(), "encounter_create", req, result.RawResponse, http.StatusCreated)

	response.Success(c, http.StatusCreated, "Successfully created Encounter", result.FullResponse)
}

func (h *EncounterHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req encounter.EncounterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		customErr := validator.TranslateError(err)
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan tidak valid", customErr)
		return
	}

	result, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully updated Encounter", result.FullResponse)
}

func (h *EncounterHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req encounter.EncounterPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		customErr := validator.TranslateError(err)
		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format permintaan patch tidak valid", customErr)
		return
	}

	result, err := h.service.Patch(c.Request.Context(), id, req)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully patched Encounter", result.FullResponse)
}

func (h *EncounterHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved Encounter", result.FullResponse)
}

func (h *EncounterHandler) Search(c *gin.Context) {
	// Get query parameters string natively and fetch it directly (like: ?patient=xxx&status=active)
	queryParams := c.Request.URL.Query()
	result, err := h.service.Search(c.Request.Context(), queryParams)
	if err != nil {
		handleSatuSehatError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Successfully retrieved Encounters", result.FullResponse)
}

// func (h *EncounterHandler) SyncFromSIMRS(c *gin.Context) {
// 	idxdaftarStr := c.Param("idxdaftar")
// 	idxdaftar, err := strconv.ParseInt(idxdaftarStr, 10, 64)
// 	if err != nil {
// 		response.ErrorWithLog(c, err, http.StatusBadRequest, "Format idxdaftar tidak valid", nil)
// 		return
// 	}

// 	result, err := h.service.SyncFromSIMRS(c.Request.Context(), idxdaftar)
// 	if err != nil {
// 		handleSatuSehatError(c, err)
// 		return
// 	}
// 	response.Success(c, http.StatusOK, "Successfully synced Encounter from SIMRS", result.FullResponse)
// }
