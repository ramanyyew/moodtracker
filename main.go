package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	Mood      int    `json:"mood"`
	CreatedAt string `json:"created_at"`
}

var (
	tasks  = make(map[int]Task)
	nextID = 1
	mu     sync.Mutex
)

func createTask(w http.ResponseWriter, r *http.Request) {
	var t Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mu.Lock()
	defer mu.Unlock()
	t.ID = nextID
	t.CreatedAt = time.Now().Format(time.RFC3339)
	nextID++
	tasks[t.ID] = t
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func getAllTasks(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	list := []Task{}
	for _, t := range tasks {
		list = append(list, t)
	}
	json.NewEncoder(w).Encode(list)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	var updated Task
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	if _, ok := tasks[id]; !ok {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	updated.ID = id
	updated.CreatedAt = tasks[id].CreatedAt
	tasks[id] = updated
	json.NewEncoder(w).Encode(updated)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	mu.Lock()
	defer mu.Unlock()
	if _, ok := tasks[id]; !ok {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	delete(tasks, id)
	w.WriteHeader(http.StatusNoContent)
}

func moodStats(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	total := 0
	count := 0
	for _, t := range tasks {
		if t.Mood > 0 {
			total += t.Mood
			count++
		}
	}
	avg := 0.0
	if count > 0 {
		avg = float64(total) / float64(count)
	}
	result := map[string]interface{}{
		"average_mood": avg,
		"total_tasks":  len(tasks),
	}
	json.NewEncoder(w).Encode(result)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/tasks", createTask).Methods("POST")
	r.HandleFunc("/tasks", getAllTasks).Methods("GET")
	r.HandleFunc("/tasks/{id}", updateTask).Methods("PUT")
	r.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")
	r.HandleFunc("/stats/mood", moodStats).Methods("GET")

	c := cors.AllowAll()
	handler := c.Handler(r)

	fmt.Println("🚀 Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}