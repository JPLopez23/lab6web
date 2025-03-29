package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Match struct {
	ID        int    `json:"id"`
	HomeTeam  string `json:"homeTeam"`
	AwayTeam  string `json:"awayTeam"`
	MatchDate string `json:"matchDate"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./matches.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create matches table if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS matches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		home_team TEXT NOT NULL,
		away_team TEXT NOT NULL,
		match_date TEXT NOT NULL
	)`)

	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	// CORS middleware
	r.Use(corsMiddleware)

	// Define routes
	r.HandleFunc("/api/matches", getMatches).Methods("GET")
	r.HandleFunc("/api/matches/{id}", getMatchByID).Methods("GET")
	r.HandleFunc("/api/matches", createMatch).Methods("POST")
	r.HandleFunc("/api/matches/{id}", updateMatch).Methods("PUT")
	r.HandleFunc("/api/matches/{id}", deleteMatch).Methods("DELETE")

	// Optional PATCH endpoints
	r.HandleFunc("/api/matches/{id}/goals", registerGoal).Methods("PATCH")
	r.HandleFunc("/api/matches/{id}/yellowcards", registerYellowCard).Methods("PATCH")
	r.HandleFunc("/api/matches/{id}/redcards", registerRedCard).Methods("PATCH")
	r.HandleFunc("/api/matches/{id}/extratime", setExtraTime).Methods("PATCH")

	// Manejar todas las solicitudes OPTIONS (preflight)
	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("Server starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Permitir todos los orígenes
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Métodos HTTP permitidos
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")

		// Cabeceras permitidas
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Manejar preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Get all matches
func getMatches(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, home_team, away_team, match_date FROM matches")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var m Match
		if err := rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		matches = append(matches, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

// Get match by ID
func getMatchByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var m Match
	err := db.QueryRow("SELECT id, home_team, away_team, match_date FROM matches WHERE id = ?", id).
		Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate)
	if err == sql.ErrNoRows {
		http.Error(w, "Match not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// Create a new match
func createMatch(w http.ResponseWriter, r *http.Request) {
	var m Match
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO matches (home_team, away_team, match_date) VALUES (?, ?, ?)",
		m.HomeTeam, m.AwayTeam, m.MatchDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	m.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)

}

// Update match
func updateMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var m Match
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE matches SET home_team = ?, away_team = ?, match_date = ? WHERE id = ?",
		m.HomeTeam, m.AwayTeam, m.MatchDate, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// Delete a match
func deleteMatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	_, err := db.Exec("DELETE FROM matches WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Placeholder functions for optional PATCH endpoints
func registerGoal(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Goal registered"))
}

func registerYellowCard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Yellow card registered"))
}

func registerRedCard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Red card registered"))
}

func setExtraTime(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Extra time set"))
}
