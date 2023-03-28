package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT")))

	// db, err := sql.Open("mysql", "root:pass1234@tcp(0.0.0.0:3306)/")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Create the database if it does not exist
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS mydatabase")
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = db.Exec("USE mydatabase")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Database created successfully")

	createTable(db)

	router := mux.NewRouter()

	router.HandleFunc("/users", getUsers(db)).Methods("GET")
	router.HandleFunc("/users", createUser(db)).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func createTable(db *sql.DB) {
	// Create the mytable table.

	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS mytableuser (
            id INT PRIMARY KEY AUTO_INCREMENT,
            username VARCHAR(255) NOT NULL,
            email VARCHAR(255) NOT NULL
        );
    `)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("mytable created successfully")
}

func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users := []User{}

		rows, err := db.Query("SELECT id, username, email FROM mytableuser")
		if err != nil {
			log.Println(err)
			http.Error(w, "Error getting users", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var u User
			err := rows.Scan(&u.ID, &u.Username, &u.Email)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error getting users", http.StatusInternalServerError)
				return
			}
			users = append(users, u)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			log.Println(err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		result, err := db.Exec("INSERT INTO mytableuser (username, email) VALUES (?, ?)", u.Username, u.Email)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			log.Println(err)
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}

		u.ID = int(id)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u)
	}
}
