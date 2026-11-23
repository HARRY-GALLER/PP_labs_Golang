package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	url := os.Getenv("API_URL")
	if url == "" {
		url = "http://localhost:8080"
	}
	client := NewClient(url)
	console := bufio.NewReader(os.Stdin)
	fmt.Println("Клиент для Users API. Сервер:", url)

	for {
		fmt.Println("\n=== МЕНЮ ===")
		fmt.Println("1. Регистрация")
		fmt.Println("2. Войти")
		fmt.Println("3. Список пользователей")
		fmt.Println("4. Найти пользователя по ID")
		fmt.Println("5. Добавить пользователя")
		fmt.Println("6. Изменить пользователя")
		fmt.Println("7. Удалить пользователя")
		fmt.Println("8. Выйти из учетной записи")
		fmt.Println("0. Закрыть программу")
		choice := read(console, "Выберите пункт: ")
		if choice == "0" {
			fmt.Println("Работа завершена")
			return
		}
		if choice != "1" && choice != "2" && client.token == "" {
			fmt.Println("Сначала войдите в учетную запись или зарегистрируйтесь")
			continue
		}
		if err := runAction(client, console, choice); err != nil {
			fmt.Println("Ошибка:", err)
		}
	}
}

func runAction(c *Client, console *bufio.Reader, choice string) error {
	switch choice {
	case "1":
		in, err := inputUser(console, true)
		if err != nil {
			return err
		}
		u, err := c.Register(in)
		if err != nil {
			return err
		}
		fmt.Printf("Зарегистрирован пользователь ID=%d. Теперь войдите через пункт 2.\n", u.ID)
	case "2":
		email := read(console, "Email: ")
		password := read(console, "Пароль: ")
		res, err := c.Login(email, password)
		if err != nil {
			return err
		}
		fmt.Printf("Вход выполнен. Ваш ID: %d, роль: %s, токен действует до: %s\n",
			res.UserID, res.Role, res.ExpiresAt.Local().Format("02.01.2006 15:04"))
	case "3":
		pageText := read(console, "Номер страницы (Enter = 1): ")
		if pageText == "" {
			pageText = "1"
		}
		page, err := strconv.Atoi(pageText)
		if err != nil || page < 1 {
			return fmt.Errorf("страница должна быть положительным числом")
		}
		result, err := c.List(page)
		if err != nil {
			return err
		}
		fmt.Printf("Всего: %d, страница %d из %d\n", result.Meta.Total, result.Meta.Page, result.Meta.TotalPages)
		for _, u := range result.Data {
			printUser(u)
		}
		if len(result.Data) == 0 {
			fmt.Println("Пользователи не найдены")
		}
	case "4":
		id, err := inputID(console)
		if err != nil {
			return err
		}
		u, err := c.Get(id)
		if err != nil {
			return err
		}
		printUser(u)
	case "5":
		in, err := inputUser(console, true)
		if err != nil {
			return err
		}
		u, err := c.Create(in)
		if err != nil {
			return err
		}
		fmt.Println("Пользователь создан:")
		printUser(u)
	case "6":
		id, err := inputID(console)
		if err != nil {
			return err
		}
		fmt.Println("Введите новые данные пользователя (все три поля обязательны)")
		in, err := inputUser(console, false)
		if err != nil {
			return err
		}
		u, err := c.Update(id, in)
		if err != nil {
			return err
		}
		fmt.Println("Пользователь обновлен:")
		printUser(u)
	case "7":
		id, err := inputID(console)
		if err != nil {
			return err
		}
		if read(console, "Подтвердите удаление (да/нет): ") != "да" {
			fmt.Println("Отменено")
			return nil
		}
		if err := c.Delete(id); err != nil {
			return err
		}
		fmt.Println("Пользователь удален")
	case "8":
		if err := c.Logout(); err != nil {
			return err
		}
		fmt.Println("Выход из учетной записи выполнен")
	default:
		fmt.Println("Такого пункта нет")
	}
	return nil
}

func read(r *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func inputID(r *bufio.Reader) (int64, error) {
	id, err := strconv.ParseInt(read(r, "ID пользователя: "), 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("ID должен быть положительным целым числом")
	}
	return id, nil
}

func inputUser(r *bufio.Reader, withPassword bool) (UserInput, error) {
	name := read(r, "Имя: ")
	email := read(r, "Email: ")
	age, err := strconv.Atoi(read(r, "Возраст: "))
	if err != nil || age < 1 || age > 150 {
		return UserInput{}, fmt.Errorf("возраст должен быть от 1 до 150")
	}
	in := UserInput{Name: name, Email: email, Age: age}
	if withPassword {
		in.Password = read(r, "Пароль (минимум 8 символов): ")
		if len(in.Password) < 8 {
			return UserInput{}, fmt.Errorf("пароль слишком короткий")
		}
	}
	return in, nil
}

func printUser(u User) {
	fmt.Printf("ID: %-4d | Имя: %-20s | Email: %-28s | Возраст: %d | Роль: %s\n", u.ID, u.Name, u.Email, u.Age, u.Role)
}
