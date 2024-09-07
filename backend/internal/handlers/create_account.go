package handlers

import (
	"encoding/json"
	"net/http"
	"shortly/internal/models"
	"shortly/internal/utils"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// creates new user then logs them in
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	log.WithFields(log.Fields{
		"method": r.Method,
		"url":    r.URL.Path,
		"remote": r.RemoteAddr,
	}).Info("Create user request received")

	// parse req body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"remote":  r.RemoteAddr,
			"payload": req,
		}).Warn("Invalid input for create user")
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err := utils.FindUserByUsername(req.Username)
	if err == nil {
		log.WithFields(log.Fields{
			"username": req.Username,
			"remote":   r.RemoteAddr,
		}).Warn("Attempted to create account with existing username")
		http.Error(w, "Username already exists", http.StatusBadRequest)
		return
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err.Error(),
			"username": req.Username,
		}).Error("Failed to hash password")
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// create user and save into db
	user := models.User{
		Username:     req.Username,
		PasswordHash: string(hashed),
	}

	if err := utils.SaveUser(&user); err != nil {
		log.WithFields(log.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Error("Failed to save user in the database")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"username": user.Username,
	}).Info("User account created successfully")

	// log user in after account creation
	if err := utils.Authenticate(w, req.Username, req.Password); err != nil {
		log.WithFields(log.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Error("Failed to log user in after account creation")
		http.Error(w, "Failed to log in after account creation", http.StatusInternalServerError)
		return
	}

	log.WithFields(log.Fields{
		"username": user.Username,
	}).Info("User logged in successfully after account creation")

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Account created and logged in successfully"))
}
