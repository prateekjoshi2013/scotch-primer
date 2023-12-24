package data

import (
	"fmt"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	db2 "github.com/upper/db/v4"
)

func TestNew(t *testing.T) {
	fakeDB, _, _ := sqlmock.New()

	defer fakeDB.Close()

	_ = os.Setenv("DATABASE_TYPE", "postgres")
	m := New(fakeDB)
	if fmt.Sprintf("%T", m) != "data.Models" {
		t.Errorf("expected data.Models, got %T", m)
	}

	_ = os.Setenv("DATABASE_TYPE", "mysql")
	m = New(fakeDB)
	if fmt.Sprintf("%T", m) != "data.Models" {
		t.Errorf("expected data.Models, got %T", m)
	}

}

func TestGetInsertID(t *testing.T) {
	var id db2.ID
	// postgred ID conversion test
	id = int64(1)
	returnedID := getInsertID(id)
	if fmt.Sprintf("%T", returnedID) != "int" {
		t.Error("wrong type returned")
	}
	// mariadb/mysql ID conversion test
	id = 1
	returnedID = getInsertID(id)
	if fmt.Sprintf("%T", returnedID) != "int" {
		t.Error("wrong type returned")
	}
}


