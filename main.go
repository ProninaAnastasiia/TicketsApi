package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type User struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Surename string    `json:"sureName"`
	Passport string    `json:"passportNumber"`
	Age      int       `json:"age"`
	Ticket   time.Time `json:"dateOfTicketExpiry"`
	Price    float32   `json:"price"`
}

// in-memory collection of users.
var users = make(map[int]User)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Get("/users", listUsers)
	r.Post("/users", createUser)
	r.Get("/users/{id}", getUser)

	// Start the server
	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", r)
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	userList := make([]User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userList)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.ID == 0 {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	if user.Age <= 0 {
		http.Error(w, "Age cannot be less than zero", http.StatusBadRequest)
		return
	}

	regex, err := regexp.Compile(`^[0-9]+$`)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(user.Passport) != 9 || !regex.MatchString(user.Passport) {
		http.Error(w, "Invalid passport number", http.StatusBadRequest)
		return
	}

	_, ok := users[user.ID]
	if ok {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	newUser := calculateTicket(user)

	users[newUser.ID] = newUser

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newUser)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	user, ok := users[id]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func calculateTicket(user User) User {
	currentDate := time.Now().UTC().Truncate(24 * time.Hour)
	user.Ticket = currentDate.AddDate(0, 0, 3)

	if user.Age >= 60 {
		user.Price = 40.0
	} else {
		user.Price = 70.0
	}

	return user
}
