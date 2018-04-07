package mysql

import (
	"context"
	"database/sql"

	"github.com/nasa9084/ident/domain/entity"
)

func ExistUser(ctx context.Context, tx *sql.Tx, userID string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM users WHERE user_id = ?)`
	const exist = 1

	row := tx.QueryRowContext(ctx, query, userID)
	var resp int
	if err := row.Scan(&resp); err != nil {
		return false, err
	}
	return resp == exist, nil
}

func FindUser(ctx context.Context, tx *sql.Tx, userID string) (entity.User, error) {
	const query = `SELECT user_id, password, totp_secret, email FROM users WHERE user_id = ?`
	row := tx.QueryRowContext(ctx, query, userID)
	var u entity.User
	if err := row.Scan(&u.ID, &u.Password, &u.TOTPSecret, &u.Email); err != nil {
		return entity.User{}, err
	}
	u.TOTPVerified = true
	return u, nil
}

func UpdateUser(ctx context.Context, tx *sql.Tx, u entity.User) error {
	const query = `UPDATE users SET password=?, email=? WHERE user_id=?`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(u.Password, u.Email, u.ID); err != nil {
		return err
	}
	return nil
}

func CreateUser(ctx context.Context, tx *sql.Tx, u entity.User) error {
	const query = `INSERT INTO users(user_id, password, totp_secret, email) VALUES(?, ?, ?, ?)`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(u.ID, u.Password, u.TOTPSecret, u.Email); err != nil {
		return err
	}
	return nil
}

func DeleteUser(ctx context.Context, tx *sql.Tx, u entity.User) error {
	const query = `DELETE FROM users WHERE user_id = ?`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(u.ID); err != nil {
		return err
	}
	return nil
}
