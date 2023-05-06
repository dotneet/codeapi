package handler

import (
	"github.com/dotneet/codeapi/storage"
)

type Handlers struct {
	Bucket storage.ImageBucket
}

func NewHandlers(bucket storage.ImageBucket) *Handlers {
	return &Handlers{
		Bucket: bucket,
	}
}
