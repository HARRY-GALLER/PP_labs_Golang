package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type postgresRepository struct {
	db *sql.DB
}

func (p *postgresRepository) List(ctx context.Context, name string, age, page, limit int) (UserList, error) {
	result := UserList{Users: make([]User, 0), Page: page, Limit: limit}
	const filter = `WHERE ($1 = '' OR strpos(lower(name), lower($1)) > 0)
		AND ($2 = 0 OR age = $2)`
	if err := p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users `+filter, name, age).Scan(&result.Total); err != nil {
		return result, err
	}
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, name, age, email FROM users `+filter+
			` ORDER BY id LIMIT $3 OFFSET $4`, name, age, limit, (page-1)*limit)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Age, &user.Email); err != nil {
			return result, err
		}
		result.Users = append(result.Users, user)
	}
	return result, rows.Err()
}

func (p *postgresRepository) Get(ctx context.Context, id int64) (User, error) {
	var user User
	err := p.db.QueryRowContext(ctx,
		`SELECT id, name, age, email FROM users WHERE id=$1`, id).
		Scan(&user.ID, &user.Name, &user.Age, &user.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return user, ErrNotFound
	}
	return user, err
}

func (p *postgresRepository) Create(ctx context.Context, input UserInput) (User, error) {
	var user User
	err := p.db.QueryRowContext(ctx,
		`INSERT INTO users (name, age, email) VALUES ($1, $2, $3)
		 RETURNING id, name, age, email`, input.Name, input.Age, input.Email).
		Scan(&user.ID, &user.Name, &user.Age, &user.Email)
	return user, mapDBError(err)
}

func (p *postgresRepository) Update(ctx context.Context, id int64, input UserInput) (User, error) {
	var user User
	err := p.db.QueryRowContext(ctx,
		`UPDATE users SET name=$1, age=$2, email=$3 WHERE id=$4
		 RETURNING id, name, age, email`, input.Name, input.Age, input.Email, id).
		Scan(&user.ID, &user.Name, &user.Age, &user.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return user, ErrNotFound
	}
	return user, mapDBError(err)
}

func (p *postgresRepository) Delete(ctx context.Context, id int64) error {
	result, err := p.db.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func mapDBError(err error) error {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { //нарушение уникальности
		return ErrEmailExists
	}
	return err
}
