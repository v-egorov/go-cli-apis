package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
)

func TestListAction(t *testing.T) {
	testCases := []struct {
		name     string
		expError error
		expOut   string
		resp     struct {
			Status int
			Body   string
		}
		closeServer bool
	}{
		{
			name:     "Results",
			expError: nil,
			expOut:   "- 1 Task 1\n- 2 Task 2\n",
			resp:     testResp["resultsMany"],
		},
		{
			name:     "NoResults",
			expError: nil,
			resp:     testResp["noResults"],
		},
		{
			name:        "IvalidURL",
			expError:    ErrConnection,
			resp:        testResp["noResults"],
			closeServer: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url, cleanup := mockServer(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.resp.Status)
					fmt.Fprintln(w, tc.resp.Body)
				})
			defer cleanup()

			if tc.closeServer {
				cleanup()
			}

			var out bytes.Buffer
			err := listAction(&out, url)
			if tc.expError != nil {
				if err == nil {
					t.Fatalf("Ожидали ошибку %q, но не получили никакой ошибки", tc.expError)
				}
				if !errors.Is(err, tc.expError) {
					t.Fatalf("Ожидали ошибку %q, а получили %q", tc.expError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Не ожидали ошибку, а получили: %q", err)
			}

			if tc.expOut != out.String() {
				t.Errorf("Ожидали получить %q, а получили: %q", tc.expOut, out.String())
			}
		})
	}
}

func TestViewAction(t *testing.T) {
	testCases := []struct {
		name     string
		expError error
		expOut   string
		resp     struct {
			Status int
			Body   string
		}
		id string
	}{
		{
			name:     "ResultsOne",
			expError: nil,
			expOut: `Дело:		Task 1
Создано:	Oct/28 @08:23
Завершено:	Нет
`,
			resp: testResp["resultsOne"],
			id:   "1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url, cleanup := mockServer(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.resp.Status)
					fmt.Fprintln(w, tc.resp.Body)
				})
			defer cleanup()

			var out bytes.Buffer

			err := viewAction(&out, url, tc.id)
			if tc.expError != nil {
				if err == nil {
					t.Fatalf("Ожидали ошибку: %q, а никакой ошибки не получили", tc.expError)
				}
				if !errors.Is(err, tc.expError) {
					t.Errorf("Ожидали ошибку: %q, а получили: %q", tc.expError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Не ожидали ошибок, а получили: %q", err)
			}

			if out.String() != tc.expOut {
				t.Errorf("Ожидали результат:\n%q, а получили:\n%q", tc.expOut, out.String())
			}
		})
	}
}

func TestAddAction(t *testing.T) {
	expURLPath := "/todo"
	expMethod := http.MethodPost
	expBody := "{\"task\":\"Task 1\"}\n"
	expContentType := "application/json"
	expOut := "Добавляем дело \"Task 1\" в список\n"
	args := []string{"Task", "1"}

	log.Println("TestAddAction: create mock server")
	url, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expURLPath {
			t.Errorf("Ожидали URLPath: %q, получили: %q", expURLPath, r.URL.Path)
		}
		if r.Method != expMethod {
			t.Errorf("Ожидали http method: %q, получили: %q", expMethod, r.Method)
		}

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != expBody {
			t.Errorf("Ожидали body: %q, а получили: %q", expBody, string(body))
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != expContentType {
			t.Errorf("Ожидали Content-Type: %q, а получили: %q", expContentType, contentType)
		}

		w.WriteHeader(testResp["created"].Status)
		fmt.Fprintln(w, testResp["created"].Body)
	})
	defer cleanup()
	log.Printf("TestAddAction: created, apiRoot: %s", url)

	var out bytes.Buffer

	if err := addAction(&out, url, args); err != nil {
		t.Fatalf("Не ожидали ошибку, а получили: %q", err)
	}

	if out.String() != expOut {
		t.Errorf("Ожидали: %q, получили: %q", expOut, out.String())
	}
}

func TestComplteAction(t *testing.T) {
	expURLPath := "/todo/1"
	expMethod := http.MethodPatch
	expQuery := "complete"
	expOut := "Дело 1 отмечено как выполненное\n"
	arg := "1"

	url, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expURLPath {
			t.Errorf("Ожидали путь %q, а получили %q", expURLPath, r.URL.Path)
		}
		if r.Method != expMethod {
			t.Errorf("Ожидали метод %q, а получили %q", expMethod, r.Method)
		}
		if _, ok := r.URL.Query()[expQuery]; !ok {
			t.Errorf("Ожидаемый запрос %q не найден в URL", expQuery)
		}

		w.WriteHeader(testResp["noContent"].Status)
		fmt.Fprintln(w, testResp["noContent"].Body)
	})
	defer cleanup()

	var out bytes.Buffer

	if err := completeAction(&out, url, arg); err != nil {
		t.Fatalf("Не ожидали ошибку, а получили: %q", err)
	}
	if out.String() != expOut {
		t.Errorf("Ожидали вывод:\n%q\nа получили:\n%q", expOut, out.String())
	}
}

func TestDelAction(t *testing.T) {
	expURLPath := "/todo/1"
	expMethod := http.MethodDelete
	expOut := "Дело 1 удалено\n"
	arg := "1"

	log.Println("Creating mockServer")
	url, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Inside mockServer func: %s, %s\n", r.URL.Path, r.Method)
		if r.URL.Path != expURLPath {
			t.Errorf("Ожидали путь %q, а получили %q", expURLPath, r.URL.Path)
		}
		if r.Method != expMethod {
			t.Errorf("Ожидали метод %q, а получили %q", expMethod, r.Method)
		}

		w.WriteHeader(testResp["noContent"].Status)
		fmt.Fprintln(w, testResp["noContent"].Body)
	})
	defer cleanup()
	log.Printf("mockServer: %s\n", url)

	var out bytes.Buffer

	if err := deleteAction(&out, url, arg); err != nil {
		t.Fatalf("Не ожидали получить ошибку, а получили: %q", err)
	}
	if out.String() != expOut {
		t.Errorf("Ожидали вывод:\n%q\nа получили:\n%q", expOut, out.String())
	}
}
