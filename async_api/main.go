package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type ApplicationAction struct {
	ID            int
	TypeAction    sql.NullString
	Description   string
	ActionID      int
	ApplicationID int
}

var db *sql.DB

func init() {
	var err error
	connectionString := "postgres://lab2:lab2@localhost:8081/rip2?sslmode=disable"
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected to database")
}

func CloseDB() error {
	return db.Close()
}

func MakeAnswer(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	defer r.Body.Close()

	parts := strings.Split(string(body), "=")
	if len(parts) > 1 {
		if parts[1] != "xg12j4" {
			http.Error(w, "Неправильный секретный ключ", http.StatusBadRequest)
			return
		}
	} else {
		fmt.Println("Key not found")
	}
	url := "http://0.0.0.0:8000/actions/process/response"

	pk := r.URL.Query().Get("pk")
	if pk == "" {
		http.Error(w, "Missing 'pk' parameter", http.StatusBadRequest)
		return
	}

	// Преобразование pk в int
	pkInt, err := strconv.Atoi(pk)
	if err != nil {
		http.Error(w, "Invalid 'pk' parameter", http.StatusBadRequest)
		return
	}

	// Отправляем ответ HTTP с кодом состояния 203, указывающим на то, что процесс продолжается
	w.WriteHeader(http.StatusNonAuthoritativeInfo)
	w.Write([]byte("Processing..."))

	// Начинаем работу в новой горутине
	go func() {
		rows, err := db.Query("SELECT * FROM applications_actions WHERE application_id = $1", pkInt)
		if err != nil {
			// Обработка ошибки
			return
		}
		defer rows.Close()

		for rows.Next() {
			var action ApplicationAction
			err := rows.Scan(&action.ID, &action.TypeAction, &action.Description, &action.ActionID, &action.ApplicationID)
			if err != nil {
				// Обработка ошибки
				return
			}
			rand.Seed(time.Now().UnixNano())
			randomNumber := rand.Intn(5) + 1
			time.Sleep(time.Duration(randomNumber) * time.Second)

			tmp := ApplicationAction{
				ID:            action.ID,
				TypeAction:    action.TypeAction,
				Description:   "Какой-то ответ",
				ActionID:      action.ActionID,
				ApplicationID: action.ApplicationID,
			}

			jsonAction, err := json.Marshal(tmp)
			if err != nil {
				// Обработка ошибки
				return
			}
			fmt.Println(string(jsonAction))

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonAction))
			if err != nil {
				// обработка ошибки
				return
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				// обработка ошибки
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				// Обработка ошибки
				return
			}

			fmt.Println("response Body:", string(body))
		}
	}()
}

func main() {
	http.HandleFunc("/makeanswer", MakeAnswer)
	http.ListenAndServe(":8080", nil)
	defer CloseDB()
}
