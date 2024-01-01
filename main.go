package main

import (
	"dmdb/media"
	"dmdb/storage"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/google/uuid"
)

var cache = map[string]media.Media{}
var db = storage.FileStore{BasePath: "db"}

var authorizedKeys = []string{}

func main() {
	authorizedKeys = strings.Split(os.Getenv("AUTHORIZED_KEYS"), ",")
	fmt.Printf("Allowed keys: %v\n", authorizedKeys)

	http.HandleFunc("/v1/gmid/", handleGMID)

	err := http.ListenAndServe(":9999", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Shutting down")
	} else if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

func handleGMID(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	requestID := uuid.New().String()
	fmt.Printf("[%s] %s: %s\n", requestID, r.Method, r.URL.Path)

	pathParts := strings.Split(r.URL.Path, "/")
	gmid := pathParts[len(pathParts)-1]

	if !media.IsValidGMID(gmid) {
		fmt.Printf("[%s] %s\n", requestID, "Invalid GMID")
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
		return
	case http.MethodGet:
		m := getMedia(gmid)

		bytes, err := json.Marshal(m)
		if err != nil {
			fmt.Printf("[%s] (%s) %s\n", requestID, gmid, err)
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(bytes)
		return
	case http.MethodPost:
		key := r.URL.Query().Get("api_key")
		if key == "" || !slices.Contains(authorizedKeys, key) {
			fmt.Printf("[%s] (%s) %s\n", requestID, gmid, "Unauthorized")
			http.Error(w, "401", http.StatusUnauthorized)
			return
		}

		m := getMedia(gmid)

		var updatedMedia media.Media
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&updatedMedia)
		if err != nil {
			fmt.Printf("[%s] (%s) %s\n", requestID, gmid, err)
			http.Error(w, "400", http.StatusBadRequest)
			return
		}

		if updatedMedia.GMID != m.GMID {
			fmt.Printf("[%s] (%s) %s\n", requestID, gmid, "Bad request")
			http.Error(w, "400", http.StatusBadRequest)
			return
		}

		for provider, id := range updatedMedia.IDs {
			fmt.Printf("[%s] (%s) Updating ID %s=%s\n", requestID, gmid, provider, id)
			m.IDs[provider] = id
		}

		db.UpdateMedia(m)
		cache[m.GMID] = m

		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "404 page not found", http.StatusNotFound)
}

// gets a media element, or creates a new one if not found
func getMedia(gmid string) media.Media {
	if m, ok := cache[gmid]; ok {
		//fmt.Printf("Cache hit: %s\n", gmid)
		return m
	}

	if m, err := db.GetMedia(gmid); err == nil {
		//fmt.Printf("Storage hit: %s\n", gmid)
		cache[gmid] = *m

		return *m
	}

	//fmt.Printf("Creating new media for: %s\n", gmid)
	m := media.New(gmid)
	cache[gmid] = m

	db.UpdateMedia(m)

	return m
}
