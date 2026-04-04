package model

import "time"

type Media struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	ImgbbID   *string   `gorm:"type:varchar(100)" json:"imgbb_id"`
	URL       string    `gorm:"type:text;not null" json:"url"`
	Type      *string   `gorm:"type:varchar(50)" json:"type"`
	Meta      Meta      `gorm:"serializer:json;type:jsonb" json:"meta"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Meta struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	Size       int64     `json:"size"`
	Time       int64     `json:"time"`
	Image      ImageMeta `json:"image"`
	Thumb      ImageMeta `json:"thumb"`
	Title      string    `json:"title"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	DeleteURL  string    `json:"delete_url"`
	Expiration int       `json:"expiration"`
	URLViewer  string    `json:"url_viewer"`
	DisplayURL string    `json:"display_url"`
}

type ImageMeta struct {
	URL       string `json:"url"`
	Mime      string `json:"mime"`
	Name      string `json:"name"`
	Filename  string `json:"filename"`
	Extension string `json:"extension"`
}
