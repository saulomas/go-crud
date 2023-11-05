package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

type User struct {
	Id    int
	Name  string
	Email string
	Age   int
}

func Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	u := User{}
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		fmt.Println("Server failed to handle", err)
		return
	}

	_, err = db.Exec("INSERT INTO users (name, email, age) VALUES ($1, $2, $3)", u.Name, u.Email, u.Age)
	if err != nil {
		fmt.Println("Server failed to handle", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func Read(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		fmt.Println("Server failed to handle", err)
		return
	}

	defer rows.Close()
	data := make([]User, 0)

	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.Age)
		if err != nil {
			fmt.Println("Server failed to handle", err)
			return
		}
		data = append(data, user)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("Server failed to handle", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	up := User{}
	err := json.NewDecoder(r.Body).Decode(&up)
	if err != nil {
		fmt.Println("Server failed to handle", err)
		return
	}

	row := db.QueryRow("SELECT * FROM users WHERE id = $1", id)
	u := User{}
	err = row.Scan(&u.Id, &u.Name, &u.Email, &u.Age)

	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
	}

	if up.Name != "" {
		u.Name = up.Name
	}

	if up.Email != "" {
		u.Email = up.Email
	}

	if up.Age != 0 {
		u.Age = up.Age
	}

	_, err = db.Exec("UPDATE users SET name = $1, email = $2, age = $3 WHERE id = $4", u.Name, u.Email, u.Age, u.Id)
	if err != nil {
		fmt.Println("Server failed to handle", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")

	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		fmt.Println("Server failed to handle", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://saulo:saulomas@postgres/crud?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("You connected to your database.")
}

func main() {
	http.HandleFunc("/users/create", Create)
	http.HandleFunc("/users/read", Read)
	http.HandleFunc("/users/update", Update)
	http.HandleFunc("/users/delete", Delete)
	http.ListenAndServe(":8080", nil)
}
