package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type Player struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Team         string `json:"team"`
	Position     string `json:"position"`
	JerseyNumber int    `json:"jersey_number"`
	BirthYear    int    `json:"birth_year"`
	Touchdowns   int    `json:"touchdowns"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var players []Player

func main() {
	loadPlayers()

	http.HandleFunc("/api/ping", pingHandler)
	http.HandleFunc("/api/players", playersHandler)

	log.Println("NFL Players API running on :24092")
	log.Fatal(http.ListenAndServe(":24092", nil))
}

func loadPlayers() {
	file, err := os.ReadFile("./data/players.json")
	if err != nil {
		log.Fatal("Error reading file:", err)
	}
	err = json.Unmarshal(file, &players)
	if err != nil {
		log.Fatal("Error parsing JSON:", err)
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "pong"})
}

func playersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetPlayers(w, r)
	case http.MethodPost:
		handleCreatePlayer(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func handleGetPlayers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, players)
}

func handleCreatePlayer(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusCreated, map[string]string{"message": "coming soon"})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}