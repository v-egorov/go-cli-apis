package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"vegorov.ru/go-cli/todo"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrInvalidData = errors.New("invalid data")
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("rootHandler: %s %s\n", r.URL, r.Method)
	if r.URL.Path != "/" {
		log.Printf("Not found: %s\n", r.URL)
		replyError(w, r, http.StatusNotFound, "")
		return
	}
	content := "There's an API here\n"
	replyTextContent(w, r, http.StatusOK, content)
}

func todoRouter(todoFile string, l sync.Locker) http.HandlerFunc {
	log.Printf("Creating todoRouter serving todoFile: %s\n", todoFile)
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("todoRouter start: r.URL: %s, todoFile: %s\n", r.URL, todoFile)
		list := &todo.List{}

		l.Lock()
		defer l.Unlock()

		log.Printf("Reading todoFile: %s\n", todoFile)
		if err := list.Get(todoFile); err != nil {
			replyError(w, r, http.StatusInternalServerError, err.Error())
			return
		}

		if r.URL.Path == "" {
			// это /todo/ без параметров
			switch r.Method {
			case http.MethodGet:
				getAllHandler(w, r, list)
			case http.MethodPost:
				addHandler(w, r, list, todoFile)
			default:
				message := "HTTP метод не поддерживается"
				replyError(w, r, http.StatusMethodNotAllowed, message)
			}
			// Завершили обработку без запроса по путти /todo/ без параметров
			return
		}

		// Сюда попадаем только в том случае, если URL запроса содержит параметры
		id, err := validateID(r.URL.Path, list)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				replyError(w, r, http.StatusNotFound, err.Error())
				return
			}
			replyError(w, r, http.StatusBadRequest, err.Error())
			return
		}

		switch r.Method {
		case http.MethodGet:
			getOneHandler(w, r, list, id)
		case http.MethodDelete:
			deleteHandler(w, r, list, id, todoFile)
		case http.MethodPatch:
			patchHandler(w, r, list, id, todoFile)
		default:
			message := "Метод не поддерживается"
			replyError(w, r, http.StatusMethodNotAllowed, message)
		}
	}
}

func validateID(path string, list *todo.List) (int, error) {
	log.Printf("validateID: %s\n", path)
	id, err := strconv.Atoi(path)
	if err != nil {
		return 0, fmt.Errorf("%w: невалидный ID: %s", ErrInvalidData, err)
	}
	if id < 1 {
		return 0, fmt.Errorf("%w: невалидный ID, меньше 1: %d", ErrInvalidData, id)
	}
	if id > len(*list) {
		return id, fmt.Errorf("%w: ID: %d не найден", ErrNotFound, id)
	}
	return id, nil
}

func getAllHandler(w http.ResponseWriter, r *http.Request, list *todo.List) {
	log.Printf("getAllHandler: %s", r.URL)
	resp := &todoResponse{
		Results: *list,
	}
	replyJSONContent(w, r, http.StatusOK, resp)
}

func getOneHandler(w http.ResponseWriter, r *http.Request, list *todo.List, id int) {
	log.Printf("getOneHandler: %s", r.URL)
	resp := &todoResponse{
		Results: (*list)[id-1 : id],
	}
	replyJSONContent(w, r, http.StatusOK, resp)
}

func addHandler(w http.ResponseWriter, r *http.Request, list *todo.List, todoFile string) {
	log.Printf("addHandler: todoFile: %s\n", todoFile)
	item := struct {
		Task string `json:"task"`
	}{}

	log.Println("Decode JSON from request")
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		message := fmt.Sprintf("Невалидный JSON: %s", err)
		replyError(w, r, http.StatusBadRequest, message)
		return
	}

	list.Add(item.Task)
	if err := list.Save(todoFile); err != nil {
		replyError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	replyTextContent(w, r, http.StatusCreated, "")
}

func deleteHandler(w http.ResponseWriter, r *http.Request, list *todo.List, id int, todoFile string) {
	log.Printf("deleteHandler: id: %d, todoFile: %s\n", id, todoFile)
	list.Delete(id)
	if err := list.Save(todoFile); err != nil {
		log.Printf("Error in list.Save: %s\n", err.Error())
		replyError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	replyTextContent(w, r, http.StatusNoContent, "")
}

func patchHandler(w http.ResponseWriter, r *http.Request, list *todo.List, id int, todoFile string) {
	log.Printf("patchHandler: id %d, todoFile: %s\n", id, todoFile)

	q := r.URL.Query()
	if _, ok := q["complete"]; !ok {
		message := "Отсутствует параметр запроса 'complete'"
		replyError(w, r, http.StatusBadRequest, message)
		return
	}

	if err := list.Complete(id); err != nil {
		replyError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	if err := list.Save(todoFile); err != nil {
		replyError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	replyTextContent(w, r, http.StatusNoContent, "")
}
