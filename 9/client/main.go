package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	url := os.Getenv("LAB9_API_URL")
	if url == "" {
		url = "http://localhost:8080"
	}
	if err := run(os.Stdin, os.Stdout, NewAPIClient(url)); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}

func ask(reader *bufio.Reader, out io.Writer, prompt string) (string, error) {
	fmt.Fprint(out, prompt)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	if err == io.EOF && line == "" {
		return "", io.EOF
	}
	return strings.TrimSpace(line), nil
}

func askID(reader *bufio.Reader, out io.Writer) (int64, error) {
	text, err := ask(reader, out, "ID пользователя: ")
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseInt(text, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("ID должен быть положительным числом")
	}
	return id, nil
}

func askUser(reader *bufio.Reader, out io.Writer) (UserInput, error) {
	var u UserInput
	var err error
	if u.Name, err = ask(reader, out, "Имя: "); err != nil {
		return u, err
	}
	text, err := ask(reader, out, "Возраст: ")
	if err != nil {
		return u, err
	}
	if u.Age, err = strconv.Atoi(text); err != nil {
		return u, fmt.Errorf("возраст должен быть числом")
	}
	if u.Email, err = ask(reader, out, "Email: "); err != nil {
		return u, err
	}
	return u, nil
}

func printUser(out io.Writer, u User) {
	fmt.Fprintf(out, "ID: %d | Имя: %s | Возраст: %d | Email: %s\n", u.ID, u.Name, u.Age, u.Email)
}

func printUsers(out io.Writer, users UserList) {
	fmt.Fprintf(out, "Найдено: %d (страница %d, лимит %d)\n", users.Total, users.Page, users.Limit)
	if len(users.Users) == 0 {
		fmt.Fprintln(out, "Пользователей нет")
	}
	for _, u := range users.Users {
		printUser(out, u)
	}
}

func run(in io.Reader, out io.Writer, client *APIClient) error {
	reader := bufio.NewReader(in)
	fmt.Fprintln(out, "REST API — клиент")
	for {
		username, err := ask(reader, out, "Логин: ")
		if err != nil {
			return err
		}
		password, err := ask(reader, out, "Пароль: ")
		if err != nil {
			return err
		}
		if err := client.Login(username, password); err != nil {
			fmt.Fprintln(out, "Ошибка авторизации:", err)
			continue
		}
		fmt.Fprintln(out, "Авторизация успешна")
		break
	}

	for {
		fmt.Fprintln(out, "\n1. Список пользователей")
		fmt.Fprintln(out, "2. Пользователь по ID")
		fmt.Fprintln(out, "3. Добавить пользователя")
		fmt.Fprintln(out, "4. Обновить пользователя")
		fmt.Fprintln(out, "5. Удалить пользователя")
		fmt.Fprintln(out, "0. Выход")
		choice, err := ask(reader, out, "Выберите действие: ")
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch choice {
		case "0":
			return nil
		case "1":
			result, err := client.ListUsers()
			if err != nil {
				fmt.Fprintln(out, "Ошибка:", err)
			} else {
				printUsers(out, result)
			}
		case "2":
			id, err := askID(reader, out)
			if err != nil {
				fmt.Fprintln(out, "Ошибка:", err)
				continue
			}
			user, err := client.GetUser(id)
			if err != nil {
				fmt.Fprintln(out, "Ошибка:", err)
			} else {
				printUser(out, user)
			}
		case "3":
			input, err := askUser(reader, out)
			if err != nil {
				fmt.Fprintln(out, "Ошибка:", err)
				continue
			}
			user, err := client.CreateUser(input)
			if err != nil {
				fmt.Fprintln(out, "Ошибка:", err)
			} else {
				fmt.Fprintln(out, "Пользователь создан:")
				printUser(out, user)
			}
		case "4":
			id, err := askID(reader, out)
			if err != nil {
				fmt.Fprintln(out, "Ошибка:", err)
				continue
			}
			input, err := askUser(reader, out)
			if err != nil {
				fmt.Fprintln(out, "Ошибка:", err)
				continue
			}
			user, err := client.UpdateUser(id, input)
			if err != nil {
				fmt.Fprintln(out, "Ошибка:", err)
			} else {
				fmt.Fprintln(out, "Пользователь обновлён:")
				printUser(out, user)
			}
		case "5":
			id, err := askID(reader, out)
			if err != nil {
				fmt.Fprintln(out, "Ошибка:", err)
				continue
			}
			if err := client.DeleteUser(id); err != nil {
				fmt.Fprintln(out, "Ошибка:", err)
			} else {
				fmt.Fprintln(out, "Пользователь удалён")
			}
		default:
			fmt.Fprintln(out, "Неизвестный пункт меню")
		}
	}
}
