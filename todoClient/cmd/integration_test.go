package cmd

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func randomTaskName(t *testing.T) string {
	t.Helper()
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	r := rand.New(rand.NewSource(time.Now().Local().UnixNano()))

	var p strings.Builder
	for range 32 {
		p.WriteByte(chars[r.Intn(len(chars))])
	}

	return p.String()
}

func TestIntegration(t *testing.T) {
	apiRoot := "http://localhost:8080"

	if os.Getenv("TODO_API_ROOT") != "" {
		apiRoot = os.Getenv("TODO_API_ROOT")
	}

	//	today := time.Now().Format("02/Jan")
	task := randomTaskName(t)
	//	taskId := ""

	// AddTask
	t.Run("AddTask", func(t *testing.T) {
		args := []string{task}
		expOut := fmt.Sprintf("Добавляем дело %q в список\n", task)

		var out bytes.Buffer
		if err := addAction(&out, apiRoot, args); err != nil {
			t.Fatalf("Не ожидали ошибку, а получили: %q", err)
		}
		if expOut != out.String() {
			t.Errorf("Ожидали получить:\n%q\nа получили:\n%q\n", expOut, out.String())
		}
	})
}
