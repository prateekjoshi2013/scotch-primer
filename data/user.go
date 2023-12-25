package data

import (
	"errors"
	"time"

	up "github.com/upper/db/v4"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `db:"id,omitempty"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Active    int       `db:"user_active"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Token     Token     `db:"-"`
}

func (u *User) Table() string {
	return "users"
}

func (u *User) GetAll() ([]*User, error) {
	collection := upper.Collection(u.Table())
	var users []*User
	err := collection.Find().OrderBy("last_name").All(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (u *User) GetByEmail(email string) (*User, error) {
	var theUser User
	collection := upper.Collection(u.Table())
	err := collection.Find(up.Cond{"email =": email}).One(&theUser)
	if err != nil {
		return nil, err
	}
	var token Token
	collection = upper.Collection(token.Table())
	err = collection.Find(up.Cond{"user_id =": theUser.ID, "expiry >": time.Now()}).OrderBy("created_at desc").One(&token)
	if err != nil {
		if err != up.ErrNilRecord && err != up.ErrNoMoreRows {
			return nil, err
		}
	}
	theUser.Token = token
	return &theUser, nil
}

func (u *User) Get(id int) (*User, error) {
	var theUser User
	collection := upper.Collection(u.Table())
	err := collection.Find(up.Cond{"id =": id}).One(&theUser)
	if err != nil {
		return nil, err
	}
	var token Token
	collection = upper.Collection(token.Table())
	err = collection.Find(up.Cond{"user_id =": theUser.ID, "expiry >": time.Now()}).OrderBy("created_at desc").One(&token)
	if err != nil {
		if err != up.ErrNilRecord && err != up.ErrNoMoreRows {
			return nil, err
		}
	}
	theUser.Token = token
	return &theUser, nil

}

func (u *User) Update(theUser *User) error {
	theUser.UpdatedAt = time.Now()
	collection := upper.Collection(u.Table())
	res := collection.Find(theUser.ID)
	err := res.Update(theUser)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) Delete(id int) error {
	collection := upper.Collection(u.Table())
	res := collection.Find(id)
	err := res.Delete()
	if err != nil {
		return err
	}
	return nil
}

func (u *User) Insert(theUser User) (int, error) {
	newHash, err := bcrypt.GenerateFromPassword([]byte(theUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	theUser.Password = string(newHash)
	theUser.CreatedAt = time.Now()
	theUser.UpdatedAt = time.Now()
	collection := upper.Collection(u.Table())
	res, err := collection.Insert(theUser)
	if err != nil {
		return 0, err
	}
	id := getInsertID(res.ID())
	return id, nil
}

func (u *User) ResetPassword(id int, password string) error {
	newHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	theUser, err := u.Get(id)
	if err != nil {
		return err
	}
	u.Password = string(newHash)
	err = theUser.Update(u)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) PasswordMatches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}
