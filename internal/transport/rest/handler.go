package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"hippo/internal/domain"
	"hippo/internal/platform/logger"
)

type Medicine interface {
	Create(ctx context.Context, medicament domain.Medicine) (int64, error)
	Update(ctx context.Context, id int64, upd domain.UpdateMedicine) error
	Delete(ctx context.Context, id int64) error
	GetAll(ctx context.Context) ([]domain.Medicine, error)
	GetByID(ctx context.Context, id int64) (domain.Medicine, error)
}

type User interface {
	SignUp(ctx context.Context, sInfo domain.SignUpInfo) (int64, error)
	SignIn(ctx context.Context, sInfo domain.SignInInfo) (string, string, error)
	ParseToken(ctx context.Context, accessToken string) (int64, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
}

type Handler struct {
	medicinesService Medicine
	usersService     User
	log              logger.Logger
	timeout          time.Duration
}

func NewHandler(med Medicine, usr User, log logger.Logger, timeout time.Duration) *Handler {
	return &Handler{
		medicinesService: med,
		usersService:     usr,
		log:              log,
		timeout:          timeout,
	}
}

func (h *Handler) InitRouter() *mux.Router {
	r := mux.NewRouter()

	r.Use(
		h.timeoutMiddleware,
		h.loggingMiddleware,
	)

	auth := r.PathPrefix("/auth").Subrouter()
	{
		auth.HandleFunc("/sign-up", h.handleSignUp).Methods(http.MethodPost)
		auth.HandleFunc("/sign-in", h.handleSignIn).Methods(http.MethodGet)
		auth.HandleFunc("/refresh", h.handleRefresh).Methods(http.MethodGet)
	}

	api := r.PathPrefix("/api/v1").Subrouter()
	{
		api.Use(h.authMiddleware)

		medicines := api.PathPrefix("/medicines").Subrouter()
		{
			medicines.HandleFunc("", h.handleCreateMedicine).Methods(http.MethodPost)
			medicines.HandleFunc("", h.handleGetAllMedicines).Methods(http.MethodGet)
			medicines.HandleFunc("/{id:[0-9]+}", h.handleGetMedicineByID).Methods(http.MethodGet)
			medicines.HandleFunc("/{id:[0-9]+}", h.handleUpdateMedicine).Methods(http.MethodPut)
			medicines.HandleFunc("/{id:[0-9]+}", h.handleDeleteMedicine).Methods(http.MethodDelete)
		}
	}

	// TODO Health check endpoint
	//r.HandleFunc("/health", h.handleHealthCheck).Methods(http.MethodGet)

	return r
}

func (h *Handler) logError(handlerName string, err error) {
	h.log.Error("gRPC audit server connection established",
		logger.Err(err),
		logger.String("handler", handlerName),
	)
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, status int, op string, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			w.Write([]byte(`{"code":"encoding_error","message":"Failed to encode response"}`))
			h.logError(op, err)
		}
	}
}
