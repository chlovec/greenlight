package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	// Use version in the where clause to manage race conditions.
	// If the version has changed, it will result in sql.ErrNoRows
	// and we will return edit conflict error
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version
	`
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice as a
	// variadic parameter and scanning the new version value into the movie struct
	err := m.DB.QueryRow(query, args...).Scan(&movie.Version)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return ErrEditConflict
	}

	return err
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

// Method for fetching a specific movie record.
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	// Add an ORDER BY clause and interpolate the sort column and direction. Importantly
	// notice that we also include a secondary sort on the movie ID to ensure a
	// consistent ordering.
	query := fmt.Sprintf(`
        SELECT id, created_at, title, year, runtime, genres, version
        FROM movies
        WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '') 
        AND (genres @> $2 OR $2 = '{}')     
        ORDER BY %s %s, id ASC
		Limit $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// As our SQL query now has quite a few placeholder parameters, let's collect the
	// values for the placeholders in a slice. Notice here how we call the limit() and
	// offset() methods on the Filters struct to get the appropriate values for the
	// LIMIT and OFFSET clauses.
	args := []any{title, pq.Array(genres), filters.limit(), filters.offset()}

	// And then pass the args slice to QueryContext() as a variadic parameter.
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Importantly, defer a call to rows.Close() to ensure that the resultset is closed
	// before GetAll() returns.
	defer rows.Close()

	// Initialize an empty slice to hold the movie data.
	movies := []*Movie{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var movie Movie

		// Scan the values from the row into the Movie struct. Again, note that we're
		// using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, err
		}

		// Add the Movie struct to the slice.
		movies = append(movies, &movie)
	}

	// After the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// If everything went OK, then return the slice of movies.
	return movies, nil
}
