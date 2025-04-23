package cmd

import (
	"bytes"
	"errors"
	"fmt"
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
			expOut: `Дело: Дело №1
			Создано: 
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
				t.Errorf("Ожидали результат: %q, а получили: %q", tc.expOut, out.String())
			}
		})
	}
}
