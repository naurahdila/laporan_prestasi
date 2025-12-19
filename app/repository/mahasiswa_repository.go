package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MahasiswaRepository struct {
	PgPool *pgxpool.Pool
}

func NewMahasiswaRepository(pg *pgxpool.Pool) *MahasiswaRepository {
	return &MahasiswaRepository{PgPool: pg}
}

type MahasiswaData struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	Name          string `json:"name"`
	NIM           string `json:"nim"`
	ProgramStudy  string `json:"program_study"`
	AcademicYear  string `json:"academic_year"`
	AdvisorName   string `json:"advisor_name"`
	AdvisorID     string `json:"advisor_id"`
	AdvisorUserID string `json:"advisor_user_id"` 
}

func (r *MahasiswaRepository) GetAll(ctx context.Context) ([]MahasiswaData, error) {
	query := `
		SELECT 
			m.id, m.user_id, u.full_name, m.nim, m.program_study, m.academic_year,
			COALESCE(ud.full_name, 'Belum Ada') as advisor_name,
			COALESCE(m.advisor_id, '00000000-0000-0000-0000-000000000000') as advisor_id,
			COALESCE(d.user_id, '00000000-0000-0000-0000-000000000000') as advisor_user_id
		FROM mahasiswa m
		JOIN users u ON m.user_id = u.id
		LEFT JOIN dosen d ON m.advisor_id = d.id
		LEFT JOIN users ud ON d.user_id = ud.id
		ORDER BY m.nim ASC`
	
	rows, err := r.PgPool.Query(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()

	var list []MahasiswaData
	for rows.Next() {
		var m MahasiswaData
		err := rows.Scan(&m.ID, &m.UserID, &m.Name, &m.NIM, &m.ProgramStudy, &m.AcademicYear, &m.AdvisorName, &m.AdvisorID, &m.AdvisorUserID)
		if err != nil { return nil, err }
		list = append(list, m)
	}
	return list, nil
}

func (r *MahasiswaRepository) GetByID(ctx context.Context, id string) (*MahasiswaData, error) {
    query := `
        SELECT 
            m.id, m.user_id, u.full_name, m.nim, m.program_study, m.academic_year,
            COALESCE(ud.full_name, 'Belum Ada') as advisor_name,
            COALESCE(m.advisor_id, '00000000-0000-0000-0000-000000000000') as advisor_id,
            COALESCE(d.user_id, '00000000-0000-0000-0000-000000000000') as advisor_user_id
        FROM mahasiswa m
        JOIN users u ON m.user_id = u.id
        LEFT JOIN dosen d ON m.advisor_id = d.id
        LEFT JOIN users ud ON d.user_id = ud.id
        WHERE m.id = $1`
    
    var m MahasiswaData
    err := r.PgPool.QueryRow(ctx, query, id).Scan(
        &m.ID, &m.UserID, &m.Name, &m.NIM, &m.ProgramStudy, &m.AcademicYear,
        &m.AdvisorName, &m.AdvisorID, &m.AdvisorUserID,
    )
    if err != nil { return nil, err }
    return &m, nil
}

func (r *MahasiswaRepository) UpdateAdvisor(ctx context.Context, studentID, advisorID string) error {
	_, err := r.PgPool.Exec(ctx, "UPDATE mahasiswa SET advisor_id = $1 WHERE id = $2", advisorID, studentID)
	return err
}