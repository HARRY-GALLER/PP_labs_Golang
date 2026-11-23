# Лабораторная работа №10 — Go


## 1. Хэширование

```powershell
go run ./cmd/hash
```

Выбрать MD5, SHA-256 или SHA-512; ввести строку. Для проверки целостности ввести строку и её хэш в hex. Будет выведено совпадение/несовпадение.

## 2. AES

```powershell
go run ./cmd/aes
```

Ввести строку и ключ **ровно 16, 24 или 32 байта** (1234567890abcdef), получить шифртекст Base64. При повторном вводе правильного ключа выполняется расшифровка. Используется AES-GCM со случайным nonce.

## 3. Асимметричное шифрование и цифровая подпись

```powershell
go run ./cmd/sign
```

Создаётся пара RSA-ключей `keys/private.pem` и `keys/public.pem`.

## 4. TCP через TLS с взаимной аутентификацией

Один раз перед запуском создать локальный CA и сертификаты сервера/клиента:

```powershell
go run ./cmd/certs
```

**Терминал 1:**

```powershell
go run ./cmd/tls-server
```

**Терминал 2:**

```powershell
go run ./cmd/tls-client
```

Клиент проверяет цепочку сертификата сервера и имя `localhost`; сервер требует сертификат клиента и проверяет его по тому же CA. Адрес: `127.0.0.1:8443`. Передаваемая строка возвращается сервером через TLS.

Файлы в `certs/` — только для локальной работы

## 5. REST API — токены и роли


**Терминал 1:**

```powershell
docker compose up -d
```

**Терминал 2:**

```powershell
go run .
```

**Терминал 3:**

```powershell
go run ./cmd/client
```

1. Через меню клиента зарегистрировать пользователя, затем войти по email и паролю.
2. Каждый вход выдаёт случайный Bearer-токен, CSRF-токен и создаёт сессию на 24 часа. Токен клиент передаёт в `Authorization: Bearer <token>`.
3. По умолчанию регистрация создаёт роль `user`. Она может выполнять `GET /users` и `GET /users/{id}`, но не POST/PUT/DELETE пользователей.
4. Для проверки роли `admin` зарегистрировать `admin@example.com`, затем выполнить команду:

```powershell
docker compose exec db psql -U postgres -d users -c "UPDATE users SET role='admin' WHERE email='admin@example.com';"
```

5. Выйти и снова войти под `admin@example.com` (роль записывается в сессию при входе). Теперь доступны POST/PUT/DELETE `/users`.

Список защищённых маршрутов:

| Маршрут | Доступ |
|---|---|
| `GET /users`, `GET /users/{id}` | `user`, `admin` |
| `POST /users`, `PUT /users/{id}`, `DELETE /users/{id}` | только `admin` |
| `POST /auth/logout` | авторизованный пользователь |
| `POST /auth/register`, `POST /auth/login` | без авторизации |

### Проверка валидного и невалидного токена

В PowerShell после регистрации соответствующего пользователя:

```powershell
$login = Invoke-RestMethod -Method POST -Uri 'http://localhost:8080/auth/login' -ContentType 'application/json' -Body '{"email":"admin@example.com","password":"11111111"}'

# Валидный токен -> 200
Invoke-WebRequest -Uri 'http://localhost:8080/users' -Headers @{Authorization="Bearer $($login.token)"}

# Невалидный токен -> 401
Invoke-WebRequest -Uri 'http://localhost:8080/users' -Headers @{Authorization='Bearer invalid'}
```

В приведённом примере пароль `11111111` нужно заменить тем, который указан при регистрации.

## 6. CSRF и сессии

На запросах с изменениями (`POST /users`, `PUT /users/{id}`, `DELETE /users/{id}`, `POST /auth/logout`) помимо Bearer-токена требуется `X-CSRF-Token`, выданный для **той же сессии** при входе. Без него или с неверным значением сервер возвращает **403**. Консольный клиент автоматически добавляет оба заголовка.

Пример проверки (после `$login` из предыдущего раздела):

```powershell
$body = '{"name":"qwe","email":"qwe@example.com","age":20}'

# Отсутствует CSRF-токен -> 403
Invoke-WebRequest -Method POST -Uri 'http://localhost:8080/users' -ContentType 'application/json' -Headers @{Authorization="Bearer $($login.token)"} -Body $body

# Верный CSRF-токен и роль admin -> 201
Invoke-WebRequest -Method POST -Uri 'http://localhost:8080/users' -ContentType 'application/json' -Headers @{Authorization="Bearer $($login.token)"; 'X-CSRF-Token'=$login.csrf_token} -Body $body
```

При выходе сессия удаляется, Bearer-токен и CSRF-токен больше не работают. Сессии хранятся в памяти сервера и сбрасываются при его перезапуске. Авторизованные запросы не используют cookie, но в соответствии с заданием проверяют отдельный CSRF-токен.


