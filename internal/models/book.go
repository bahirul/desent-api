package models

type Book struct {
	ID     int64
	Title  string
	Author string
	Year   int
}

type BookListQuery struct {
	Author string
	Page   int
	Limit  int
}

type CreateBookRequest struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

type BookResponse struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

func ToBookResponse(book Book) BookResponse {
	return BookResponse{
		ID:     book.ID,
		Title:  book.Title,
		Author: book.Author,
		Year:   book.Year,
	}
}
