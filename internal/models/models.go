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
	Bio       string         `gorm:"size:50" json:"bio"`
	Role      string         `gorm:"default:user" json:"role"` // user, admin, editor
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Posts     []Post         `json:"posts,omitempty"`
	Comments  []Comment      `json:"comments,omitempty"`
}

type Category struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Slug        string         `gorm:"uniqueIndex;size:50;not null" json:"slug"`
	Description string         `gorm:"size:200" json:"description"`
	ParentID    *uint          `json:"parent_id"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdateAt    time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Posts       []Post         `gorm:"many2many:post_categories;" json:"posts,omitempty"`
}

type Tag struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Name      string         `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Slug      string         `gorm:"uniqueIndex;size:50;not null" json:"slug"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Posts     []Post         `gorm:"many2many:post_tags;" json:"posts,omitempty"`
}

type Post struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Title        string         `gorm:"size:200;not null" json:"title"`
	Slug         string         `gorm:"uniqueIndex;size:200;not null" json:"slug"`
	Content      string         `gorm:"type:longtext;not null" json:"content"`
	Summary      string         `gorm:"size:500" json:"summary"`
	Cover        string         `json:"cover"`
	Views        int            `gorm:"default:0" json:"views"`
	Likes        int            `gorm:"default:0" json:"likes"`
	Status       string         `gorm:"default:draft" json:"status"` // draft, published
	UserID       uint           `gorm:"not null" json:"user_id"`
	User         User           `json:"user,omitempty"`
	CategoryID   *uint          `json:"category_id"`
	Category     *Category      `json:"category,omitempty"`
	Tags         string         `gorm:"many2many:post_tags;" json:"tags,omitempty"` // JSON string or comma separated
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	PublicshedAt *time.Time     `json:"published_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Comments     []Comment      `json:"comments,omitempty"`
}

type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `json:"user,omitempty"`
	PostID    uint           `gorm:"not null;index" json:"post_id"`
	Post      Post           `json:"-"`
	ParentID  *uint          `json:"parent_id"`
	Status    string         `gorm:"default:pending" json:"status"` // pending, approved, spam, deleted
	IPAddress string         `json:"ip_address"`
	UserAgent string         `json:"usre_agent"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Replies   []Comment      `gorm:"foreignkey:ParentID" json:"replies,omitempty"`
}

type Upload struct {
	ID           uint           `gorm:"priamryKey" json:"id"`
	Filename     string         `gorm:"size:255;not null;" json:"filename"`
	OriginalName string         `gorm:"size:255;not null" json:"original_name"`
	Path         string         `gorm:"size:500;not null" json:"path"`
	URL          string         `gorm:"size:500;not null" json:"url"`
	Size         int64          `json:"size"`
	MimeType     string         `gorm:"size:100" json:"mime_type"`
	UserID       uint           `gorm:"not null" json:"user_id"`
	User         User           `json:"user,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
