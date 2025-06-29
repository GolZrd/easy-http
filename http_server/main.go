package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// Определяем константы для удобства
const (
	baseUrl       = "localhost:8080"
	createPostfix = "/notes"
	getPostfix    = "/notes/{id}"
)

// Определим модель данных
type NoteInfo struct {
	Title    string `json:"title"`
	Context  string `json:"context"`
	Author   string `json:"author"`
	IsPublic bool   `json:"is_public"`
}

type Note struct {
	ID        int64     `json:"id"`
	Info      NoteInfo  `json:"info"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// У нас будет мапа для хранения заметок
type SyncMap struct {
	elems map[int64]*Note
	mu    sync.RWMutex
}

var notes = &SyncMap{
	elems: make(map[int64]*Note),
}

func main() {
	r := chi.NewRouter()

	r.Post(createPostfix, createNoteHandler)
	r.Get(getPostfix, getNoteHandler)

	log.Printf("Server started on %s", baseUrl)
	err := http.ListenAndServe(baseUrl, r)
	if err != nil {
		log.Fatal(err)
	}
}

// Обработчик создания заметки
func createNoteHandler(w http.ResponseWriter, r *http.Request) {
	info := &NoteInfo{}

	if err := json.NewDecoder(r.Body).Decode(info); err != nil {
		http.Error(w, "Failed to decode note data", http.StatusBadRequest)
		return
	}

	now := time.Now()

	note := &Note{
		ID:        rand.Int63(),
		Info:      *info,
		CreatedAt: now,
		UpdatedAt: now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(note); err != nil {
		http.Error(w, "Failed to encode note data", http.StatusInternalServerError)
		return
	}

	notes.mu.Lock()
	defer notes.mu.Unlock()

	notes.elems[note.ID] = note
}

func getNoteHandler(w http.ResponseWriter, r *http.Request) {
	//Вытаскиваем id из url
	noteID := chi.URLParam(r, "id")
	id, err := parseNoteID(noteID)
	if err != nil {
		http.Error(w, "Failed to parse note id", http.StatusBadRequest)
		return
	}

	notes.mu.RLock()
	defer notes.mu.RUnlock()

	note, ok := notes.elems[id]
	if !ok {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(note); err != nil {
		http.Error(w, "Failed to encode note data", http.StatusInternalServerError)
		return
	}
}

// Функция для преобразования в INT
func parseNoteID(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}
