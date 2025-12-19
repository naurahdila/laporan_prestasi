package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepository struct {
	PgPool *pgxpool.Pool
}

func NewReportRepository(pg *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{PgPool: pg}
}

type GlobalStatistics struct {
	TotalMahasiswa  int            `json:"total_mahasiswa"`
	TotalPrestasi   int            `json:"total_prestasi"`
	PrestasiByStatus map[string]int `json:"prestasi_by_status"`
}

func (r *ReportRepository) GetGlobalStats(ctx context.Context) (*GlobalStatistics, error) {
	stats := &GlobalStatistics{
		PrestasiByStatus: make(map[string]int),
	}

	err := r.PgPool.QueryRow(ctx, "SELECT COUNT(*) FROM mahasiswa").Scan(&stats.TotalMahasiswa)
	if err != nil { return nil, err }

	err = r.PgPool.QueryRow(ctx, "SELECT COUNT(*) FROM achievement_references").Scan(&stats.TotalPrestasi)
	if err != nil { return nil, err }

	rows, err := r.PgPool.Query(ctx, "SELECT status, COUNT(*) FROM achievement_references GROUP BY status")
	if err != nil { return nil, err }
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		stats.PrestasiByStatus[status] = count
	}

	return stats, nil
}

func (r *ReportRepository) GetStudentStats(ctx context.Context, userID string) (map[string]int, error) {
	stats := make(map[string]int)
	query := `SELECT status, COUNT(*) FROM achievement_references WHERE student_id = $1 GROUP BY status`
	
	rows, err := r.PgPool.Query(ctx, query, userID)
	if err != nil { return nil, err }
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		stats[status] = count
	}
	return stats, nil
}