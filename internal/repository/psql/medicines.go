package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"hippo/internal/domain"
	"hippo/internal/repository"
)

type Medicines struct {
	db *sql.DB
}

func NewMedicines(db *sql.DB) *Medicines {
	return &Medicines{
		db: db,
	}
}

func (m *Medicines) Create(ctx context.Context, medicine domain.Medicine) (int64, error) {
	const op = "repository.psql.medicines.Create"
	const query = `
		INSERT INTO medicines 
			(ndc, name, dosage, form, active_ingredient, pharma_company) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id
	`

	var id int64
	err := m.db.QueryRowContext(ctx, query,
		medicine.NDC,
		medicine.Name,
		medicine.Dosage,
		medicine.Form,
		medicine.ActiveIngredient,
		medicine.PharmaCompany,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("%s: failed to create medicine: %w", op, err)
	}

	return id, nil
}

func (m *Medicines) GetAll(ctx context.Context) ([]domain.Medicine, error) {
	const op = "repository.psql.medicines.GetAll"
	const query = `
		SELECT 
			id, ndc, name, dosage, form, active_ingredient, pharma_company 
		FROM medicines
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get medicines: %w", op, err)
	}
	defer rows.Close()

	var medicines []domain.Medicine
	for rows.Next() {
		var medicine domain.Medicine
		if err = rows.Scan(
			&medicine.ID,
			&medicine.NDC,
			&medicine.Name,
			&medicine.Dosage,
			&medicine.Form,
			&medicine.ActiveIngredient,
			&medicine.PharmaCompany,
		); err != nil {
			return nil, fmt.Errorf("%s: failed to scan medicine row: %w", op, err)
		}
		medicines = append(medicines, medicine)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: error during rows iteration: %w", op, err)
	}

	return medicines, nil
}

func (m *Medicines) GetByID(ctx context.Context, id int64) (domain.Medicine, error) {
	const op = "repository.psql.medicines.GetByID"
	const query = `
		SELECT 
			id, ndc, name, dosage, form, active_ingredient, pharma_company 
		FROM medicines 
		WHERE id = $1
	`

	var medicine domain.Medicine
	err := m.db.QueryRowContext(ctx, query, id).Scan(
		&medicine.ID,
		&medicine.NDC,
		&medicine.Name,
		&medicine.Dosage,
		&medicine.Form,
		&medicine.ActiveIngredient,
		&medicine.PharmaCompany,
	)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.Medicine{}, repository.NewNotFoundError(op, "medicine", id)
	case err != nil:
		return domain.Medicine{}, fmt.Errorf("%s: failed to get medicine by id: %w", op, err)
	}

	return medicine, nil
}

func (m *Medicines) Update(ctx context.Context, id int64, upd domain.UpdateMedicine) error {
	const op = "repository.psql.medicines.Update"
	var (
		setValues []string
		args      []interface{}
		argID     = 1
	)

	if upd.NDC != nil {
		setValues = append(setValues, fmt.Sprintf("ndc = $%d", argID))
		args = append(args, *upd.NDC)
		argID += 1
	}

	if upd.Name != nil {
		setValues = append(setValues, fmt.Sprintf("name = $%d", argID))
		args = append(args, *upd.Name)
		argID += 1
	}

	if upd.Dosage != nil {
		setValues = append(setValues, fmt.Sprintf("dosage = $%d", argID))
		args = append(args, *upd.Dosage)
		argID += 1
	}

	if upd.Form != nil {
		setValues = append(setValues, fmt.Sprintf("form = $%d", argID))
		args = append(args, *upd.Form)
		argID += 1
	}

	if upd.ActiveIngredient != nil {
		setValues = append(setValues, fmt.Sprintf("active_ingredient = $%d", argID))
		args = append(args, *upd.ActiveIngredient)
		argID += 1
	}

	if upd.PharmaCompany != nil {
		setValues = append(setValues, fmt.Sprintf("pharma_company = $%d", argID))
		args = append(args, *upd.PharmaCompany)
		argID += 1
	}

	if len(setValues) == 0 {
		return repository.NewErrEmptyUpdate(op, "medicine")
	}

	args = append(args, id)
	query := fmt.Sprintf(
		"UPDATE medicines SET %s WHERE id = $%d",
		strings.Join(setValues, ", "),
		argID,
	)

	result, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: failed to update medicine: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return repository.NewNotFoundError(op, "medicine", 0)
	}
	return nil
}

func (m *Medicines) Delete(ctx context.Context, id int64) error {
	const op = "repository.psql.medicines.Delete"

	result, err := m.db.ExecContext(ctx, "DELETE FROM medicines WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("%s: failed to delete medicine: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return repository.NewNotFoundError(op, "medicine", id)
	}

	return nil
}
