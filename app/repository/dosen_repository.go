package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DosenRepository struct {
	PgPool *pgxpool.Pool
}

func NewDosenRepository(pg *pgxpool.Pool) *DosenRepository {
	return &DosenRepository{PgPool: pg}
}

// Struct data dosen sesuai kolom di DB
type DosenData struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	NIP    string `json:"nip"`
}

func (r *DosenRepository) GetAll(ctx context.Context) ([]DosenData, error) {
	// Gunakan u.full_name sesuai struct User Anda
	query := `
		SELECT d.id, d.user_id, u.full_name, d.dosen_id
		FROM dosen d
		JOIN users u ON d.user_id = u.id
		ORDER BY u.full_name ASC`
	
	rows, err := r.PgPool.Query(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()

	var list []DosenData
	for rows.Next() {
		var d DosenData
		// Pastikan scan 4 kolom
		err := rows.Scan(&d.ID, &d.UserID, &d.Name, &d.NIP)
		if err != nil { return nil, err }
		list = append(list, d)
	}
	return list, nil
}

func (r *DosenRepository) GetByID(ctx context.Context, id string) (*DosenData, error) {
	query := `
		SELECT d.id, d.user_id, u.full_name, d.dosen_id 
		FROM dosen d 
		JOIN users u ON d.user_id = u.id 
		WHERE d.id = $1`
	
	var d DosenData
	// Pastikan scan 4 kolom
	err := r.PgPool.QueryRow(ctx, query, id).Scan(&d.ID, &d.UserID, &d.Name, &d.NIP)
	if err != nil { return nil, err }
	return &d, nil
}

type AdviseeData struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	NIM          string `json:"nim"`
	ProgramStudy string `json:"program_study"`
}

func (r *DosenRepository) GetAdvisees(ctx context.Context, dosenID string) ([]AdviseeData, error) {
	query := `
		SELECT m.id, u.full_name, m.nim, m.program_study 
		FROM mahasiswa m 
		JOIN users u ON m.user_id = u.id 
		WHERE m.advisor_id = $1 
		ORDER BY m.nim ASC`
	
	rows, err := r.PgPool.Query(ctx, query, dosenID)
	if err != nil { return nil, err }
	defer rows.Close()

	var list []AdviseeData
	for rows.Next() {
		var a AdviseeData
		// Scan 4 kolom: id, name, nim, program_study
		err := rows.Scan(&a.ID, &a.Name, &a.NIM, &a.ProgramStudy)
		if err != nil { return nil, err }
		list = append(list, a)
	}
	return list, nil
}