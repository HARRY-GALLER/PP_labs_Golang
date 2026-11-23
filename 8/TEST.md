## 1. Запуск сервера 

go run .



## 2. Проверка всех пяти REST API маршрутов

# POST Создание пользователя

$email = "lab8$(Get-Random)@example.com"
$body = @{
    name = "Ivan"
    age = 20
    email = $email
} | ConvertTo-Json

$user = Invoke-RestMethod -Uri "http://localhost:8080/users" -Method Post -ContentType "application/json" -Body $body

$user
$id = $user.id


# GET Получение списка пользователей

Invoke-RestMethod -Uri "http://localhost:8080/users" | ConvertTo-Json -Depth 5


# GET Получение пользователя по ID

Invoke-RestMethod -Uri "http://localhost:8080/users/$id"


# PUT Обновление пользователя

$body = @{
    name = "Ivan Petrov"
    age = 21
    email = $email
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/users/$id" -Method Put -ContentType "application/json" -Body $body


# DELETE Удаление пользователя

Invoke-RestMethod -Uri "http://localhost:8080/users/$id" -Method Delete



## 3. Проверка пагинации и фильтрации

# Фильтрация по имени
Invoke-RestMethod "http://localhost:8080/users?name=Ivan"

# Фильтрация по возрасту
Invoke-RestMethod "http://localhost:8080/users?age=21"

# Пагинация: первая страница, один пользователь
Invoke-RestMethod "http://localhost:8080/users?page=1&limit=1"

# Вторая страница
Invoke-RestMethod "http://localhost:8080/users?page=2&limit=1"


## 4. Проверка валидации и ошибок

# name - От 2 до 100 символов
# age - От 1 до 120
# email	- Корректный и уникальный
# id - Положительное целое число


# Неправильные данные:

$body = @{
    name = "A"
    age = 150
    email = "wrong"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/users" -Method Post -ContentType "application/json" -Body $body


# Несуществующий пользователь

Invoke-RestMethod "http://localhost:8080/users/999999"



## 5. Проверка PostgreSQL

docker compose exec -e PGPASSWORD=123 postgres psql -h 127.0.0.1 -U lab8 -d lab8 -c "SELECT * FROM users;"



## 6. Проверка unit-тестов

go test ./... -v















