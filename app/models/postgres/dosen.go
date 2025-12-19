package postgres

import "time"

type Dosen struct { 
    ID         string    `json:"id"`
    UserID     string    `json:"user_id"`
    DosenID    string    `json:"dosen_id"` 
    Department string    `json:"department"`
    CreatedAt  time.Time `json:"created_at"`
}