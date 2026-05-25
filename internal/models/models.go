package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email     string         `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	Avatar    string         `json:"avatar"`
	Role      string         `gorm:"default:user" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Posts     []Post         `json:"posts,omitempty"`
}

type Post struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `gorm:"size:200;not null" json:"title"`
	Slug      string         `gorm:"uniqueIndex;size:200;not null" json:"slug"`
	Content   string         `gorm:"type:longtext;not null" json:"content"`
	Summary   string         `gorm:"size:500" json:"summary"`
	Cover     string         `json:"cover"`
	Views     int            `gorm:"default:0" json:"views"`
	Likes     int            `gorm:"default:0" json:"likes"`
	Status    string         `gorm:"default:draft" json:"status"` // draft, published
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `json:"user,omitempty"`
	Tags      string         `json:"tags"` // JSON string or comma separated
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Comments  []Comment      `json:"comments,omitempty"`
}

type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `json:"user,omitempty"`
	PostID    uint           `gorm:"not null" json:"post_id"`
	ParentID  *uint          `json:"parent_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
