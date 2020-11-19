package repositories

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

// User model
type User struct {
	ID         string `json:"id" db:"id"`
	Name       string `json:"name" db:"name"`
	Email      string `json:"email" db:"email"`
	Picture    string `json:"picture" db:"picture"`
	OIDCUserId string `json:"-" db:"oidc_userid"`
}

// UsersRepositoryInterface defines the set of User related methods available
type UsersRepositoryInterface interface {
	GetAll(ctx context.Context) ([]*User, error)
	FindByID(ctx context.Context, ID string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindOrCreateUser(ctx context.Context, userData *User) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, ID string) (bool, error)
}

// UsersRepository implements UsersRepositoryInterface
type UsersRepository struct {
	db *sqlx.DB
}

// NewUsersRepository returns a configured UsersRepository object
func NewUsersRepository(db *sqlx.DB) *UsersRepository {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) GetDB() *sqlx.DB {
	return r.db
}

// GetAll fetches all users, returns an empty slice if no user exists
func (r *UsersRepository) GetAll(ctx context.Context) ([]*User, error) {
	users := []*User{}
	err := r.db.SelectContext(ctx, &users, "SELECT id, name, email FROM users")
	if err != nil {
		return nil, err
	}

	return users, nil
}

// FindByID finds a user by ID, returns nil if not found
func (r *UsersRepository) FindByID(ctx context.Context, ID string) (*User, error) {
	user := &User{}
	err := r.db.GetContext(ctx, user, "SELECT id, name, email FROM users WHERE id = $1", ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// FindByEmail finds a user by email, returns nil if not found
func (r *UsersRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	stmt := "SELECT id, name, email FROM users WHERE email = $1"
	err := r.db.GetContext(ctx, user, stmt, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// FindOrCreateUser finds a user by email and creates it if not found
// TODO deal with passing txn around
func (r *UsersRepository) FindOrCreateUser(ctx context.Context, userData *User) (*User, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user := &User{}
	selectStmt := "SELECT id, name, email FROM users WHERE email = $1"
	err = tx.GetContext(ctx, user, selectStmt, userData.Email)
	if err == nil {
		return user, nil
	}

	insertStmt := "INSERT INTO users (name, email, oidc_userid) VALUES ($1, $2, $3) RETURNING id"
	res, err := tx.ExecContext(ctx, insertStmt, userData.Name, userData.Email, userData.OIDCUserId)
	if err != nil {
		return nil, parseError(err)
	}

	if rows, err := res.RowsAffected(); err != nil {
		if rows == 0 {
			return nil, nil
		}
		return nil, parseError(err)
	}

	err = tx.GetContext(ctx, user, selectStmt, userData.Email)
	if err != nil {
		return nil, parseError(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, parseError(err)
	}

	return user, nil
}

// Create creates a new user, returning the full model
func (r *UsersRepository) Create(ctx context.Context, user *User) (*User, error) {
	stmt := "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"
	row := r.db.QueryRowxContext(ctx, stmt, user.Name, user.Email)
	err := row.Scan(&user.ID)
	if err != nil {
		return nil, parseError(err)
	}
	return user, nil
}

// Update updates a user, returning the updated model or nil if no rows were affected
func (r *UsersRepository) Update(ctx context.Context, user *User) (*User, error) {
	stmt := "UPDATE users SET name = $1, email = $2 WHERE id = $3"
	res, err := r.db.ExecContext(ctx, stmt, user.Name, user.Email, user.ID)
	if err != nil {
		return nil, parseError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, nil
	}
	return user, nil
}

// Delete deletes a user, only returns error if action fails
func (r *UsersRepository) Delete(ctx context.Context, ID string) (bool, error) {
	stmt := "DELETE FROM users WHERE id = $1 RETURNING id"
	res, err := r.db.ExecContext(ctx, stmt, ID)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}