package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/chlovec/greenlight/internal/data"
	"github.com/chlovec/greenlight/internal/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		// Use the new badRequestResponse() helper.
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// Initialize a new Validator.
	v := validator.New()

	// Call the ValidateMovie() function, and if any checks fail, return a response
	// containing the errors.
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Create the movie using the Insert() method on the movie model,
	// passing in a pointer to the validated movie struct. This will
	// create a record in the database and update the movie struct
	// with the system generated information
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Create a location header to be included in the http response to
	// let the client know which url they can find newly created
	// resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Read and validate id param.
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific movie.
	// Check if record was not found and respond with notFoundResponse()
	// If any other error is returned, respond with serverErrorResponse()
	movie, err := app.models.Movies.Get(id)
	if err != nil && errors.Is(err, data.ErrRecordNotFound) {
		app.notFoundResponse(w, r)
		return
	} else if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	env := envelope{"movie": movie}
	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Read and validate id param from request URL.
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing movie from the database and send 404 Not Found
	// to the client if no matching record was found
	movie, err := app.models.Movies.Get(id)
	if err != nil && errors.Is(err, data.ErrRecordNotFound) {
		app.notFoundResponse(w, r)
		return
	} else if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Declare an input struct to hold the expected data from the client.
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// Read the JSON request body data into the input struct.
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the request body to the appropriate fields of the movie record
	movie.Title = input.Title
	movie.Year = input.Year
	movie.Runtime = input.Runtime
	movie.Genres = input.Genres

	// Validate the updated movie record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update the movie record
	err = app.models.Movies.Update(movie)
	if err != nil && errors.Is(err, data.ErrEditConflict) {
		app.editConflictResponse(w, r)
		return
	} else if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Read and validate id param from request URL.
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the movie from the database. Send a 404 Not Found response to the
	// client there is no matching record.
	err = app.models.Movies.Delete(id)
	if err != nil && errors.Is(err, data.ErrRecordNotFound) {
		app.notFoundResponse(w, r)
		return
	} else if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send 200 OK with a message if serving only humans
	// Send just 204 No Content status code if our clients are not humans
	// or are a mix of humans and machines
	env := envelope{"message": "movie successfully deleted"}
	err = app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) patchMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Read and validate id param from request URL.
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing movie from the database and send 404 Not Found
	// to the client if no matching record was found
	movie, err := app.models.Movies.Get(id)
	if err != nil && errors.Is(err, data.ErrRecordNotFound) {
		app.notFoundResponse(w, r)
		return
	} else if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Declare an input struct to hold the expected data from the client.
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	// Read the JSON request body data into the input struct.
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy Title if provided
	if input.Title != nil {
		movie.Title = *input.Title
	}

	// Copy Year if provided
	if input.Year != nil {
		movie.Year = *input.Year
	}

	// Copy Runtime if provided
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	// Copy Genres if provided
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	// Validate the updated movie record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update the movie record
	err = app.models.Movies.Update(movie)
	if err != nil && errors.Is(err, data.ErrRecordNotFound) {
		app.editConflictResponse(w, r)
		return
	} else if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
