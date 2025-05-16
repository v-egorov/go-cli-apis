package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var (
	ErrConnection      = errors.New("ошибка соединения")
	ErrNotFound        = errors.New("не найдено")
	ErrInvalidResponse = errors.New("невалидный ответ сервера")
	ErrInvalid         = errors.New("невалидные данные")
	ErrNotNumber       = errors.New("неверное число")
)

type item struct {
	Task        string
	Done        bool
	CreatedAt   time.Time
	CompletedAt time.Time
}

type response struct {
	Results      []item `json:"results"`
	Date         int    `json:"date"`
	TotalResults int    `json:"total_results"`
}

func newClient() *http.Client {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}
	return c
}

func getItems(url string) ([]item, error) {
	r, err := newClient().Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConnection, err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("невозможно прочитать body ответа: %w", err)
		}
		err = ErrInvalidResponse
		if r.StatusCode == http.StatusNotFound {
			err = ErrNotFound
		}
		return nil, fmt.Errorf("%w: %s", err, msg)
	}

	var resp response

	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if resp.TotalResults == 0 {
		log.Println("Zero records in response")
		return nil, fmt.Errorf("%w: запрос вернул 0 записей", ErrNotFound)
	}

	return resp.Results, nil
}

func getAll(apiRoot string) ([]item, error) {
	log.Println("Client: getAll")
	u := fmt.Sprintf("%s/todo", apiRoot)
	return getItems(u)
}

func getOne(apiRoot string, id int) (item, error) {
	log.Println("Client: getOne")
	u := fmt.Sprintf("%s/todo/%d", apiRoot, id)

	items, err := getItems(u)
	if err != nil {
		log.Printf("Error: %s", err)
		return item{}, err
	}

	if len(items) != 1 {
		log.Println("Error: getItems returned more then one record")
		return item{}, fmt.Errorf("%w: Невалидный результат - получили более 1 записи", ErrInvalid)
	}

	log.Println("Clent: getOne ok")
	return items[0], nil
}

func sendRequest(url, method, contentType string, expStatus int, body io.Reader) error {
	log.Println("Client: sendRequest")
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("content-type", contentType)
	}

	log.Println("Client: Do")
	r, err := newClient().Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != expStatus {
		msg, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("невозможно прочитать body ответа: %w", err)
		}
		err = ErrInvalidResponse
		if r.StatusCode == http.StatusNotFound {
			err = ErrNotFound
		}
		return fmt.Errorf("%w: %s", err, msg)
	}
	return nil
}

func addItem(apiRot, task string) error {
	log.Println("Client: addItem")
	u := fmt.Sprintf("%s/todo", apiRot)

	item := struct {
		Task string `json:"task"`
	}{
		Task: task,
	}

	var body bytes.Buffer

	log.Println("addItem: encode request body")
	if err := json.NewEncoder(&body).Encode(item); err != nil {
		return err
	}

	log.Println("addItem: sendRequest")
	return sendRequest(u, http.MethodPost, "application/json", http.StatusCreated, &body)
}

func completeItem(apiRoot string, id int) error {
	log.Println("Client: completeItem")
	u := fmt.Sprintf("%s/todo/%d?complete", apiRoot, id)

	log.Printf("PATCH: %s\n", u)
	return sendRequest(u, http.MethodPatch, "", http.StatusNoContent, nil)
}

func deteleItem(apiRoot string, id int) error {
	log.Println("Client: deleteItem")
	u := fmt.Sprintf("%s/todo/%d", apiRoot, id)
	log.Printf("DELETE: %s\n", u)
	return sendRequest(u, http.MethodDelete, "", http.StatusNoContent, nil)
}
