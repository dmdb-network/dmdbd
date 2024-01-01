package storage

import "dmdb/media"

type Storage interface {
	getMedia(gmid string) (*media.Media, error)
	updateMedia(gmid string) (*media.Media, error)
}
