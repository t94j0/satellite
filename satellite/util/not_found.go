package util

import "errors"

type NotFound struct {
	Redirect string
	Render   string
}

var ErrNotFoundConfig = errors.New("both not_found redirect and render cannot be set at the same time")

func NewNotFound(redirect, render string) (NotFound, error) {
	if redirect != "" && render != "" {
		return NotFound{}, ErrNotFoundConfig
	}

	return NotFound{
		Redirect: redirect,
		Render:   render,
	}, nil
}

// ShouldWarn returns true if the redirect or render portion is empty
func (nf NotFound) ShouldWarn() bool {
	return nf.Redirect == "" && nf.Render == ""
}
