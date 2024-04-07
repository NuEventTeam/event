package types

import (
	"github.com/bytedance/sonic"
	"time"
)

type Date time.Time

func (f Date) MarshalJSON() ([]byte, error) {
	s := time.Time(f).Format(time.DateOnly)
	b, err := sonic.Marshal(s)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (f *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := sonic.Unmarshal(b, &s); err != nil {
		return err
	}
	value, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return err
	}

	a := Date(value)
	*f = a
	return nil
}

func (f Date) Before(t time.Time) bool {
	return time.Time(f).Before(t)
}

type DateTime time.Time

func (f *DateTime) FromTime(t *time.Time) *DateTime {
	a := DateTime(*t)
	return &a
}
func (f DateTime) MarshalJSON() ([]byte, error) {
	s := time.Time(f).Format(time.DateTime)
	b, err := sonic.Marshal(s)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (f *DateTime) UnmarshalJSON(b []byte) error {
	var s string
	if err := sonic.Unmarshal(b, &s); err != nil {
		return err
	}
	value, err := time.Parse(time.DateTime, s)
	if err != nil {
		return err
	}

	a := DateTime(value)
	*f = a
	return nil
}

func (f DateTime) Before(t time.Time) bool {
	return time.Time(f).Before(t)
}
