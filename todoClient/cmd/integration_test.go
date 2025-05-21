package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
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

	t.Run("ListTasks", func(t *testing.T) {
		var out bytes.Buffer
		if err := listAction(&out, apiRoot); err != nil {
			t.Fatalf("Не ожидали ошибку, а получили: %q", err)
		}

		outList := ""
		scanner := bufio.NewScanner(&out)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), task) {
				outList = scanner.Text()
				break
			}
		}
		if outList == "" {
			t.Errorf("Дело %q не в списке", task)
		}
		log.Printf("Дело: %s\n", outList)

		taskCompleteStatus := strings.Fields(outList)[0]
		log.Printf("taskCompleteStatus: %s\n", taskCompleteStatus)
		if taskCompleteStatus != "-" {
			t.Errorf("Ожидали статус дела: %q, а получили: %q", "-", taskCompleteStatus)
		}

		taskId := strings.Fields(outList)[1]
		log.Printf("taskId: %s", taskId)
	})
}
