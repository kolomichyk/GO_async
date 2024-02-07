package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	_ "github.com/lib/pq"
)

type ApplicationAction struct {
	ID            int
	TypeAction    string
	Description   string
	ActionID      int
	ApplicationID int
}

func MakeAnswer(w http.ResponseWriter, r *http.Request) {
	url := "http://0.0.0.0:8000/actions/process/response"

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	fmt.Println("Request Body:", string(body))

	// Проверяем, что тело запроса содержит данные
	if len(body) == 0 {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}

	// Теперь мы ожидаем массив объектов ApplicationAction
	var actions []ApplicationAction
	err = json.Unmarshal(body, &actions)
	if err != nil {
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}
	fmt.Println("OK")
	w.WriteHeader(http.StatusNonAuthoritativeInfo)
	w.Write([]byte("Processing..."))

	go func() {
		// Обрабатываем каждый объект в массиве
		for _, action := range actions {
			jsonAction, err := json.Marshal(action)
			if err != nil {
				// Обработка ошибки
				continue
			}
			fmt.Println(string(jsonAction))
			// "xg12j4"
			req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonAction))
			if err != nil {
				// обработка ошибки
				continue
			}
			req.Header.Set("Secret-Key", "xg12j4")
			req.Header.Add("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				// обработка ошибки
				continue
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				// Обработка ошибки
				continue
			}

			fmt.Println("response Body:", string(body))
		}
	}()
}

func main() {
	http.HandleFunc("/makeanswer", MakeAnswer)
	http.ListenAndServe(":8080", nil)
}
