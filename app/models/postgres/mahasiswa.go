package postgres

import "time"

type Mahasiswa struct { 
    ID           string    `json:"id"`
    UserID       string    `json:"user_id"`
    NIM          string    `json:"nim"`          
    ProgramStudy string    `json:"program_study"`
    AcademicYear string    `json:"academic_year"`
    AdvisorID    *string   `json:"advisor_id"`     
    CreatedAt    time.Time `json:"created_at"`
}