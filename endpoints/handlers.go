package endpoints

import (
	"auth_service_template/models"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	uuid "github.com/satori/go.uuid"

	"golang.org/x/crypto/bcrypt"
)

func (r *Router) handleSignin() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		// Parse and decode the request body into a new `Credentials` instance
		creds := &models.Credentials{}
		err := json.NewDecoder(request.Body).Decode(creds)
		if err != nil {
			// If there is something wrong with the request body, return a 400 status
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
			return
		} else if len(creds.Login) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"Error": `Request data not valid, "login" field was empty`})
			return
		}

		if user, err := r.db.FindUserByLoginOrEmail(&creds.Login); err != nil {
			if err == models.ErrUserNotFound {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		} else {
			// Compare the stored hashed password, with the hashed version of the password that was received
			if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
				// If the two passwords don't match, return a 401 status
				w.WriteHeader(http.StatusUnauthorized)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"Error": gocql.ErrNotFound.Error()})
			} else {
				// Create a new random session token
				if uuid, err := uuid.NewV4(); err == nil {
					w.Header().Set("Content-Type", "application/json")
					sessionToken := base64.StdEncoding.EncodeToString(uuid.Bytes())
					if err := r.cache.Put(creds.Login, sessionToken, map[string]interface{}{
						"duration": "120",
					}); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
						return
					}
					// Set session token cookie
					http.SetCookie(w, &http.Cookie{
						Name:    "session_token",
						Value:   sessionToken,
						Expires: time.Now().Add(120 * time.Second),
					})

					w.WriteHeader(http.StatusOK)
					//json.NewEncoder(w).Encode(map[string]models.User{"User": *user})
					json.NewEncoder(w).Encode(map[string]string{"session_token": sessionToken})
					return
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
					return
				}
			}
		}
	}
}

func (r *Router) handleSignup() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		// Parse and decode the request body into a new `Credentials` instance
		creds := &models.User{}
		err := json.NewDecoder(request.Body).Decode(creds)
		if err != nil {
			// If there is something wrong with the request body, return a 400 status
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
			return
		}
		// Salt and hash the password using the bcrypt algorithm
		// The second argument is the cost of hashing, which we arbitrarily set as 16 (this value can be more or less, depending on the computing power you wish to utilize)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 16)

		creds.Password = string(hashedPassword)

		if err := r.db.AddUser(creds); err != nil {
			// If there is any issue with inserting into the database, return a 500 error
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}
}

func (r *Router) handleRefresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// use thing
	}
}

func (r *Router) handleAny() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"Success": true})
		fmt.Println("Success")
	}
}
