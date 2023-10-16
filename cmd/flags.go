package cmd

import (
	openapi_types "github.com/oapi-codegen/runtime/types"
	"time"
)

type UUID struct {
	openapi_types.UUID
}

func (u *UUID) Set(s string) error {
	return u.UnmarshalText([]byte(s))
}

func (u *UUID) Type() string {
	return "UUID"
}

type Time struct {
	time.Time
}

func (t *Time) Set(s string) error {
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	t.Time = tm
	return nil
}

func (t *Time) Type() string {
	return "Time"
}

type genericVar struct {
	s string
	t string
}

func (t *genericVar) String() string {
	return t.s
}

func (t *genericVar) Set(s string) error {
	t.s = s
	return nil
}

func (t *genericVar) Type() string {
	return t.t
}
