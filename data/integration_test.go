//go:build integration

// run tests with the following command count=1 to not cache tests : go test . --tags integration --count=1
package data

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	host     = "localhost"
	user     = "postgres"
	password = "secret"
	dbName   = "scotch-app-test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var dummyUser = User{
	FirstName: "Some",
	LastName:  "Guy",
	Email:     "me@here.com",
	Active:    1,
	Password:  "password",
}

var models Models
var testDB *sql.DB
var resource *dockertest.Resource
var pool *dockertest.Pool

func TestMain(m *testing.M) {
	os.Setenv("DATABASE_TYPE", "postgres")
	os.Setenv("UPPER_DB_LOG", "ERROR")
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	pool = p
	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13.4",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {{HostIP: "0.0.0.0", HostPort: port}},
		},
	}
	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		// _ = pool.Purge(resource)
		log.Fatalf("Could not start resource: %s", err)
	}

	// to poll whether db is open for connection
	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("Could not connect to docker: %s", err)
	}

	err = createTables(testDB)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("Could not create tables: %s", err)
	}

	models = New(testDB)
	codeL := m.Run()

	defer func() {
		pool.Purge(resource)
		os.Exit(codeL)
	}()
}

func createTables(db *sql.DB) error {
	stmt := `
	
	CREATE OR REPLACE FUNCTION trigger_set_timestamp()
		RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = NOW();
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	drop table if exists users cascade;

	CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		first_name character varying(255) NOT NULL,
		last_name character varying(255) NOT NULL,
		user_active integer NOT NULL DEFAULT 0,
		email character varying(255) NOT NULL UNIQUE,
		password character varying(60) NOT NULL,
		created_at timestamp without time zone NOT NULL DEFAULT now(),
		updated_at timestamp without time zone NOT NULL DEFAULT now()
	);

	CREATE TRIGGER set_timestamp
		BEFORE UPDATE ON users
		FOR EACH ROW
		EXECUTE PROCEDURE trigger_set_timestamp();

	drop table if exists remember_tokens;

	CREATE TABLE remember_tokens (
		id SERIAL PRIMARY KEY,
		user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
		remember_token character varying(100) NOT NULL,
		created_at timestamp without time zone NOT NULL DEFAULT now(),
		updated_at timestamp without time zone NOT NULL DEFAULT now()
	);

	CREATE TRIGGER set_timestamp
		BEFORE UPDATE ON remember_tokens
		FOR EACH ROW
		EXECUTE PROCEDURE trigger_set_timestamp();

	drop table if exists tokens;

	CREATE TABLE tokens (
		id SERIAL PRIMARY KEY,
		user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
		first_name character varying(255) NOT NULL,
		email character varying(255) NOT NULL,
		token character varying(255) NOT NULL,
		token_hash bytea NOT NULL,
		created_at timestamp without time zone NOT NULL DEFAULT now(),
		updated_at timestamp without time zone NOT NULL DEFAULT now(),
		expiry timestamp without time zone NOT NULL
	);

	CREATE TRIGGER set_timestamp
		BEFORE UPDATE ON tokens
		FOR EACH ROW
		EXECUTE PROCEDURE trigger_set_timestamp();


	`
	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}

func TestUser_Table(t *testing.T) {
	s := models.Users.Table()
	if s != "users" {
		t.Errorf("expected users, got %s", s)
	}
}

func TestUser_Insert(t *testing.T) {
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	if id != 1 {
		t.Errorf("expected 1, got %d", id)
	}
}
func TestUser_Get(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error(err)
	}
	if u.ID != 1 {
		t.Errorf("expected id to be 1, got %d", u.ID)
	}
}
func TestUser_GetAll(t *testing.T) {
	_, err := models.Users.GetAll()
	if err != nil {
		t.Error(err)
	}
}

func TestUser_GetByEmail(t *testing.T) {
	u, err := models.Users.GetByEmail("me@here.com")
	if err != nil {
		t.Error(err)
	}
	if u.ID != 1 {
		t.Errorf("expected id to be 1, got %d", u.ID)
	}
}

func TestUser_Update(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error(err)
	}
	u.LastName = "Smith"
	err = u.Update(u)
	if err != nil {
		t.Error(err)
	}
	u, err = models.Users.Get(1)
	if err != nil {
		t.Error(err)
	}
	if u.LastName != "Smith" {
		t.Errorf("expected last name to be Smith, got %s", u.LastName)
	}

}

func TestUser_PasswordMatches(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error(err)
	}
	matches, err := u.PasswordMatches("password")
	if err != nil {
		t.Error(err)
	}

	if !matches {
		t.Errorf("expected matches to be true, got %t", matches)
	}
}

func TestUser_ResetPassword(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error(err)
	}
	err = u.ResetPassword(1, "<PASSWORD>")
	if err != nil {
		t.Error(err)
	}
	err = models.Users.ResetPassword(2, "new_password")
	if err == nil {
		t.Error("did not throw error when trying to reset password for a non-existent user")
	}
}

func TestUser_Delete(t *testing.T) {
	err := models.Users.Delete(1)
	if err != nil {
		t.Error("failed to delete user", err)
	}
	_, err = models.Users.Get(1)
	if err == nil {
		t.Error("failed to delete user")
	}
}

func TestToken_Table(t *testing.T) {
	s := models.Tokens.Table()
	if s != "tokens" {
		t.Errorf("expected tokens, got %s", s)
	}
}

func TestToken_GenerateToken(t *testing.T) {
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Error("error inserting user")
	}

	_, err = models.Tokens.GenerateToken(id, time.Hour*24*365)
	if err != nil {
		t.Error("error generating token")
	}
}

func TestToken_Insert(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user by email")
	}
	token, err := models.Tokens.GenerateToken(u.ID, time.Hour*24*365)
	if err != nil {
		t.Error("failed to generate token")
	}
	err = models.Tokens.Insert(*token, *u)
	if err != nil {
		t.Error("failed to insert token")
	}
}

func TestToken_GetUserForToken(t *testing.T) {
	token := "abc"
	_, err := models.Tokens.GetUserForToken(token)
	if err == nil {
		t.Error("failed to get user for token")
	}
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user by email")
	}
	_, err = models.Tokens.GenerateToken(u.ID, time.Hour*24*365)
	if err != nil {
		t.Error("failed to generate token")
	}

}

func TestToken_GetTokensForUser(t *testing.T) {
	tokens, err := models.Tokens.GetTokensForUser(1)
	if err != nil {
		t.Error("failed to get tokens for user")
	}
	if len(tokens) > 0 {
		t.Error("failed to get tokens for user")
	}
}

func TestToken_Get(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user by email")

	}

	_, err = models.Tokens.Get(u.Token.ID)
	if err != nil {
		t.Error("failed to get token")
	}
}

func TestToken_GetByToken(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user by email")

	}

	_, err = models.Tokens.GetByToken(u.Token.PlainText)
	if err != nil {
		t.Error("failed to get token")
	}

	_, err = models.Tokens.GetByToken("123")
	if err == nil {
		t.Error("failed to get token")
	}
}

var authData = []struct {
	name        string
	token       string
	email       string
	errExpected bool
	message     string
}{
	{"invalid", "abcdefghijklmnopqrstuvwxyz", "a@here.com", true, "invalid token accepted as valid"},
	{"invalid_length", "abcdefghijklmnopqrstuvwxyz", "a@here.com", true, "token of wrong length accepted as valid token"},
	{"no_user", "abcdefghijklmnopqrstuvwxyz", "a@here.com", true, "no user, but token accepted as valid"},
	{"valid", "", "me@here.com", false, "valid token reported as invalid"},
}

func TestToken_AuthenticateToken(t *testing.T) {
	for _, tt := range authData {
		token := ""
		if tt.email == dummyUser.Email {

			user, err := models.Users.GetByEmail(tt.email)
			if err != nil {
				t.Error("failed to get user by email:", tt.email, err)
			}
			token = user.Token.PlainText
		} else {
			token = tt.token
		}

		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		_, err := models.Tokens.Authenticate(req)
		if tt.errExpected && err == nil {
			t.Errorf("expected error for %s", tt.name)
		} else if !tt.errExpected && err != nil {
			t.Errorf("expected no error for %s", tt.name)
		} else {
			t.Logf("%s: %s", tt.name, err)
		}

	}
}

func TestToken_Delete(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user by email")
	}
	err = models.Tokens.DeleteByToken(u.Token.PlainText)
	if err != nil {
		t.Error("failed to delete token:", err)
	}

}

func TestToken_ExpiredToken(t *testing.T) {
	// insert a token
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user by email")
	}
	token, err := models.Tokens.GenerateToken(u.ID, -time.Hour*24*365)
	if err != nil {
		t.Error("failed to generate token")
	}
	err = models.Tokens.Insert(*token, *u)
	if err != nil {
		t.Error("failed to delete token:", err)
	}

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token.PlainText)
	_, err = models.Tokens.Authenticate(req)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestToken_BadHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	_, err := models.Tokens.Authenticate(req)
	if err == nil {
		t.Error("expected error for bad header")
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "abc")
	_, err = models.Tokens.Authenticate(req)
	if err == nil {
		t.Error("expected error for bad header")
	}
	newUser := User{
		FirstName: "temp",
		LastName:  "temp_last",
		Email:     "you@there.com",
		Active:    1,
		Password:  "abc",
	}
	id, err := models.Users.Insert(newUser)
	if err != nil {
		t.Error("failed to insert user")
	}
	token, err := models.Tokens.GenerateToken(id, time.Hour*24*365)
	err = models.Tokens.Insert(*token, newUser)
	if err != nil {
		t.Error("failed to delete token:", err)
	}
	err = models.Users.Delete(id)
	if err != nil {
		t.Error("failed to delete user:", err)
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	_, err = models.Tokens.Authenticate(req)
	if err == nil {
		t.Error("expected error for bad header")
	}

}

func TestToken_DeleteNonExistentToken(t *testing.T) {
	err := models.Tokens.DeleteByToken("abcasd")
	if err != nil {
		t.Error("expected error for non-existent token")
	}
}

func TestToken_ValidateToken(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get user by email")
	}
	token, err := models.Tokens.GenerateToken(u.ID, time.Hour*24)
	err = models.Tokens.Insert(*token, *u)
	if err != nil {
		t.Error("failed to delete token:", err)
	}
	okay, err := models.Tokens.ValidToken(token.PlainText)
	if err != nil {
		t.Error("failed to validate token:", err)
	}
	if !okay {
		t.Error("valid token reported invalid")
	}
	okay, err = models.Tokens.ValidToken("abc")
	if okay {
		t.Error("invalid token reported as valid")
	}
}
