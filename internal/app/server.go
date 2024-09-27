package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/RobinHood3082/simplebank/internal/token"
	"github.com/RobinHood3082/simplebank/pkg/router"
	"github.com/RobinHood3082/simplebank/util"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for our banking service
type Server struct {
	store      persistence.Store
	router     *router.Router
	logger     *slog.Logger
	validate   *validator.Validate
	tokenMaker token.Maker
	config     util.Config
}

// NewServer creates a new HTTP server and set up routing
func NewServer(store persistence.Store, logger *slog.Logger, validate *validator.Validate, tokenMaker token.Maker, config util.Config) *Server {
	server := &Server{store: store, logger: logger, validate: validate, tokenMaker: tokenMaker, config: config}
	server.getRoutes()
	return server
}

// Start runs the HTTP server on a specific address
func (server *Server) Start(addr string) error {
	server.logger.Info(fmt.Sprintf("Starting server on %s", addr))
	return server.router.Serve(addr)
}

// ErrorResponse represents an error message from the server
type ErrorResponse struct {
	Message string `json:"message"`
}

// writeError writes an error message and status code to the response writer
func (server *Server) writeError(w http.ResponseWriter, status int, err error) {
	var response ErrorResponse
	response.Message = err.Error()
	_ = server.writeJSON(w, status, response, nil)
}

// writeJSON writes the data to the response writer
func (server *Server) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(js)
	return nil
}

// readJSON reads the data from the http.Request
func (server *Server) readJSON(w http.ResponseWriter, r *http.Request, v any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(v)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// bindData binds the data from the http.Request to the struct v
func (server *Server) bindData(w http.ResponseWriter, r *http.Request, v any) error {
	err := server.readJSON(w, r, v)
	if err != nil {
		server.logger.Error(err.Error())
		return err
	}

	err = server.validate.Struct(v)
	if err != nil {
		server.logger.Error(err.Error())
		return err
	}

	return nil
}

// The readString() helper returns a string value from the query string, or the provided
// default value if no matching key could be found.
func (server *Server) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	return s
}

// The readCSV() helper reads a string value from the query string and then splits it
// into a slice on the comma character. If no matching key could be found, it returns
// the provided default value.
func (server *Server) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

// The readInt32() helper reads a string value from the query string and converts it to an
// int32 before returning. If no matching key could be found it returns the provided
// default value. If the value couldn't be converted to an integer, then we return an
// error.
func (server *Server) readInt32(qs url.Values, key string, defaultValue int, dest *int32) error { // Extract the value from the query string.
	s := qs.Get(key)
	if s == "" {
		*dest = int32(defaultValue)
		return nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		*dest = int32(defaultValue)
		return fmt.Errorf("invalid value for %s", key)
	}

	*dest = int32(i)
	return nil
}

// The readInt64() helper reads a string value from the query string and converts it to an
// int64 before returning. If no matching key could be found it returns the provided
// default value. If the value couldn't be converted to an integer, then we return an
// error.
func (server *Server) readInt64(qs url.Values, key string, defaultValue int, dest *int64) error {
	s := qs.Get(key)
	if s == "" {
		*dest = int64(defaultValue)
		return nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		*dest = int64(defaultValue)
		return fmt.Errorf("invalid value for %s", key)
	}

	*dest = int64(i)
	return nil
}
