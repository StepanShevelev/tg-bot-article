package db

import (
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	Title        string  `json:"title" db:"title" gorm:"unique" `
	Text         string  `json:"text" db:"text"`
	SubText      string  `json:"sub_text" db:"sub_text"`
	Images       []Image `json:"images" db:"images" gorm:"foreignKey:PostID"`
	ArticleUrl   string  `json:"article_url" db:"article_url"`
	WhoTookMe    string  `json:"who_took_me" db:"who_took_me"`
	WhoCreatedMe string  `json:"who_created_me" db:"who_created_me"`
	Name         string  `json:"name" db:"name"`
}

type Image struct {
	gorm.Model
	Name   []byte `json:"name" db:"name"`
	PostID uint   `json:"post_id" db:"post_id"`
}

// ErrLogs storage some error logs
type ErrLogs struct {
	gorm.Model
	Error string `json:"error" db:"error"`
	Place string `json:"place" db:"place"`
	Count int    `json:"count" db:"count"`
}

type User struct {
	gorm.Model
	Name     string `json:"name" db:"name"`
	Position string `json:"position" db:"position"`
}
