package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Tasks struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	UserID    int    `json:"user_id"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	initDB()
	http.HandleFunc("/tasks", getTasks)
	http.HandleFunc("/create-task", createTask)
	http.ListenAndServe(":8080", nil)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	title := r.FormValue("title")
	completed := r.FormValue("completed")
	userID := r.FormValue("user_id")

	db.Exec("INSERT INTO tasks (title, completed, user_id) VALUES (?, ?, ?)", title, completed, userID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task created successfully"})
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userID := r.URL.Query().Get("user_id")

	rows, err := db.Query("SELECT * FROM tasks WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Tasks
	for rows.Next() {
		var task Tasks
		err := rows.Scan(&task.ID, &task.Title, &task.Completed, &task.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func initDB() {
	db, err := sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		password TEXT NOT NULL
	);
	`

	_, err = db.Exec(createUsersTable)
	if err != nil {
		log.Fatal(err)
	}

	createTasksTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed BOOLEAN DEFAULT FALSE,
		user_id INTEGER,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	_, err = db.Exec(createTasksTable)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database initialized successfully")
}
