package data

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type MovieModel struct {
	DB *sql.DB
}

// Method for fetching a specific movie record.
func (m MovieModel) Get(id int64) (*Movie, error) {
	// SQL query for retrieving the movie data
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1
	`

	// Declare a movie struct to hold the data returned by the query.
	var movie Movie

	// Execute the query using QueryRow() method, passing in the
	// provided id value as a placeholder parameter, and scan the
	// response data into the fields of the movie struct.
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	// Handle any errors. If there was no record found, Scan()
	// will return a sql.ErrNoRows error. Check for this and
	// return the custom ErrRecordNotFound
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRecordNotFound
	} else if err != nil {
		return nil, err
	}

	// Otherwise return a pointer to the movie struct
	return &movie, nil
}

// Method for inserting a new movie record in the movies table.
func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
	`

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Method for updating a specific movie record in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	// SQL query for updating a movie record and returning the new version
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5
		RETURNING version
	`
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct
	return m.DB.QueryRow(query, args...).Scan(&movie.Version)
}

// Method for deleting a specific movie record.
func (m MovieModel) Delete(id int64) error {
	query := `DELETE FROM movies WHERE id = $1`

	// Execute SQL query using the Exec() method, passing in the id variable as
	// the value for the placeholder parameter. The Exec() method returns a sql.Result
	// value
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
