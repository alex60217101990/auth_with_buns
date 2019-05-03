package endpoints

import (
	"auth_service_template/models"
	"encoding/json"
	"net/http"
)

func (r *Router) middlewareSignin(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse and decode the request body into a new `Credentials` instance
		creds := &models.User{}
		err := json.NewDecoder(r.Body).Decode(creds)
		if err != nil {
			// If there is something wrong with the request body, return a 400 status
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(err)
			return
		}
		// // Get the existing entry present in the database for the given username
		// result := db.QueryRow("select password from users where username=$1", creds.Username)
		// if err != nil {
		// 	// If there is an issue with the database, return a 500 error
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }
		// // We create another instance of `Credentials` to store the credentials we get from the database
		// storedCreds := &Credentials{}
		// // Store the obtained password in `storedCreds`
		// err = result.Scan(&storedCreds.Password)
		// if err != nil {
		// 	// If an entry with the username does not exist, send an "Unauthorized"(401) status
		// 	if err == sql.ErrNoRows {
		// 		w.WriteHeader(http.StatusUnauthorized)
		// 		return
		// 	}
		// 	// If the error is of any other type, send a 500 status
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		// // Compare the stored hashed password, with the hashed version of the password that was received
		// if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
		// 	// If the two passwords don't match, return a 401 status
		// 	w.WriteHeader(http.StatusUnauthorized)
		// }

		// // If we reach this point, that means the users password was correct, and that they are authorized
		// // The default 200 status is sent

		// log.Println(r.Method, "-", r.RequestURI)
		// cookie, _ := r.Cookie("username")
		// if cookie != nil {
		// 	//Add data to context
		// 	ctx := context.WithValue(r.Context(), "Username", cookie.Value)
		// 	next.ServeHTTP(w, r.WithContext(ctx))
		// } else {
		// 	next.ServeHTTP(w, r)
		// }
		h(w, r)
	})
}
