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
	Completed *bool  `json:"completed"`
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
	http.HandleFunc("/update-task", updateTask)
	http.HandleFunc("/delete-task", deleteTask)
	http.HandleFunc("/new-user", newUser)
	http.HandleFunc("/delete-user", deleteUser)
	http.HandleFunc("/auth", userAuth)
	http.ListenAndServe(":8080", nil)
}

func userAuth(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var dbUser User
	err = db.QueryRow("SELECT id, username, password FROM users WHERE username = ? AND password = ?", user.Username, user.Password).Scan(&dbUser.ID, &dbUser.Username, &dbUser.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Incorret login or password", http.StatusUnauthorized)

		} else {
			http.Error(w, "Error while checking user credentials", http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successful login"})

}

func newUser(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var existingUsername string
	err = db.QueryRow("SELECT username FROM users WHERE username = ?", user.Username).Scan(&existingUsername)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Error while checking user credentials", http.StatusInternalServerError)
		return
	}
	if existingUsername != "" {
		http.Error(w, "User with this username already exists", http.StatusConflict)
		return
	}

	db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, user.Password)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "New user added successfully"})
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		http.Error(w, "Error deleteing task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task deleted successfully"})

}

func updateTask(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	var updatedTask Tasks
	err = json.NewDecoder(r.Body).Decode(&updatedTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := "UPDATE tasks SET "
	params := []interface{}{}

	if updatedTask.Title != "" {
		query += "title = ?, "
		params = append(params, updatedTask.Title)
	}

	if updatedTask.Completed != nil {
		query += "completed = ?, "
		params = append(params, updatedTask.Completed)
	}

	if len(params) > 0 {
		query = query[:len(query)-2]
		query += " WHERE id = ?"
		params = append(params, taskID)

		_, err = db.Exec(query, params...)
		if err != nil {
			http.Error(w, "Error updating task", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Task updated successfully"})
	} else {
		http.Error(w, "No fields to update", http.StatusBadRequest)
	}

}

func createTask(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var task Tasks
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db.Exec("INSERT INTO tasks (title, completed, user_id) VALUES (?, ?, ?)", task.Title, task.Completed, task.UserID)

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
