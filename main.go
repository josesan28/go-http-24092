package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
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
	http.HandleFunc("/api/players/", playerByIDHandler)

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
	query := r.URL.Query()
	idParam := query.Get("id")
	teamParam := query.Get("team")
	positionParam := query.Get("position")

	if idParam != "" {
		id, err := strconv.Atoi(idParam)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid id parameter")
			return
		}

		for _, player := range players {
			if player.ID == id {
				writeJSON(w, http.StatusOK, player)
				return
			}
		}

		writeError(w, http.StatusNotFound, "Player not found")
		return
	}

	result := []Player{}
	for _, player := range players {
		if teamParam != "" && player.Team != teamParam {
			continue
		}
		if positionParam != "" && player.Position != positionParam {
			continue
		}
		result = append(result, player)
	}

	writeJSON(w, http.StatusOK, result)
}

func handleCreatePlayer(w http.ResponseWriter, r *http.Request) {
	var newPlayer Player

	err := json.NewDecoder(r.Body).Decode(&newPlayer)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if newPlayer.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if newPlayer.Team == "" {
		writeError(w, http.StatusBadRequest, "Team is required")
		return
	}
	if newPlayer.Position == "" {
		writeError(w, http.StatusBadRequest, "Position is required")
		return
	}
	if newPlayer.JerseyNumber <= 0 {
		writeError(w, http.StatusBadRequest, "Jersey number must be greater than 0")
		return
	}
	if newPlayer.BirthYear < 1970 || newPlayer.BirthYear > 2010 {
		writeError(w, http.StatusBadRequest, "Birth year must be between 1970 and 2010")
		return
	}
	if newPlayer.Touchdowns < 0 {
		writeError(w, http.StatusBadRequest, "Touchdowns cannot be negative")
		return
	}

	newPlayer.ID = generateNextID()
	players = append(players, newPlayer)
	savePlayers()

	writeJSON(w, http.StatusCreated, newPlayer)
}

func generateNextID() int {
	maxID := 0
	for _, player := range players {
		if player.ID > maxID {
			maxID = player.ID
		}
	}
	return maxID + 1
}

func playerByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/players/"):]

	if idStr == "" {
		writeError(w, http.StatusBadRequest, "Missing player ID")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid player ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGetPlayerByID(w, id)
	case http.MethodPut:
		handleUpdatePlayer(w, r, id)
	case http.MethodPatch:
		handlePatchPlayer(w, r, id)
	case http.MethodDelete:
		handleDeletePlayer(w, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func handleGetPlayerByID(w http.ResponseWriter, id int) {
	for _, player := range players {
		if player.ID == id {
			writeJSON(w, http.StatusOK, player)
			return
		}
	}
	writeError(w, http.StatusNotFound, "Player not found")
}

func handleDeletePlayer(w http.ResponseWriter, id int) {
	for i, player := range players {
		if player.ID == id {
			players = append(players[:i], players[i+1:]...)
			savePlayers()
			writeJSON(w, http.StatusOK, map[string]string{"message": "Player deleted successfully"})
			return
		}
	}
	writeError(w, http.StatusNotFound, "Player not found")
}

func handleUpdatePlayer(w http.ResponseWriter, r *http.Request, id int) {
	var updated Player

	err := json.NewDecoder(r.Body).Decode(&updated)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if updated.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if updated.Team == "" {
		writeError(w, http.StatusBadRequest, "Team is required")
		return
	}
	if updated.Position == "" {
		writeError(w, http.StatusBadRequest, "Position is required")
		return
	}
	if updated.JerseyNumber <= 0 {
		writeError(w, http.StatusBadRequest, "Jersey number must be greater than 0")
		return
	}
	if updated.BirthYear < 1970 || updated.BirthYear > 2010 {
		writeError(w, http.StatusBadRequest, "Birth year must be between 1970 and 2010")
		return
	}
	if updated.Touchdowns < 0 {
		writeError(w, http.StatusBadRequest, "Touchdowns cannot be negative")
		return
	}

	for i, player := range players {
		if player.ID == id {
			updated.ID = id
			players[i] = updated
			savePlayers()
			writeJSON(w, http.StatusOK, updated)
			return
		}
	}

	writeError(w, http.StatusNotFound, "Player not found")
}

func handlePatchPlayer(w http.ResponseWriter, r *http.Request, id int) {
	var partial map[string]interface{}

	err := json.NewDecoder(r.Body).Decode(&partial)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	for i, player := range players {
		if player.ID == id {
			if name, ok := partial["name"].(string); ok {
				if name == "" {
					writeError(w, http.StatusBadRequest, "Name cannot be empty")
					return
				}
				players[i].Name = name
			}
			if team, ok := partial["team"].(string); ok {
				if team == "" {
					writeError(w, http.StatusBadRequest, "Team cannot be empty")
					return
				}
				players[i].Team = team
			}
			if position, ok := partial["position"].(string); ok {
				if position == "" {
					writeError(w, http.StatusBadRequest, "Position cannot be empty")
					return
				}
				players[i].Position = position
			}
			if jersey, ok := partial["jersey_number"].(float64); ok {
				if int(jersey) <= 0 {
					writeError(w, http.StatusBadRequest, "Jersey number must be greater than 0")
					return
				}
				players[i].JerseyNumber = int(jersey)
			}
			if birthYear, ok := partial["birth_year"].(float64); ok {
				if int(birthYear) < 1970 || int(birthYear) > 2010 {
					writeError(w, http.StatusBadRequest, "Birth year must be between 1970 and 2010")
					return
				}
				players[i].BirthYear = int(birthYear)
			}
			if tds, ok := partial["touchdowns"].(float64); ok {
				if int(tds) < 0 {
					writeError(w, http.StatusBadRequest, "Touchdowns cannot be negative")
					return
				}
				players[i].Touchdowns = int(tds)
			}

			savePlayers()
			writeJSON(w, http.StatusOK, players[i])
			return
		}
	}

	writeError(w, http.StatusNotFound, "Player not found")
}

func savePlayers() {
	data, err := json.MarshalIndent(players, "", "  ")
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}
	err = os.WriteFile("./data/players.json", data, 0644)
	if err != nil {
		log.Println("Error writing file:", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}