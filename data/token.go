package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"net/http"
	"strings"
	"time"

	up "github.com/upper/db/v4"
)

type Token struct {
	ID        int       `db:"id,omitempty" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	FirstName string    `db:"first_name" json:"first_name"`
	Email     string    `db:"email" json:"email"`
	PlainText string    `db:"token" json:"token"`
	Hash      []byte    `db:"token_hash" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Expires   time.Time `db:"expiry" json:"expiry"`
}

func (t *Token) Table() string {
	return "tokens"
}

func (t *Token) GetUserForToken(token string) (*User, error) {
	var theUser User
	var theToken Token

	collection := upper.Collection(t.Table())
	err := collection.Find(up.Cond{"token": token}).One(&theToken)
	if err != nil {
		return nil, err
	}
	collection = upper.Collection(theUser.Table())
	err = collection.Find(up.Cond{"id =": theToken.UserID}).One(&theUser)
	if err != nil {
		return nil, err
	}
	theUser.Token = theToken
	return &theUser, nil
}

func (t *Token) GetTokensForUser(id int) ([]*Token, error) {
	var tokens []*Token
	collection := upper.Collection(t.Table())
	err := collection.Find(up.Cond{"user_id": id}).All(&tokens)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (t *Token) Get(id int) (*Token, error) {
	var theToken Token
	collection := upper.Collection(t.Table())
	err := collection.Find(up.Cond{"id =": id}).One(&theToken)
	if err != nil {
		return nil, err
	}
	return &theToken, nil
}

func (t *Token) GetByToken(plainText string) (*Token, error) {
	var theToken Token
	collection := upper.Collection(t.Table())
	err := collection.Find(up.Cond{"token": plainText}).One(&theToken)
	if err != nil {
		return nil, err
	}
	return &theToken, nil
}

func (t *Token) Delete(id int) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(id)
	err := res.Delete()
	if err != nil {
		return err
	}
	return nil
}

func (t *Token) DeleteByToken(plainText string) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": plainText})
	err := res.Delete()
	if err != nil {
		return err
	}
	return nil
}

func (t *Token) Insert(theToken Token, u User) error {
	collection := upper.Collection(t.Table())

	// delete existing tokens
	res := collection.Find(up.Cond{"user_id =": u.ID})
	err := res.Delete()
	if err != nil {
		return err
	}
	theToken.CreatedAt = time.Now()
	theToken.UpdatedAt = time.Now()
	theToken.FirstName = u.FirstName
	theToken.Email = u.Email

	_, err = collection.Insert(theToken)
	if err != nil {
		return err
	}
	return nil
}

func (t *Token) GenerateToken(userID int, ttl time.Duration) (*Token, error) {
	token := &Token{
		UserID:  userID,
		Expires: time.Now().Add(ttl),
	}
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]
	return token, nil
}

func (t *Token) Authenticate(r *http.Request) (*User, error) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return nil, errors.New("no authorization header provided")
	}

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("no authorization header provided")
	}

	token := headerParts[1]
	if len(token) != 26 {
		return nil, errors.New("token wrong size")
	}

	t, err := t.GetByToken(token)
	if err != nil {
		return nil, err
	}
	if t.Expires.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	user, err := t.GetUserForToken(token)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (t *Token) ValidToken(theToken string) (bool, error) {
	user, err := t.GetUserForToken(theToken)
	if err != nil {
		return false, errors.New("no matching user found")
	}
	if user.Token.PlainText == "" {
		return false, errors.New("no matching token found")
	}

	if user.Token.Expires.Before(time.Now()) {
		return false, errors.New("token expired")
	}
	return true, nil
}
