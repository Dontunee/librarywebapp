package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"html/template"
	"net/http"
	"strconv"
)

var (
	db  *sql.DB
	tpl *template.Template
)

//init in the pq package helps to register the driver with database/sql
func init() {
	var err error
	//initialize a new sql.DB
	db, err = sql.Open("postgres", "postgres://bond:password@localhost/bookstore?sslmode=disable")
	if err != nil {
		panic(err)
	}

	//verify connection to database
	if err = db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("Connected to database")
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

//export fields to templates
//fields changed to uppercase
type Book struct {
	Isbn        string
	Title       string
	Author      string
	Price       float32
	IsAvailable bool
}

func main() {
	//add routes and servers
	http.HandleFunc("/", index)
	http.HandleFunc("/books", libraryIndex)
	http.HandleFunc("/books/show", booksShow)
	http.HandleFunc("/books/create", booksCreateForm)
	http.HandleFunc("/books/create/process", booksCreateProcess)
	http.HandleFunc("/books/update", booksUpdateForm)
	http.HandleFunc("/books/update/process", booksUpdateProcess)
	http.HandleFunc("/books/delete/process", booksDeleteProcess)
	http.ListenAndServe(":8080", nil)
}

func index(responseWriter http.ResponseWriter, request *http.Request) {
	http.Redirect(responseWriter, request, "/books", http.StatusSeeOther)
}

func libraryIndex(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		http.Error(responseWriter, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
	//Query the db
	rows, err := db.Query("SELECT * FROM books")
	if err != nil {
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	books := make([]Book, 0)
	//iterate through results, prepares the first and each tp be acted upon the scan method
	for rows.Next() {
		book := Book{}
		//use Scan() to copy the values from each field in the row to a new book object created in line above
		err := rows.Scan(&book.Isbn, &book.Title, &book.Author, &book.Price, &book.IsAvailable)
		if err != nil {
			http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		//if no error occurred add to the list of books
		books = append(books, book)
	}
	//check if errors occurred in the iteration
	err = rows.Err()
	if err != nil {
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(responseWriter, "library-books.gohtml", books)
}

func booksIndex(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		http.Error(responseWriter, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
	//Query the db
	rows, err := db.Query("SELECT * FROM books")
	if err != nil {
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	books := make([]Book, 0)
	//iterate through results, prepares the first and each tp be acted upon the scan method
	for rows.Next() {
		book := Book{}
		//use Scan() to copy the values from each field in the row to a new book object created in line above
		err := rows.Scan(&book.Isbn, &book.Title, &book.Author, &book.Price)
		if err != nil {
			http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		//if no error occurred add to the list of books
		books = append(books, book)
	}
	//check if errors occurred in the iteration
	err = rows.Err()
	if err != nil {
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(responseWriter, "books.gohtml", books)
}

func booksShow(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		http.Error(responseWriter, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	isbn := request.FormValue("isbn")
	if isbn == "" {
		http.Error(responseWriter, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM books WHERE isbn =$1", isbn)

	book := Book{}
	err := row.Scan(&book.Isbn, &book.Title, &book.Author, &book.Price)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(responseWriter, request)
		return
	case err != nil:
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(responseWriter, "show.gohtml", book)
}

func booksCreateForm(responseWriter http.ResponseWriter, request *http.Request) {
	tpl.ExecuteTemplate(responseWriter, "create.gohtml", nil)
}

func booksCreateProcess(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(responseWriter, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	//get form values
	book := Book{}
	book.Isbn = request.FormValue("isbn")
	book.Title = request.FormValue("title")
	book.Author = request.FormValue("author")
	price := request.FormValue("price")

	//validate form values
	if book.Isbn == "" || book.Title == "" || book.Author == "" || price == "" {
		http.Error(responseWriter, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	//convert form values
	validPrice, err := strconv.ParseFloat(price, 32)
	if err != nil {
		http.Error(responseWriter, http.StatusText(http.StatusNotAcceptable)+"Please hit back and enter a number for the price",
			http.StatusNotAcceptable)
		return
	}
	book.Price = float32(validPrice)

	//insert values
	_, err = db.Exec("INSERT INTO books(isbn,title,author,price) VALUES ($1,$2,$3,$4)",
		book.Isbn, book.Title, book.Author, book.Price)
	if err != nil {
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	//confirm insertion
	tpl.ExecuteTemplate(responseWriter, "created.gohtml", book)
}

func booksUpdateForm(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		http.Error(responseWriter, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	isbn := request.FormValue("isbn")
	if isbn == "" {
		http.Error(responseWriter, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM  WHERE isbn = $1", isbn)

	book := Book{}
	err := row.Scan(&book.Isbn, &book.Title, &book.Author, &book.Price)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(responseWriter, request)
		return
	case err != nil:
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(responseWriter, "update.gohtml", book)
}

func booksUpdateProcess(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		http.Error(responseWriter, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	//get form values
	book := Book{}
	book.Isbn = request.FormValue("isbn")
	book.Title = request.FormValue("title")
	book.Author = request.FormValue("author")
	price := request.FormValue("price")

	//validate form values
	if book.Isbn == "" || book.Title == "" || book.Author == "" || price == "" {
		http.Error(responseWriter, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	//convert form values
	validPrice, err := strconv.ParseFloat(price, 32)
	if err != nil {
		http.Error(responseWriter, http.StatusText(http.StatusNotAcceptable)+"Please hit back and enter a number for the price",
			http.StatusNotAcceptable)
		return
	}
	book.Price = float32(validPrice)

	//insert values
	_, err = db.Exec("UPDATE books SET isbn = $1, title = $2 , author = $3, price = $4 WHERE isbn = $1;",
		book.Isbn, book.Title, book.Author, book.Price)
	if err != nil {
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	//confirm insertion
	tpl.ExecuteTemplate(responseWriter, "updated.gohtml", book)
}

func booksDeleteProcess(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		http.Error(responseWriter, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	isbn := request.FormValue("isbn")
	if isbn == "" {
		http.Error(responseWriter, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	//delete book
	_, err := db.Exec("DELETE FROM books WHERE isbn = $1;", isbn)
	if err != nil {
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(responseWriter, request, "/books", http.StatusSeeOther)
}
