package handler

import (
	"github.com/dotneet/codeapi/storage"
)

type Handlers struct {
	ContainerImageName string
	Bucket             storage.ImageBucket
}

func NewHandlers(containerImageName string, bucket storage.ImageBucket) *Handlers {
	return &Handlers{
		ContainerImageName: containerImageName,
		Bucket:             bucket,
	}
}
