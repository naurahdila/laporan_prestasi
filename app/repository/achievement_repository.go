package repository

import (
	"context"
	"time"

	"pelaporan_prestasi/app/models/mongo"
	"pelaporan_prestasi/app/models/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementRepository struct {
	PgPool    *pgxpool.Pool
	MongoColl *mongo.Collection
}

func NewAchievementRepository(pg *pgxpool.Pool, mongoDB *mongo.Database) *AchievementRepository {
	coll := mongoDB.Collection("achievement") 
	return &AchievementRepository{PgPool: pg, MongoColl: coll}
}

// --- MONGO OPERATIONS ---

func (r *AchievementRepository) InsertMongo(ctx context.Context, data *mongodb.Achievement) (string, error) {
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	// Init array/map biar gak null di JSON
	if data.Attachments == nil { data.Attachments = []mongodb.Attachment{} }
	if data.Details == nil { data.Details = make(map[string]interface{}) }
	if data.Tags == nil { data.Tags = []string{} }

	res, err := r.MongoColl.InsertOne(ctx, data)
	if err != nil { return "", err }
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *AchievementRepository) FindContentByMongoID(ctx context.Context, hexID string) (*mongodb.Achievement, error) {
	oid, err := primitive.ObjectIDFromHex(hexID)
	if err != nil { return nil, err }
	var data mongodb.Achievement
	if err := r.MongoColl.FindOne(ctx, bson.M{"_id": oid}).Decode(&data); err != nil { return nil, err }
	return &data, nil
}

func (r *AchievementRepository) UpdateContentMongo(ctx context.Context, hexID string, data mongodb.Achievement) error {
	oid, _ := primitive.ObjectIDFromHex(hexID)
	update := bson.M{"$set": bson.M{
		"title":           data.Title,
		"description":     data.Description,
		"achievementType": data.AchievementType,
		"points":          data.Points,
		"tags":            data.Tags,
		"details":         data.Details,
		"updatedAt":       time.Now(),
	}}
	_, err := r.MongoColl.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

// Fitur Push Attachment (Array)
func (r *AchievementRepository) AddAttachmentMongo(ctx context.Context, hexID string, att mongodb.Attachment) error {
	oid, _ := primitive.ObjectIDFromHex(hexID)
	update := bson.M{
		"$push": bson.M{"attachments": att},
		"$set":  bson.M{"updatedAt": time.Now()},
	}
	_, err := r.MongoColl.UpdateOne(ctx, bson.M{"_id": oid}, update)
	return err
}

// --- POSTGRES OPERATIONS ---

func (r *AchievementRepository) InsertPostgres(ctx context.Context, ref *postgres.AchievementReference) error {
	query := `INSERT INTO achievement_references (student_id, mongo_achievement_id, status, created_at, updated_at) VALUES ($1, $2, 'DRAFT', NOW(), NOW()) RETURNING id`
	return r.PgPool.QueryRow(ctx, query, ref.StudentID, ref.MongoAchievementID).Scan(&ref.ID)
}

func (r *AchievementRepository) FindRefByID(ctx context.Context, id string) (*postgres.AchievementReference, error) {
	query := `SELECT id, student_id, mongo_achievement_id, status, rejection_note FROM achievement_references WHERE id = $1`
	var ref postgres.AchievementReference
	err := r.PgPool.QueryRow(ctx, query, id).Scan(&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status, &ref.RejectionNote)
	if err != nil { return nil, err }
	return &ref, nil
}

func (r *AchievementRepository) FindRefsByStudentID(ctx context.Context, studentID string) ([]postgres.AchievementReference, error) {
	query := `SELECT id, student_id, mongo_achievement_id, status FROM achievement_references WHERE student_id = $1 ORDER BY created_at DESC`
	return r.fetchRefs(ctx, query, studentID)
}

func (r *AchievementRepository) FindRefsByAdvisorID(ctx context.Context, advisorID string) ([]postgres.AchievementReference, error) {
	query := `SELECT ar.id, ar.student_id, ar.mongo_achievement_id, ar.status FROM achievement_references ar JOIN mahasiswa m ON ar.student_id = m.user_id WHERE m.advisor_id = $1 ORDER BY ar.created_at DESC`
	return r.fetchRefs(ctx, query, advisorID)
}

func (r *AchievementRepository) FindAllRefs(ctx context.Context) ([]postgres.AchievementReference, error) {
	query := `SELECT id, student_id, mongo_achievement_id, status FROM achievement_references ORDER BY created_at DESC`
	return r.fetchRefs(ctx, query)
}

func (r *AchievementRepository) UpdateStatus(ctx context.Context, id, status string, notes *string) error {
	query := `UPDATE achievement_references SET status=$1, rejection_note=$2, updated_at=NOW() WHERE id=$3`
	_, err := r.PgPool.Exec(ctx, query, status, notes, id)
	return err
}

func (r *AchievementRepository) UpdateVerified(ctx context.Context, id, verifierID string) error {
	query := `UPDATE achievement_references SET status='VERIFIED', verified_by=$1, verified_at=NOW(), updated_at=NOW() WHERE id=$2`
	_, err := r.PgPool.Exec(ctx, query, verifierID, id)
	return err
}

func (r *AchievementRepository) DeleteRef(ctx context.Context, id string) error {
	_, err := r.PgPool.Exec(ctx, "DELETE FROM achievement_references WHERE id=$1", id)
	return err
}

// --- HISTORY & CHECKS ---

func (r *AchievementRepository) AddHistory(ctx context.Context, h postgres.AchievementHistory) error {
	query := `INSERT INTO achievement_histories (achievement_id, changed_by, previous_status, new_status, remarks, created_at) VALUES ($1, $2, $3, $4, $5, NOW())`
	_, err := r.PgPool.Exec(ctx, query, h.AchievementID, h.ChangedBy, h.PreviousStatus, h.NewStatus, h.Remarks)
	return err
}

func (r *AchievementRepository) GetHistory(ctx context.Context, achievementID string) ([]postgres.AchievementHistory, error) {
	query := `SELECT id, achievement_id, changed_by, previous_status, new_status, remarks, created_at FROM achievement_histories WHERE achievement_id = $1 ORDER BY created_at DESC`
	rows, err := r.PgPool.Query(ctx, query, achievementID)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []postgres.AchievementHistory
	for rows.Next() {
		var h postgres.AchievementHistory
		rows.Scan(&h.ID, &h.AchievementID, &h.ChangedBy, &h.PreviousStatus, &h.NewStatus, &h.Remarks, &h.CreatedAt)
		list = append(list, h)
	}
	return list, nil
}

func (r *AchievementRepository) IsAdvisorOfRef(ctx context.Context, refID, advisorUserID string) (bool, error) {
    // LOGIKA FINAL: JOIN ke DOSEN untuk mencocokkan USER_ID
    query := `
        SELECT EXISTS (
            SELECT 1 
            FROM achievement_references ar 
            JOIN mahasiswa m ON ar.student_id = m.user_id
            JOIN dosen d ON m.advisor_id = d.id 
            WHERE ar.id = $1 
            AND d.user_id = $2
        )`
    
    var exists bool
    // $1 = ID Prestasi, $2 = ID Token Dosen (User ID)
    err := r.PgPool.QueryRow(ctx, query, refID, advisorUserID).Scan(&exists)
    return exists, err
}

func (r *AchievementRepository) fetchRefs(ctx context.Context, query string, args ...interface{}) ([]postgres.AchievementReference, error) {
	rows, err := r.PgPool.Query(ctx, query, args...)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []postgres.AchievementReference
	for rows.Next() {
		var ref postgres.AchievementReference
		rows.Scan(&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status)
		list = append(list, ref)
	}
	return list, nil
}