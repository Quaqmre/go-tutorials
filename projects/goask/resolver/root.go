package resolver

import (
	"github.com/pkg/errors"
	"goask/core/adapter"
)

type Root struct {
	Query
	Mutation
}


type stdResolver struct {
	data adapter.Data
	log  logger
}

func (r *stdResolver) check() error {
	if r.data == nil {
		return errors.New("stdResolver.data is not initialized")
	}
	if r.log == nil {
		return errors.New("stdResolver.log is not initialized")
	}
	return nil
}
