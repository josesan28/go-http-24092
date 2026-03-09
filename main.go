package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Message struct {
	Message string `json:"message"`
}

var teams []Team

func main() {
	loadTeams()

	http.HandleFunc("/api/ping", pingHandler)
	http.HandleFunc("/api/teams", teamsHandler)

	log.Println("POST JSON API running on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func loadTeams() {
	file, err := os.ReadFile("./data/teams.json")
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	err = json.Unmarshal(file, &teams)
	if err != nil {
		log.Fatal("Error parsing JSON:", err)
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	response := Message{
		Message: "pong",
	}

	writeJSON(w, http.StatusOK, response)
}

func teamsHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:
		handleGetTeams(w, r)

	case http.MethodPost:
		handleCreateTeam(w, r)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetTeams(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	idParam := query.Get("id")

	if idParam == "" {
		writeJSON(w, http.StatusOK, teams)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}

	for _, team := range teams {
		if team.ID == id {
			writeJSON(w, http.StatusOK, team)
			return
		}
	}

	http.Error(w, "Team not found", http.StatusNotFound)
}

func handleCreateTeam(w http.ResponseWriter, r *http.Request) {

	var newTeam Team

	err := json.NewDecoder(r.Body).Decode(&newTeam)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if newTeam.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	newTeam.ID = generateNextID()

	teams = append(teams, newTeam)
	saveTeams()

	writeJSON(w, http.StatusCreated, newTeam)
}

func generateNextID() int {
	maxID := 0

	for _, team := range teams {
		if team.ID > maxID {
			maxID = team.ID
		}
	}

	return maxID + 1
}

func saveTeams() {
	data, err := json.MarshalIndent(teams, "", "  ")
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	err = os.WriteFile("./data/teams.json", data, 0644)
	if err != nil {
			log.Println("Error writing file:", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}