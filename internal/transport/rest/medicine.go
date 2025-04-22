package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"hippo/internal/domain"
	"hippo/internal/service"
)

func (h *Handler) handleCreateMedicine(w http.ResponseWriter, r *http.Request) {
	const op = "handleCreateMedicine"
	ctx := r.Context()

	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		h.respondWithJSON(w, http.StatusUnsupportedMediaType, op, ErrorResponse{
			Code:    "invalid_content_type",
			Message: "Content-Type must be application/json",
		})
		return
	}

	var med domain.Medicine
	if err := json.NewDecoder(r.Body).Decode(&med); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, op, ErrorResponse{
			Code:    "invalid_request_body",
			Message: "Failed to parse request body",
			Details: err.Error(),
		})
		return
	}
	defer r.Body.Close()

	id, err := h.medicinesService.Create(ctx, med)
	if err != nil {
		h.logError(op, err)

		var ve *service.ValidationError
		if errors.As(err, &ve) {
			h.respondWithJSON(w, http.StatusBadRequest, op, ErrorResponse{
				Code:    "validation_failed",
				Message: "Invalid input",
				Details: ve.Error(),
			})
			return
		}

		h.respondWithJSON(w, http.StatusInternalServerError, op, ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to create medicine",
		})
		return
	}

	w.Header().Set("Location", fmt.Sprintf("api/v1/medicines/%d", id))
	h.respondWithJSON(w, http.StatusCreated, op, map[string]string{
		"message": "Medicine created successfully",
	})
}

func (h *Handler) handleGetAllMedicines(w http.ResponseWriter, r *http.Request) {
	const op = "handleGetAllMedicines"
	ctx := r.Context()

	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		h.respondWithJSON(w, http.StatusUnsupportedMediaType, op, ErrorResponse{
			Code:    "invalid_content_type",
			Message: "Content-Type must be application/json",
		})
		return
	}

	medicines, err := h.medicinesService.GetAll(ctx)
	if err != nil {
		h.logError(op, err)

		h.respondWithJSON(w, http.StatusInternalServerError, op, ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to retrieve medicines",
		})
		return
	}

	h.respondWithJSON(w, http.StatusOK, op, medicines)
}

func (h *Handler) handleGetMedicineByID(w http.ResponseWriter, r *http.Request) {
	const op = "handleGetMedicineByID"
	ctx := r.Context()

	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		h.respondWithJSON(w, http.StatusUnsupportedMediaType, op, ErrorResponse{
			Code:    "invalid_content_type",
			Message: "Content-Type must be application/json",
		})
		return
	}

	id, err := getIdFromRequest(r)
	if err != nil {
		h.logError(op, err)

		h.respondWithJSON(w, http.StatusBadRequest, op, ErrorResponse{
			Code:    "invalid_id",
			Message: "Invalid medicine ID",
		})
		return
	}

	medicament, err := h.medicinesService.GetByID(ctx, id)
	if err != nil {
		h.logError(op, err)

		var notFound *service.NotFoundError
		if errors.As(err, &notFound) {
			h.respondWithJSON(w, http.StatusNotFound, op, ErrorResponse{
				Code:    "not_found",
				Message: fmt.Sprintf("%s with ID %v not found", notFound.Entity, notFound.ID),
			})
			return
		}

		h.respondWithJSON(w, http.StatusInternalServerError, op, ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to retrieve medicine",
		})
		return
	}

	h.respondWithJSON(w, http.StatusOK, op, medicament)
}

func (h *Handler) handleUpdateMedicine(w http.ResponseWriter, r *http.Request) {
	const op = "handleUpdateMedicine"
	ctx := r.Context()

	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		h.respondWithJSON(w, http.StatusUnsupportedMediaType, op, ErrorResponse{
			Code:    "invalid_content_type",
			Message: "Content-Type must be application/json",
		})
		return
	}

	id, err := getIdFromRequest(r)
	if err != nil {
		h.logError(op, err)

		h.respondWithJSON(w, http.StatusBadRequest, op, ErrorResponse{
			Code:    "invalid_id",
			Message: "Invalid medicine ID",
		})
		return
	}

	var updMed domain.UpdateMedicine
	if err = json.NewDecoder(r.Body).Decode(&updMed); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, op, ErrorResponse{
			Code:    "invalid_request_body",
			Message: "Failed to parse request body",
			Details: err.Error(),
		})
		return
	}
	defer r.Body.Close()

	err = h.medicinesService.Update(ctx, id, updMed)
	if err != nil {
		h.logError(op, err)

		var notFound *service.NotFoundError
		if errors.As(err, &notFound) {
			h.respondWithJSON(w, http.StatusNotFound, op, ErrorResponse{
				Code:    "not_found",
				Message: fmt.Sprintf("%s with ID %v not found", notFound.Entity, notFound.ID),
			})
			return
		}

		h.respondWithJSON(w, http.StatusInternalServerError, op, ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to update medicine",
		})
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/medicines/%d", id))
	h.respondWithJSON(w, http.StatusOK, op, updMed)
}

func (h *Handler) handleDeleteMedicine(w http.ResponseWriter, r *http.Request) {
	const op = "handleDeleteMedicine"
	ctx := r.Context()

	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		h.respondWithJSON(w, http.StatusUnsupportedMediaType, op, ErrorResponse{
			Code:    "invalid_content_type",
			Message: "Content-Type must be application/json",
		})
		return
	}

	id, err := getIdFromRequest(r)
	if err != nil {
		h.logError(op, err)

		h.respondWithJSON(w, http.StatusBadRequest, op, ErrorResponse{
			Code:    "invalid_id",
			Message: "Invalid medicine ID",
		})
		return
	}

	err = h.medicinesService.Delete(ctx, id)
	if err != nil {
		h.logError(op, err)

		var notFound *service.NotFoundError
		if errors.As(err, &notFound) {
			h.respondWithJSON(w, http.StatusNotFound, op, ErrorResponse{
				Code:    "not_found",
				Message: fmt.Sprintf("%s with ID %v not found", notFound.Entity, notFound.ID),
			})
			return
		}

		h.respondWithJSON(w, http.StatusInternalServerError, op, ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to update medicine",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getIdFromRequest(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return 0, err
	}

	if id <= 0 {
		return 0, errors.New("id can't be 0")
	}

	return id, nil
}
