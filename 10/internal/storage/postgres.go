// Package storage — реализация user.Repository поверх PostgreSQL.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"userapi/internal/user"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    email      TEXT        NOT NULL UNIQUE,
    age        INTEGER     NOT NULL CHECK (age BETWEEN 1 AND 150),
    password_hash TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user';
CREATE INDEX IF NOT EXISTS idx_users_age ON users (age);
`

const columns = "id, name, email, age, role, created_at, updated_at"

type Store struct{ db *sql.DB }

var _ user.Repository = (*Store)(nil)

func New(db *sql.DB) *Store { return &Store{db: db} }

// Migrate создаёт таблицы, если их ещё нет.
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, schema)
	return err
}

type scanner interface{ Scan(dest ...any) error }

func scanUser(row scanner) (user.User, error) {
	var u user.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Age, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

// mapErr переводит ошибки драйвера в доменные.
func mapErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return user.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return user.ErrEmailTaken
	}
	return err
}

func (s *Store) List(ctx context.Context, p user.ListParams) ([]user.User, int, error) {
	var (
		conds []string
		args  []any
	)
	add := func(cond string, v any) {
		args = append(args, v)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}
	if p.Name != "" {
		esc := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(p.Name)
		add("name ILIKE $%d", "%"+esc+"%")
	}
	if p.Age != nil {
		add("age = $%d", *p.Age)
	}
	if p.MinAge != nil {
		add("age >= $%d", *p.MinAge)
	}
	if p.MaxAge != nil {
		add("age <= $%d", *p.MaxAge)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, p.Limit, p.Offset)
	q := fmt.Sprintf("SELECT %s FROM users%s ORDER BY id LIMIT $%d OFFSET $%d",
		columns, where, len(args)-1, len(args))
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := []user.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (s *Store) Get(ctx context.Context, id int64) (user.User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx, "SELECT "+columns+" FROM users WHERE id = $1", id))
	return u, mapErr(err)
}

func (s *Store) Create(ctx context.Context, in user.Input) (user.User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx,
		"INSERT INTO users (name, email, age) VALUES ($1, $2, $3) RETURNING "+columns,
		in.Name, in.Email, in.Age))
	return u, mapErr(err)
}

func (s *Store) Update(ctx context.Context, id int64, in user.Input) (user.User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx,
		"UPDATE users SET name = $1, email = $2, age = $3, updated_at = now() WHERE id = $4 RETURNING "+columns,
		in.Name, in.Email, in.Age, id))
	return u, mapErr(err)
}

func (s *Store) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return user.ErrNotFound
	}
	return nil
}

// CreateWithPassword создает пользователя и его учетные данные одним запросом.
func (s *Store) CreateWithPassword(ctx context.Context, in user.Input, passwordHash string) (user.User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx,
		"INSERT INTO users (name, email, age, password_hash) VALUES ($1, $2, $3, $4) RETURNING "+columns,
		in.Name, in.Email, in.Age, passwordHash))
	return u, mapErr(err)
}

func (s *Store) Credentials(ctx context.Context, email string) (int64, string, string, error) {
	var id int64
	var hash, role string
	err := s.db.QueryRowContext(ctx,
		"SELECT id, password_hash, role FROM users WHERE email = $1", email).Scan(&id, &hash, &role)
	return id, hash, role, mapErr(err)
}
