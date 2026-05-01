package cuid

import (
	"github.com/nrednav/cuid2"
)

func New() string {
	return cuid2.Generate()
}

func IsValid(id string) bool {
	if len(id) < 21 || len(id) > 24 {
		return false
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
