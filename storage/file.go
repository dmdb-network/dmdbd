package storage

import (
	"crypto/md5"
	"dmdb/media"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileStore struct {
	BasePath string
}

func (fs FileStore) getFilePath(gmid string) string {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(gmid)))
	filename := filepath.Join(fs.BasePath, gmid[0:1], hash[0:2], gmid+".json")

	return filename
}

func (fs FileStore) GetMedia(gmid string) (*media.Media, error) {

	file := fs.getFilePath(gmid)

	jsonFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer jsonFile.Close()

	var media media.Media
	bytes, _ := io.ReadAll(jsonFile)

	err = json.Unmarshal(bytes, &media)
	if err != nil {
		return nil, err
	}

	return &media, nil
}

func (fs FileStore) UpdateMedia(m media.Media) error {
	file := fs.getFilePath(m.GMID)

	if _, err := os.Stat(file); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(file), 0700)
	}

	jsonFile, err := os.Create(file)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer jsonFile.Close()

	bytes, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = jsonFile.Write(bytes)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
