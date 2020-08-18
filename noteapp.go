package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	log "github.com/sirupsen/logrus"
)

// initialize MySQL connection to database using GORM
//Change credentials if needed
var mysqlUser = "root"
var mysqlPassword = "root"
var dbCredentials = mysqlUser + ":" + mysqlPassword + "@/noteapp?charset=utf8&parseTime=True&loc=Local"
var db, _ = gorm.Open("mysql", dbCredentials)

// NoteItemModel : definition of note item
type NoteItemModel struct {
	ID   int    `gorm:"primary_key"`
	Name string `gorm:"type:varchar(512) not null"`
	Body string `gorm:"type:text not null"`
}

// test function that will respond every time it's invoked
func health(w http.ResponseWriter, r *http.Request) {
	log.Info("API Health is OK")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

// adding note to database
func addNote(w http.ResponseWriter, r *http.Request) {

	// if the Content-Type header is present, check that it has the value application/json
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	// to enforce a maximum read of 1MB from the request body
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	dec := json.NewDecoder(r.Body)

	// any extra unexpected fields in the JSON are forbidden now
	dec.DisallowUnknownFields()

	var note NoteItemModel
	err := dec.Decode(&note)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			http.Error(w, msg, http.StatusBadRequest)

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		// otherwise sending 500 server error response
		default:
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// Call decode again, using a pointer to an empty anonymous struct as
	// the destination. If the request body only contained a single JSON
	// object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// add note to db and check
	if createErr := db.Create(&note).Error; createErr != nil {
		msg := fmt.Sprintf("Error: 'name' and 'body' fields can contain max up to 512 and 65535 symbols respectively")
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	result := db.Last(&note)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result.Value)
}

// GetNote : getting note from db by id
func GetNote(w http.ResponseWriter, r *http.Request) {

	// Get URL parameter from mux
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	// Test if the NoteItem exist in DB
	err := GetNoteByID(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"error": "Record Not Found. Make sure that your note id is integer and exists."}`)
	} else {
		log.WithFields(log.Fields{"Id": id}).Info("Getting Note Item")
		note := &NoteItemModel{}
		result := db.First(&note, id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result.Value)
	}
}

// GetNoteByID : check if the item with given id exists in db
func GetNoteByID(ID int) bool {
	note := &NoteItemModel{}
	result := db.First(&note, ID)
	if result.Error != nil {
		log.Warn("Note Item not found in database")
		return false
	}
	return true
}

// this is to set up logrus logger settings
func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

func main() {

	// close database connection when our main() function is returned
	defer db.Close()

	// running automigration against MySQL database immediately after starting our API server
	db.Debug().DropTableIfExists(&NoteItemModel{})
	db.Debug().AutoMigrate(&NoteItemModel{})

	// initialize router
	log.Info("Starting Noteapp API server")
	router := mux.NewRouter()
	router.HandleFunc("/api/health", health).Methods("GET")
	router.HandleFunc("/api/addNote", addNote).Methods("POST")
	router.HandleFunc("/api/note", GetNote).Queries("id", "{id}").Methods("GET")
	http.ListenAndServe(":8888", router)
}
