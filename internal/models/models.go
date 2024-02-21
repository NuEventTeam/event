package models

import "time"

type User struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"userID"`
	Phone        string     `json:"phone"`
	Username     string     `json:"username"`
	Firstname    string     `json:"firstname"`
	Lastname     *string    `json:"lastname"`
	ProfileImage *string    `json:"profileImage"`
	DateOfBirth  time.Time  `json:"dateOfBirth"`
	Preferences  []Category `json:"preferences"`
}

type Permission struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Role struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Permissions []int64 `json:"permissions"`
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Manager struct {
	EventId int64 `json:"eventID"`
	User    User  `json:"user"`
	Role    Role  `json:"role"`
}

type Location struct {
	ID        int64      `json:"id"`
	EventID   int64      `json:"eventID"`
	Address   *string    `json:"address"`
	Longitude *float64   `json:"longitude"`
	Latitude  *float64   `json:"latitude"`
	StartsAt  *time.Time `json:"startsAt"`
	EndsAt    *time.Time `json:"endsAt"`
	Seats     *int64     `json:"seats"`
	Archived  bool       `json:"archived"`
}

type Image struct {
	ID        int64     `json:"id"`
	EventID   int64     `json:"eventID"`
	Url       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
}

type Event struct {
	ID          int64      `json:"id"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	MaxAge      *int64     `json:"maxAge"`
	MinAge      *int64     `json:"minAge"`
	Images      []Image    `json:"images"`
	CreatedAt   time.Time  `json:"created_at"`
	Categories  []Category `json:"categories"`
	CategoryIds []int64    `json:"-"`
	Locations   []Location `json:"locations"`
	Managers    []Manager  `json:"managers"`
	Attendees   []User     `json:"-"`
}
