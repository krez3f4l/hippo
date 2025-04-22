package service

import (
	"context"
	"errors"
	"time"

	//my 2nd github acc :)
	"github.com/krez3f4l/audit_logger/pkg/domain/audit"

	"hippo/internal/domain"
	"hippo/internal/platform/logger"
	"hippo/internal/repository"
)

type MedicationDataRepository interface {
	Create(ctx context.Context, medicament domain.Medicine) (int64, error)
	GetAll(ctx context.Context) ([]domain.Medicine, error)
	GetByID(ctx context.Context, id int64) (domain.Medicine, error)
	Update(ctx context.Context, id int64, upd domain.UpdateMedicine) error
	Delete(ctx context.Context, id int64) error
}

type Medicines struct {
	repo        MedicationDataRepository
	auditClient AuditClient
	log         logger.Logger
}

func NewMedicines(repo MedicationDataRepository, auditClient AuditClient, log logger.Logger) *Medicines {
	return &Medicines{
		repo:        repo,
		auditClient: auditClient,
		log:         log,
	}
}

func (m *Medicines) Create(ctx context.Context, medicament domain.Medicine) (int64, error) {
	/// TODO  business logic
	if medicament.Name == "" {
		return -1, ValidationError{
			Field:   "name",
			Message: "cannot be empty",
		}
	}

	id, err := m.repo.Create(ctx, medicament)
	if err != nil {
		return -1, err
	}

	go m.runAuditCall(ctx, audit.ENTITY_MEDICAMENT, audit.ACTION_CREATE, 0)

	return id, nil
}

func (m *Medicines) GetAll(ctx context.Context) ([]domain.Medicine, error) {
	medicines, err := m.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	go m.runAuditCall(ctx, audit.ENTITY_MEDICAMENT, audit.ACTION_GET, 0)

	return medicines, nil
}

func (m *Medicines) GetByID(ctx context.Context, id int64) (domain.Medicine, error) {
	medicine, err := m.repo.GetByID(ctx, id)
	if err != nil {
		var repoNotFound *repository.NotFoundError
		if errors.As(err, &repoNotFound) {
			return domain.Medicine{}, NewNotFoundError(repoNotFound.Entity, repoNotFound.ID, err)
		}
		return domain.Medicine{}, err
	}

	go m.runAuditCall(ctx, audit.ENTITY_MEDICAMENT, audit.ACTION_GET, id)

	return medicine, nil
}

func (m *Medicines) Update(ctx context.Context, id int64, med domain.UpdateMedicine) error {
	err := m.repo.Update(ctx, id, med)
	if err != nil {
		var repoNotFound *repository.NotFoundError
		if errors.As(err, &repoNotFound) {
			return NewNotFoundError(repoNotFound.Entity, repoNotFound.ID, err)
		}

		var emptyUpdate *repository.ErrEmptyUpdate
		if errors.As(err, &emptyUpdate) {
			return nil
		}

		return err
	}

	go m.runAuditCall(ctx, audit.ENTITY_MEDICAMENT, audit.ACTION_UPDATE, id)

	return nil
}

func (m *Medicines) Delete(ctx context.Context, id int64) error {
	err := m.repo.Delete(ctx, id)
	if err != nil {
		var repoNotFound *repository.NotFoundError
		if errors.As(err, &repoNotFound) {
			return NewNotFoundError(repoNotFound.Entity, repoNotFound.ID, err)
		}

		return err
	}

	go m.runAuditCall(ctx, audit.ENTITY_MEDICAMENT, audit.ACTION_DELETE, id)

	return nil
}

func (m *Medicines) runAuditCall(ctx context.Context, entity, action string, id int64) {
	logErr := m.auditClient.SendLogRequest(ctx, audit.LogItem{
		Entity:    entity,
		Action:    action,
		EntityID:  id,
		Timestamp: time.Now(),
	})
	if logErr != nil {
		m.log.Warn("audit log failed", logger.Err(logErr))
	}
}
