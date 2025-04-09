package main

import (
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
			switch r.Method {
			case http.MethodGet:
				getAllHandler(w, r, list)
			case http.MethodPost:
			default:
				message := "HTTP метод не поддерживается"
				replyError(w, r, http.StatusMethodNotAllowed, message)
			}
			return
		}

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
