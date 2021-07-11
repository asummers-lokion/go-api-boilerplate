package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"text/template"

	smtpLoginAuth "go-api-boilerplate/loginAuth"

	"github.com/joho/godotenv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type server struct{}

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "GET called"}`))
}

type ContactForm struct {
	Message string
}

func post(w http.ResponseWriter, r *http.Request) {
	enverr := godotenv.Load()
	if enverr != nil {
		log.Fatal("Could not load .env file")
	}

	var contactForm ContactForm

	structerr := json.NewDecoder(r.Body).Decode(&contactForm)
	if structerr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(structerr.Error())
		w.Write([]byte(`{"success": false}`))
		return
	}

	//Auth/sender data
	from := os.Getenv("smtpUsername")
	password := os.Getenv("smtpPassword")

	//Recipient
	to := []string{
		"email@mailinator.com",
	}

	//Server config
	smtpHost := os.Getenv("smtpHost")
	smtpPort := os.Getenv("smtpPort")

	//Template
	template, _ := template.ParseFiles("emailTemplates/email.html")
	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("Subject: Test from GO \n%s\n\n", mimeHeaders)))
	template.Execute(&body, struct {
		Message string
	}{
		Message: contactForm.Message,
	})

	auth := smtpLoginAuth.LoginAuth(from, password)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		w.Write([]byte(`{"success": false}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"success": true}`))
}

func put(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message": "put called"}`))
}

func delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "delete called"}`))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "not found"}`))
}

//Sample route with path parameters
func params(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	userID := -1
	var err error
	if val, ok := pathParams["userID"]; ok {
		userID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "Need a number"}`))
			return
		}
	}

	commentID := -1
	if val, ok := pathParams["commentID"]; ok {
		commentID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}

	query := r.URL.Query()
	location := query.Get("location")

	w.Write([]byte(fmt.Sprintf(`{"userID": %d, "commentID": %d, "location": "%s" }`, userID, commentID, location)))
}

func GetPort() string {
	var port = os.Getenv("PORT")
	if port == "" {
		port = "5000"
		fmt.Println("No port envrionment variable, using port 5000")
	}
	return ":" + port
}

func main() {

	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/", get).Methods(http.MethodGet)
	api.HandleFunc("/", post).Methods(http.MethodPost, "OPTIONS")
	api.HandleFunc("/", put).Methods(http.MethodPut)
	api.HandleFunc("/", delete).Methods(http.MethodDelete)
	api.HandleFunc("/", notFound)

	//Sample route with query parameters
	api.HandleFunc("/user/{userID}/comment/{commentID}", params).Methods(http.MethodGet)

	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"POST", "OPTIONS"})

	//No CORS: log.Fatal(http.ListenAndServe(GetPort(), r))
	log.Fatal(http.ListenAndServe(GetPort(), handlers.CORS(allowedHeaders, allowedMethods, allowedOrigins)(r)))
}
