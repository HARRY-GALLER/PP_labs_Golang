# Документация REST API

Базовый адрес: `http://localhost:8080`. Формат данных: **JSON**. Все примеры ниже предполагают, что сервер запущен.

## Структура пользователя

```json
{
  "id": 1,
  "name": "Ivan",
  "age": 20,
  "email": "ivan@example.com"
}
```

- `id`: положительное целое число, генерируется БД и не передаётся в POST/PUT.
- `name`: строка от **2 до 100 символов** (пробелы по краям удаляются).
- `age`: целое число от **1 до 120**.
- `email`: email в правильном формате, до **254 символов**, уникальный.

Для **POST** и **PUT** передаются все три поля: `name`, `age`, `email`; `PUT` полностью заменяет информацию пользователя.

## Маршруты

| Метод | Путь | Назначение | Код успеха |
|---|---|---|---|
| GET | `/users` | список пользователей | 200 |
| GET | `/users/{id}` | пользователь по ID | 200 |
| POST | `/users` | создать пользователя | 201 |
| PUT | `/users/{id}` | заменить данные пользователя | 200 |
| DELETE | `/users/{id}` | удалить пользователя | 204 |

### GET /users

**Параметры запроса:**

| Параметр | Тип | По умолчанию | Назначение |
|---|---|---|---|
| `page` | int, 1–100000 | 1 | номер страницы |
| `limit` | int, 1–100 | 10 | пользователей на странице |
| `name` | строка ≤100 символов | нет | поиск подстроки в имени без учёта регистра |
| `age` | int, 1–120 | нет | точный возраст |

Пример:

```http
GET /users?name=iv&age=20&page=1&limit=5 HTTP/1.1
Host: localhost:8080
```

Возможный ответ:

```json
{
  "users": [{"id":1,"name":"Ivan","age":20,"email":"ivan@example.com"}],
  "total": 1,
  "page": 1,
  "limit": 5
}
```

`total` — количество пользователей по фильтру **до** пагинации. Если совпадений нет, `users` — пустой массив `[]`.

### GET /users/{id}

```http
GET /users/1 HTTP/1.1
Host: localhost:8080
```

Ответ `200`:

```json
{"id":1,"name":"Ivan","age":20,"email":"ivan@example.com"}
```

Если ID отсутствует — `404`. Если ID недопустим — `400`.

### POST /users

```http
POST /users HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{"name":"Ivan","age":20,"email":"ivan@example.com"}
```

Ответ `201 Created`:

```json
{"id":1,"name":"Ivan","age":20,"email":"ivan@example.com"}
```

Заголовок `Location: /users/1` указывает на созданный ресурс.

### PUT /users/{id}

```http
PUT /users/1 HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{"name":"Ivan Petrov","age":21,"email":"ivan@example.com"}
```

Ответ `200 OK`:

```json
{"id":1,"name":"Ivan Petrov","age":21,"email":"ivan@example.com"}
```

Если ID отсутствует — `404`.

### DELETE /users/{id}

```http
DELETE /users/1 HTTP/1.1
Host: localhost:8080
```

Ответ `204 No Content`, без тела. Повторное удаление отсутствующего пользователя — `404`.

## Ошибки

API использует общий формат JSON:

```json
{"error":"Пользователь не найден"}
```

| Статус | Причина |
|---|---|
| 400 Bad Request | некорректный ID, JSON, параметры фильтра/пагинации, имя, возраст или email |
| 404 Not Found | пользователя с таким ID нет |
| 405 Method Not Allowed | метод не поддерживается для данного маршрута |
| 409 Conflict | email уже принадлежит другому пользователю |
| 500 Internal Server Error | ошибка работы сервера или базы данных |

Все ошибки, в том числе `404` для неизвестных путей и `405` для неподдерживаемых методов, возвращаются в общем формате JSON. Для `405` сервер также отправляет заголовок `Allow`.

## Примеры curl (bash / macOS / Linux)

```bash
curl http://localhost:8080/users
curl http://localhost:8080/users/1
curl -X POST http://localhost:8080/users -H 'Content-Type: application/json' -d '{"name":"Ivan","age":20,"email":"ivan@example.com"}'
curl -X PUT http://localhost:8080/users/1 -H 'Content-Type: application/json' -d '{"name":"Ivan Petrov","age":21,"email":"ivan@example.com"}'
curl -X DELETE http://localhost:8080/users/1
```
