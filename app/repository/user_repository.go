package repository

import (
	"context"
	"fmt"
	"pelaporan_prestasi/app/models/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: db}
}

// FindByUsername untuk Login
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*postgres.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, r.name as role_name, u.is_active
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1
	`
	var user postgres.User
	// Scan role_name ke field sementara atau abaikan jika belum butuh di struct User
	var roleName string 

	err := r.DB.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName, &user.RoleID, &roleName, &user.IsActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByID untuk Get Profile
func (r *UserRepository) FindByID(ctx context.Context, id string) (*postgres.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.full_name, u.role_id, r.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`
	var user postgres.User
	var roleName string

	err := r.DB.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName, &user.RoleID, &roleName,
	)

	if err != nil {
		return nil, err
	}
	// Opsional: Kamu bisa masukkan roleName ke struct User jika ada fieldnya
	return &user, nil
}

func (r *UserRepository) FindAll(ctx context.Context) ([]postgres.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.full_name, u.role_id, r.name as role_name, u.is_active, u.created_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		ORDER BY u.created_at DESC
	`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []postgres.User
	for rows.Next() {
		var user postgres.User
		var roleName string // Variable dummy buat nampung role name jika belum ada di struct User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FullName, 
			&user.RoleID, &roleName, &user.IsActive, &user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// Create - Menambah user baru
func (r *UserRepository) Create(ctx context.Context, user *postgres.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`
	// Perhatikan urutan variabelnya
	err := r.DB.QueryRow(ctx, query,
		user.Username, user.Email, user.PasswordHash, user.FullName, user.RoleID, true,
	).Scan(&user.ID)
	
	return err
}

// Update - Mengubah data user
func (r *UserRepository) Update(ctx context.Context, user *postgres.User) error {
	query := `
		UPDATE users 
		SET full_name = $1, role_id = $2, email = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.DB.Exec(ctx, query, user.FullName, user.RoleID, user.Email, user.ID)
	return err
}

// UpdatePassword - Khusus ganti password (terpisah biar aman)
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, newHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.DB.Exec(ctx, query, newHash, userID)
	return err
}

func (r *UserRepository) UpdateRole(ctx context.Context, userID, roleID string) error {
	query := `UPDATE users SET role_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.DB.Exec(ctx, query, roleID, userID)
	return err
}

// Delete - Soft Delete (Ubah is_active jadi false)
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	return err
}