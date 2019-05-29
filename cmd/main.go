package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/newtondev/service_object/pkg/storage"
	"github.com/newtondev/service_object/pkg/entities"
	"github.com/newtondev/service_object/pkg/constants"
	svcerrors "github.com/newtondev/service_object/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
)

func main() {
	var (
		addr  = flag.String("addr", ":8080", "address of the http server")
		debug = flag.Bool("debug", false, "enable debug")
	)

	stdout := ioutil.Discard
	if *debug {
		stdout = os.Stdout
	}

	r := storage.MemStore{}
	s := NewServer(*addr, stdout, &r)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("start server: %v", err)
	}
}

// NewServer prepares http server.
func NewServer(addr string, stdout io.Writer, r Repository) *http.Server {
	mux := http.NewServeMux()

	srv := &Service{
		Validator: &PlayValidator{
			Validator:  validator.New(),
			Repository: r,
		},
		Repository: r,
	}

	h := RegistrationHandler{
		Registrator: NewRegistratorWithLog(srv, stdout, os.Stderr),
	}

	mux.Handle("/register", &h)

	s := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &s
}

// Repository is a data access layer.
type Repository interface {
	Unique(ctx context.Context, email string) error
	Create(context.Context, *entities.Form) (*entities.User, error)
}

// Validator validation abstraction.
type Validator interface {
	Validate(context.Context, *entities.Form) error
}

// ValidationErrors holds validation errors.
type ValidationErrors map[string]string

// Error implements error interface
func (v ValidationErrors) Error() string {
	return constants.ValidationMsg
}

// Service holds data required for registration.
type Service struct {
	Validator
	Repository
}

// Register hold registration domain logic.
func (s *Service) Register(ctx context.Context, f *entities.Form) (*entities.User, error) {
	if err := s.Validator.Validate(ctx, f); err != nil {
		return nil, errors.Wrap(err, "validator validate")
	}

	user, err := s.Create(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "repository create")
	}

	return user, nil
}

// Registrator abstraction for registration service.
type Registrator interface {
	Register(context.Context, *entities.Form) (*entities.User, error)
}

// RegistrationHandler for registration requrests.
type RegistrationHandler struct {
	Registrator
}

// ServerHTTP implements http.Handler.
func (h *RegistrationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var f entities.Form
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := h.Register(r.Context(), &f)
	if err != nil {
		switch v := errors.Cause(err).(type) {
		case ValidationErrors:
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(v)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(&u)
}

// PlayValidator holds registration form validations.
type PlayValidator struct {
	Validator *validator.Validate
	Repository
}

// Validate implements Validator.
func (v *PlayValidator) Validate(ctx context.Context, f *entities.Form) error {
	validations := make(ValidationErrors)

	err := v.Validator.Struct(f)
	if err != nil {
		if vs, ok := err.(validator.ValidationErrors); ok {
			for _, v := range vs {
				validations[v.Tag()] = fmt.Sprintf("%s is invalid", v.Tag())
			}
		}
	}

	if f.Password != f.PasswordConfirmation {
		validations["password"] = constants.PasswordMismatch
	}

	if err := v.Repository.Unique(ctx, f.Email); err != nil {
		if err != svcerrors.ErrEmailExists {
			return errors.Wrap(err, "repository unique")
		}

		validations["email"] = constants.EmailExists
	}

	if len(validations) > 0 {
		return validations
	}

	return nil
}
