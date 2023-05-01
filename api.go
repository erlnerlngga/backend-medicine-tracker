package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type APIServer struct {
	listenAddr string
	store Storage
}

func NewApiServer(listenAddr string, storage Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store: storage,
	}
} 

func (s *APIServer) Run() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	
	router.Post("/register", makeHTTPHandleFunc(s.handleCreateAccount))
	router.Post("/login", makeHTTPHandleFunc(s.handleLogIn))
	router.Get("/login/{token}", makeHTTPHandleFunc(s.handleVerifyLogin))

	router.Group(func(r chi.Router) {
		r.Use(WithJWTAuth)

		r.Get("/logout", makeHTTPHandleFunc(s.handleLogout))
		
		r.Post("/medicine", makeHTTPHandleFunc(s.handleCreateMedicine))
		r.Get("/medicine/{userId}", makeHTTPHandleFunc(s.handleGetAllMedicine))
		r.Put("/medicine", makeHTTPHandleFunc(s.handleUpdateMedicine))
		r.Delete("/medicine/{id}", makeHTTPHandleFunc(s.handleDeleteMedicine))
	})

	log.Println("Server running in Port:", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

// LOG IN
func (s *APIServer) handleLogIn(w http.ResponseWriter, r *http.Request) error {
	email := new(LogInType)
	if err:= json.NewDecoder(r.Body).Decode(email); err != nil {
		return nil
	}

	account, err := s.store.CheckEmail(email.Email)
	if err != nil {
		return err
	}

	// create JWT token
	tokenStr, err := CreateJWT(account.ID)
	if err != nil {
		return err
	}

	if err := sendEmailLogin(account.Email, tokenStr); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleVerifyLogin(w http.ResponseWriter, r *http.Request) error {
	tokenStr := chi.URLParam(r, "token")

	claims := new(ClaimsType)

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Signature Invalid"})
		}

		return WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
	}

	if !token.Valid {
		return WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "token invalid"})
	}

	http.SetCookie(w, &http.Cookie{
		Name: "token",
		Value: tokenStr,
		Expires: claims.ExpiresAt.Time,
		Domain: "http://localhost:8080",
		Path: "/",
	})

	return WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// LOGOUT
func (s *APIServer) handleLogout(w http.ResponseWriter, r *http.Request) error {
	http.SetCookie(w, &http.Cookie{
		Name: "token",
		Value: "",
		Expires: time.Now(),
		MaxAge: -1,
	})

	return WriteJSON(w, http.StatusOK, map[string]string{"status": "Logout Success"})
} 

// Create Account
func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	newAccount := new(CreateNewAccountType)

	if err := json.NewDecoder(r.Body).Decode(newAccount); err != nil {
		return err
	}

	res, err := s.store.CreateAccount(newAccount); 

	if err != nil {
		return err
	}

	// create jwt 
	tokenStr, err := CreateJWT(res.ID)
	if err != nil {
		return err
	}

	if err := sendEmailLogin(res.Email, tokenStr); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, res)
}

// Add Medicine
func (s *APIServer) handleCreateMedicine(w http.ResponseWriter, r *http.Request) error {
	medicineData := new(MedicineRequest)

	if err := json.NewDecoder(r.Body).Decode(medicineData); err != nil {
		return err
	}

	defer r.Body.Close()

	medicine := NewMedicine(medicineData.Name, medicineData.Dosage, medicineData.Frequency, medicineData.User_ID)
	if err := s.store.CreateMedicine(medicine); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, medicine)
}

// Get All Medicine
func (s *APIServer) handleGetAllMedicine(w http.ResponseWriter, r *http.Request) error {
	idParam := chi.URLParam(r, "userId")
	medicines, err := s.store.GetAllMedicine(idParam)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, medicines)
}

// Update Medicine
func (s *APIServer) handleUpdateMedicine(w http.ResponseWriter, r *http.Request) error {
	updateMedicine := new(UpdateRequest)

	if err := json.NewDecoder(r.Body).Decode(updateMedicine); err != nil {
		return nil
	}

	updateMed := UpdateOldMedicine(updateMedicine.ID, updateMedicine.Name, updateMedicine.Dosage, updateMedicine.Frequency)
	if err := s.store.UpdateMedicine(updateMed); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, updateMed)
}

// Delete Medicine
func (s *APIServer) handleDeleteMedicine(w http.ResponseWriter, r *http.Request) error {
	idParam := chi.URLParam(r, "id")

	if err := s.store.DeleteMedicine(idParam); err != nil {
		return err
	}

	return nil
}

// CREATE JWT TOKEN
func CreateJWT(user_id string) (string, error) {
	// Declare expiration time, with 5 minutes
	expirationTime := time.Now().Add(time.Hour*24)

	// create JWT claims 
	claims := &ClaimsType{
		User_ID: user_id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Declare token 
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create jwt string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}


	return tokenString, nil
}

// MIDDLEWARE TO HANDLE JWT VERIFICATION
func WithJWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")

		if err != nil {
			if err == http.ErrNoCookie {
				// if cookie is not yet set
				WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "No Cookie"})
				return
			}

			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			return
		}

		// get token from jwt
		tokenString := c.Value

		// initialize init
		claims := new(ClaimsType)

		// Parse the JWT string and store the result in `claims`.
		// Note that we are passing the key in this method as well. This method will return an error
		// if the token is invalid (if it has expired according to the expiry time we set on sign in),
		// or if the signature does not match

		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Signature Invalid"})
				return
			}

			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			return
		}

		if !token.Valid {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "token invalid"})
			return
		}

		// check is expiration time is end? if end create new one or updated
		if time.Until(claims.ExpiresAt.Time) <= 30*time.Second && time.Until(claims.ExpiresAt.Time) > 0*time.Second {
			expirationTime := time.Now().Add(24*time.Hour)
			claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			newTokenString, err := token.SignedString(jwtKey)
			if err != nil {
				WriteJSON(w, http.StatusInternalServerError, ApiError{Error: err.Error()})
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name: "token",
				Value: newTokenString,
				Expires: expirationTime,
			})
		}

		next.ServeHTTP(w, r)
	})
}

// FUNCTION HELPER
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

// SEND EMAIL 
func sendEmailLogin(email, token string) error {
	// set email auth
	authEmail := smtp.PlainAuth("", os.Getenv("EMAIL"), os.Getenv("EMAIL_PASSWORD"), "smtp.gmail.com")

	// compose email
	to := []string{email}
	msg := []byte("To: " + email + "\r\n" + "Subject: Login Link\r\n" + "\r\n" + "http://127.0.0.1:8080/login/" + token)

	if err := smtp.SendMail("smtp.gmail.com:587", authEmail, os.Getenv("EMAIL"), to, msg); err != nil {
		return err
	}

	return nil
}


