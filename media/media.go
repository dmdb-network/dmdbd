package media

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
)

func IsValidGMID(gmid string) bool {
	gmidRegex := regexp.MustCompile(`^(?:(M\.[1-9][0-9]*$|S\.[1-9][0-9]*\.[1-9][0-9]*\.[1-9][0-9]*$))`)
	return gmidRegex.MatchString(gmid)
}

type Media struct {
	GMID string            `json:"gmid"`
	TMDB string            `json:"tmdb"`
	IDs  map[string]string `json:"ids"`
	Hash string            `json:"hash"`
}

func New(GMID string) Media {
	components := strings.Split(GMID, ".")
	hash := md5.Sum([]byte(GMID))
	m := Media{
		GMID: GMID,
		TMDB: components[1],
		IDs:  map[string]string{},
		Hash: fmt.Sprintf("%x", hash),
	}

	return m
}
