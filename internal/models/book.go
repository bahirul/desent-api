package models

type Book struct {
	ID     int64
	Title  string
	Author string
	Year   int
}

type CreateBookRequest struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

type BookResponse struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

func ToBookResponse(book Book) BookResponse {
	return BookResponse{
		Title:  book.Title,
		Author: book.Author,
		Year:   book.Year,
	}
}
