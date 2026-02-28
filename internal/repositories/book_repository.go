package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"desent-api/internal/models"
)

var ErrBookNotFound = errors.New("book not found")

type BookRepository interface {
	Create(ctx context.Context, book models.Book) (models.Book, error)
	FindAll(ctx context.Context, query models.BookListQuery) ([]models.Book, error)
	FindByID(ctx context.Context, id int64) (models.Book, error)
	UpdateByID(ctx context.Context, id int64, book models.Book) (models.Book, error)
	DeleteByID(ctx context.Context, id int64) error
}

type SQLiteBookRepository struct {
	db *sql.DB
}

func NewSQLiteBookRepository(db *sql.DB) *SQLiteBookRepository {
	return &SQLiteBookRepository{db: db}
}

func InitBooksSchema(ctx context.Context, db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS books (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	author TEXT NOT NULL,
	year INTEGER NOT NULL
);`

	_, err := db.ExecContext(ctx, query)
	return err
}

func (r *SQLiteBookRepository) Create(ctx context.Context, book models.Book) (models.Book, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO books (title, author, year) VALUES (?, ?, ?)`, book.Title, book.Author, book.Year)
	if err != nil {
		return models.Book{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return models.Book{}, err
	}

	book.ID = id
	return book, nil
}

func (r *SQLiteBookRepository) FindAll(ctx context.Context, query models.BookListQuery) ([]models.Book, error) {
	statement := strings.Builder{}
	statement.WriteString(`SELECT id, title, author, year FROM books`)

	args := make([]any, 0, 3)
	if query.Author != "" {
		statement.WriteString(` WHERE author = ?`)
		args = append(args, query.Author)
	}

	statement.WriteString(` ORDER BY id ASC`)

	if query.Page > 0 && query.Limit > 0 {
		offset := (query.Page - 1) * query.Limit
		statement.WriteString(` LIMIT ? OFFSET ?`)
		args = append(args, query.Limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, statement.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := make([]models.Book, 0)
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year); err != nil {
			return nil, err
		}

		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

func (r *SQLiteBookRepository) FindByID(ctx context.Context, id int64) (models.Book, error) {
	var book models.Book
	err := r.db.QueryRowContext(ctx, `SELECT id, title, author, year FROM books WHERE id = ?`, id).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.Year,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Book{}, ErrBookNotFound
		}

		return models.Book{}, err
	}

	return book, nil
}

func (r *SQLiteBookRepository) UpdateByID(ctx context.Context, id int64, book models.Book) (models.Book, error) {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE books SET title = ?, author = ?, year = ? WHERE id = ?`,
		book.Title,
		book.Author,
		book.Year,
		id,
	)
	if err != nil {
		return models.Book{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Book{}, err
	}

	if rowsAffected == 0 {
		return models.Book{}, ErrBookNotFound
	}

	book.ID = id
	return book, nil
}

func (r *SQLiteBookRepository) DeleteByID(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM books WHERE id = ?`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrBookNotFound
	}

	return nil
}
