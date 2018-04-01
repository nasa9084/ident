package database

import (
	"context"

	"github.com/nasa9084/ident/domain/entity"
)

func (repo *userRepository) existsInMySQL(ctx context.Context, userID string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM users WHERE user_id = ?)`
	const exist = 1

	row := repo.MySQL.QueryRowContext(ctx, query, userID)
	var resp int
	if err := row.Scan(&resp); err != nil {
		return false, err
	}
	return resp == exist, nil
}

func (repo *userRepository) findFromMySQL(ctx context.Context, userID string) (entity.User, error) {
	const query = `SELECT user_id, password, totp_secret, email FROM users WHERE user_id = ?`
	row := repo.MySQL.QueryRowContext(ctx, query, userID)
	var u entity.User
	if err := row.Scan(&u.ID, &u.Password, &u.TOTPSecret, &u.Email); err != nil {
		return nilUser, err
	}
	u.TOTPVerified = true
	return u, nil
}

func (repo *userRepository) updateMySQL(ctx context.Context, u entity.User) error {
	const query = `UPDATE users SET password=?, email=? WHERE user_id=?`
	tx, err := repo.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(u.Password, u.Email, u.ID); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *userRepository) createInMySQL(ctx context.Context, u entity.User) error {
	const query = `INSERT INTO users(user_id, password, totp_secret, email) VALUES(?, ?, ?, ?)`
	tx, err := repo.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(u.ID, u.Password, u.TOTPSecret, u.Email); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *userRepository) deleteFromMySQL(ctx context.Context, u entity.User) error {
	const query = `DELETE FROM users WHERE user_id = ?`
	tx, err := repo.MySQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(u.ID); err != nil {
		return err
	}
	return tx.Commit()
}
