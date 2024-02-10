package models

import "time"

type User struct {
	ID           int64
	UserID       int64
	Phone        string
	Username     string
	Firstname    string
	Lastname     *string
	ProfileImage *string
	DateOfBirth  time.Time
	Preferences  []Category
}

type Permission struct {
	ID   int64
	Name string
}

type Role struct {
	ID          int64
	Name        string
	Permissions []int64
}

type Category struct {
	ID   int64
	Name string
}

type Event struct {
	ID          int64
	Title       string
	Description string
	MaxAge      *int64
	MinAge      *int64
	Images      []Image
	CreatedAt   time.Time
	Categories  []Category
	Locations   []Location
	Managers    []Manager
	Attendees   []User
}

type Manager struct {
	EventId int64
	User    User
	Role    Role
}

type Location struct {
	ID        int64
	EventID   int64
	Address   string
	Longitude float64
	Latitude  float64
	StartsAt  time.Time
	EndsAt    time.Time
	Seats     *int64
	Archived  bool
}

type Image struct {
	ID        int64
	EventID   int64
	Url       string
	CreatedAt time.Time
}
