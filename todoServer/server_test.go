package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"vegorov.ru/go-cli/todo"
)

func setupAPI(t *testing.T) (string, func()) {
	t.Helper()
	log.Println("setupAPI BEGIN")

	tempTodoFile, err := os.CreateTemp("", "todotest")
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("tempTodoFile: %s\n", tempTodoFile.Name())

	ts := httptest.NewServer(newMux(tempTodoFile.Name()))

	// Добавим тестовые записи
	for i := 1; i < 3; i++ {
		var body bytes.Buffer
		taskName := fmt.Sprintf("Дело № %d.", i)
		item := struct {
			Task string `json:"task"`
		}{
			Task: taskName,
		}

		log.Printf("Encode new item: %s\n", item.Task)
		if err := json.NewEncoder(&body).Encode(item); err != nil {
			t.Fatal(err)
		}

		log.Printf("Issue POST to: %s\n", ts.URL+"/todo")
		r, err := http.Post(ts.URL+"/todo", "application/json", &body)
		if err != nil {
			t.Fatal(err)
		}

		log.Printf("Response status code: %d\n", r.StatusCode)
		if r.StatusCode != http.StatusCreated {
			t.Fatalf("Не удалось создать тестовые дела: httpStatus: %d", r.StatusCode)
		}
	}

	log.Println("setupAPI END")
	return ts.URL, func() {
		ts.Close()
		os.Remove(tempTodoFile.Name())
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name       string
		path       string
		expCode    int
		expItems   int
		expContent string
	}{
		{
			name:       "GetRoot",
			path:       "/",
			expCode:    http.StatusOK,
			expContent: "There's an API here\n",
		},
		{
			name:       "GetAll",
			path:       "/todo/",
			expCode:    http.StatusOK,
			expItems:   2,
			expContent: "Дело № 1.",
		},
		{
			name:       "GetOne",
			path:       "/todo/1",
			expCode:    http.StatusOK,
			expItems:   1,
			expContent: "Дело № 1.",
		},
		{
			name:    "NotFound",
			path:    "/todo/123",
			expCode: http.StatusNotFound,
		},
	}

	// log.SetOutput(io.Discard)

	url, cleanup := setupAPI(t)
	defer cleanup()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				resp struct {
					Results      todo.List `json:"results"`
					Date         int64     `json:"date"`
					TotalResults int       `json:"total_results"`
				}
				body []byte
				err  error
			)

			r, err := http.Get(url + tc.path)
			if err != nil {
				t.Error(err)
			}
			defer r.Body.Close()

			if r.StatusCode != tc.expCode {
				t.Fatalf("Ожидали HTTP Status: %q, а получили: %q", tc.expCode, r.StatusCode)
			}

			switch {
			case r.Header.Get("Content-Type") == "application/json":
				if err = json.NewDecoder(r.Body).Decode(&resp); err != nil {
					t.Error(err)
				}
				if resp.TotalResults != tc.expItems {
					t.Errorf("Ожидали %d элементов, а получили %d", tc.expItems, resp.TotalResults)
				}
				if resp.Results[0].Task != tc.expContent {
					t.Errorf("Ожидали %q, а получили %q\n", tc.expContent, resp.Results[0].Task)
				}
			case strings.Contains(r.Header.Get("Content-Type"), "text/plain"):
				if body, err = io.ReadAll(r.Body); err != nil {
					t.Error(err)
				}
				if !strings.Contains(string(body), tc.expContent) {
					t.Errorf("Ожидали:\n%q,\nполучили:\n%q", tc.expContent, string(body))
				}
			default:
				t.Fatalf("Неподдерживаемый Content-Type: %q", r.Header.Get("Content-Type"))
			}
		})
	}
}

func TestAdd(t *testing.T) {
	url, cleanup := setupAPI(t)
	defer cleanup()

	taskName := "Task N3"
	t.Run("Add", func(t *testing.T) {
		var body bytes.Buffer
		item := struct {
			Task string `json:"task"`
		}{
			Task: taskName,
		}

		if err := json.NewEncoder(&body).Encode(item); err != nil {
			t.Fatal(err)
		}

		r, err := http.Post(url+"/todo", "application/json", &body)
		if err != nil {
			t.Fatal(err)
		}

		if r.StatusCode != http.StatusCreated {
			t.Errorf("Ожидали httpStatus %q, а получили %q",
				http.StatusText(http.StatusCreated), http.StatusText(r.StatusCode))
		}
	})

	t.Run("CheckAdd", func(t *testing.T) {
		r, err := http.Get(url + "/todo/3")
		if err != nil {
			t.Error(err)
		}

		if r.StatusCode != http.StatusOK {
			t.Fatalf("Ожидали httpStatus %q, а получили %q",
				http.StatusText(http.StatusOK), http.StatusText(r.StatusCode))
		}

		var resp todoResponse
		if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		r.Body.Close()

		if resp.Results[0].Task != taskName {
			t.Errorf("Ожидали Task %q, а получили %q", taskName, resp.Results[0].Task)
		}
	})
}
