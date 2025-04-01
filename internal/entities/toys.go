package entities

import "time"

type Category struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

type Tag struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

type Master struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"userId"`
	Info      string    `json:"info"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
