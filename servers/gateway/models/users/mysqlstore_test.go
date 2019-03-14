package users

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

const INSERTSTATEMENT = "insert into users( email,password_hash, username, first_name, last_name, photo_url) values (?,?,?,?,?,?)"
const GETSTATEMENT = "Select * From users Where id=?"

func TestMySQLStore_GetByID(t *testing.T) {
	//create a new sql mock
	db, mock, err := sqlmock.New()

	if err != nil {
		log.Fatalf("error creating sql mock: %v", err)
	}

	//ensure it's closed at the end of the test
	defer db.Close()

	// Initialize a user struct we will use as a test variable
	expectedUser := &User{
		ID:        2,
		Email:     "me@gmail.com",
		PassHash:  []byte("passhash"),
		UserName:  "coolguy",
		FirstName: "Cool",
		LastName:  "Guy",
		PhotoURL:  "website.com",
	}

	store := NewDBConnection(db)

	row := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "first_name", "last_name", "photo_url"})
	row.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)

	mock.ExpectQuery("Select (.+) From users Where ").
		WithArgs(expectedUser.ID).WillReturnRows(row)

	user, err := store.GetByID(expectedUser.ID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err == nil && !reflect.DeepEqual(user, expectedUser) {
		t.Errorf("User queried does not match expected user")
	}

	mock.ExpectQuery("Select (.+) From users Where ").
		WithArgs(-1).WillReturnError(sql.ErrNoRows)

	if _, err = store.GetByID(-1); err == nil {
		t.Errorf("Expected error: %v, but recieved nil", sql.ErrNoRows)
	}

	// Attempting to trigger a DBMS querying error
	queryingErr := fmt.Errorf("DBMS error when querying")
	mock.ExpectQuery("Select (.+) From users Where ").
		WithArgs(expectedUser.ID).WillReturnError(queryingErr)

	if _, err = store.GetByID(expectedUser.ID); err == nil {
		t.Errorf("Expected error: %v, but recieved nil", queryingErr)
	}

	// This attempts to check if there are any expectations that we haven't met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet sqlmock expectations: %v", err)
	}

}

func TestMySQLStore_GetByEmail(t *testing.T) {
	//create a new sql mock
	db, mock, err := sqlmock.New()

	if err != nil {
		log.Fatalf("error creating sql mock: %v", err)
	}

	//ensure it's closed at the end of the test
	defer db.Close()

	// Initialize a user struct we will use as a test variable
	expectedUser := &User{
		ID:        2,
		Email:     "me@gmail.com",
		PassHash:  []byte("passhash"),
		UserName:  "coolguy",
		FirstName: "Cool",
		LastName:  "Guy",
		PhotoURL:  "website.com",
	}

	// Initialize a MySQLStore struct to allow us to interface with the SQL client
	store := NewDBConnection(db)

	row := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "first_name", "last_name", "photo_url"})
	row.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)

	mock.ExpectQuery(regexp.QuoteMeta("select * from users where email = ?")).
		WithArgs(expectedUser.Email).WillReturnRows(row)

	user, err := store.GetByEmail(expectedUser.Email)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err == nil && !reflect.DeepEqual(user, expectedUser) {
		t.Errorf("User queried does not match expected user")
	}

	row = sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "first_name", "last_name", "photo_url"})
	row.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)

	mock.ExpectQuery(regexp.QuoteMeta("select * from users where email = ?")).
		WithArgs(" ").WillReturnError(sql.ErrNoRows)

	if _, err = store.GetByEmail(" "); err == nil {
		t.Errorf("Expected error: %v, but recieved nil", sql.ErrNoRows)
	}

	row = sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "first_name", "last_name", "photo_url"})
	row.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)

	// Attempting to trigger a DBMS querying error
	queryingErr := fmt.Errorf("DBMS error when querying")

	mock.ExpectQuery(regexp.QuoteMeta("select * from users where email = ?")).
		WithArgs(expectedUser.Email).WillReturnError(queryingErr)

	if _, err = store.GetByEmail(expectedUser.Email); err == nil {
		t.Errorf("Expected error: %v, but recieved nil", queryingErr)
	}

	// This attempts to check if there are any expectations that we haven't met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet sqlmock expectations: %v", err)
	}

}

func TestMySQLStore_GetByUsername(t *testing.T) {
	//create a new sql mock
	db, mock, err := sqlmock.New()

	if err != nil {
		log.Fatalf("error creating sql mock: %v", err)
	}

	//ensure it's closed at the end of the test
	defer db.Close()

	// Initialize a user struct we will use as a test variable
	expectedUser := &User{
		ID:        2,
		Email:     "me@gmail.com",
		PassHash:  []byte("passhash"),
		UserName:  "coolguy",
		FirstName: "Cool",
		LastName:  "Guy",
		PhotoURL:  "website.com",
	}

	// Initialize a MySQLStore struct to allow us to interface with the SQL client
	store := NewDBConnection(db)

	// Create a row with the appropriate fields in your SQL database
	// Add the actual values to the row
	row := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "first_name", "last_name", "photo_url"})
	row.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)

	// Expecting a successful "query"
	// This tells our db to expect this query (id) as well as supply a certain response (row)
	// REMINDER: Since sqlmock requires a regex string, in order for `?` to be interpreted, you'll
	// have to wrap it within a `regexp.QuoteMeta`. Be mindful that you will need to do this EVERY TIME you're
	// using any reserved metacharacters in regex.
	mock.ExpectQuery(regexp.QuoteMeta("select * from users where username= ?")).
		WithArgs(expectedUser.UserName).WillReturnRows(row)

	// Since we know our query is successful, we want to test whether there happens to be
	// any expected error that may occur.
	user, err := store.GetByUserName(expectedUser.UserName)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Again, since we are assuming that our query is successful, we can test for when our
	// function doesn't work as expected.
	if err == nil && !reflect.DeepEqual(user, expectedUser) {
		t.Errorf("User queried does not match expected user")
	}

	row = sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "first_name", "last_name", "photo_url"})
	row.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)

	// Expecting a unsuccessful "query"
	// Attempting to search by an id that doesn't exist. This would result in a
	// sql.ErrNoRows error
	// REMINDER: Using a constant makes your code much clear, and is highly recommended.
	mock.ExpectQuery(regexp.QuoteMeta("select * from users where username= ?")).
		WithArgs(" ").WillReturnError(sql.ErrNoRows)

	// Since we are expecting an error here, we create a condition opposing that to see
	// if our GetById is working as expected
	if _, err = store.GetByUserName(" "); err == nil {
		t.Errorf("Expected error: %v, but recieved nil", sql.ErrNoRows)
	}

	row = sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "first_name", "last_name", "photo_url"})
	row.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)

	// Attempting to trigger a DBMS querying error
	queryingErr := fmt.Errorf("DBMS error when querying")

	mock.ExpectQuery(regexp.QuoteMeta("select * from users where username= ?")).
		WithArgs(expectedUser.UserName).WillReturnError(queryingErr)

	if _, err = store.GetByUserName(expectedUser.UserName); err == nil {
		t.Errorf("Expected error: %v, but recieved nil", queryingErr)
	}

	// This attempts to check if there are any expectations that we haven't met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet sqlmock expectations: %v", err)
	}

}

func TestInsert(t *testing.T) {

	db, mock, err := sqlmock.New()

	if err != nil {
		t.Errorf("Error setting up mock database connection")
	}

	defer db.Close()

	expectedUser := &User{
		ID:        2,
		Email:     "me@gmail.com",
		PassHash:  []byte("passhash"),
		UserName:  "coolguy",
		FirstName: "Cool",
		LastName:  "Guy",
		PhotoURL:  "website.com",
	}

	store := NewDBConnection(db)

	// This tells our db to expect an insert query with certain arguments with a certain
	// return result
	mock.ExpectExec(regexp.QuoteMeta(INSERTSTATEMENT)).
		WithArgs(expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL).
		WillReturnResult(sqlmock.NewResult(2, 1))

	user, err := store.Insert(expectedUser)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if err == nil && !reflect.DeepEqual(user, expectedUser) {
		t.Errorf("User returned does not match input user")
	}

	// Inserting an invalid user
	invalidUser := &User{
		ID:        -1,
		Email:     "",
		PassHash:  []byte("passhash"),
		UserName:  "test123",
		FirstName: "Cool",
		LastName:  "Guy",
		PhotoURL:  "website.com",
	}
	insertErr := fmt.Errorf("Error executing INSERT operation")
	mock.ExpectExec(regexp.QuoteMeta(INSERTSTATEMENT)).
		WithArgs(invalidUser.Email, invalidUser.PassHash, invalidUser.UserName, invalidUser.FirstName, invalidUser.LastName, invalidUser.PhotoURL).
		WillReturnError(insertErr)

	if _, err = store.Insert(invalidUser); err == nil {
		t.Errorf("Expected error: %v but recieved nil", insertErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet sqlmock expectations: %v", err)
	}

}

/*
func TestMySQLStore_TestUpdate(t *testing.T) {

	expectedUser := &User{
		ID:        2,
		Email:     "me@gmail.com",
		PassHash:  []byte("passhash"),
		UserName:  "coolguy",
		FirstName: "Cool",
		LastName:  "Guy",
		PhotoURL:  "website.com",
	}

	update := &Updates{
		LastName:  "Testname",
		FirstName: "Kyle",
	}

	newUser := &User{
		ID:        2,
		Email:     "me@gmail.com",
		PassHash:  []byte("passhash"),
		UserName:  "coolguy",
		FirstName: "Kyle",
		LastName:  "Testname",
		PhotoURL:  "website.com",
	}

		db, mock, err := sqlmock.New()

		if err != nil {
			t.Errorf("Error setting up mock database connection")
		}

		defer db.Close()

		row := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "first_name", "last_name", "photo_url"})
		row.AddRow(expectedUser.ID, expectedUser.Email, expectedUser.PassHash, expectedUser.UserName, expectedUser.FirstName, expectedUser.LastName, expectedUser.PhotoURL)


			store := NewDBConnection(db)

			mock.ExpectPrepare(regexp.QuoteMeta("update users set first_name=? , last_name=? where id =?"))
			mock.ExpectExec(regexp.QuoteMeta("update users set first_name=? , last_name=? where id =?")).
				WithArgs(update.FirstName, update.LastName, expectedUser.ID).WillReturnResult(sqlmock.NewResult(2, 1))
			mock.ExpectQuery("Select (.+) From users Where ").
				WithArgs(expectedUser.ID).WillReturnRows(row)

			var user *User
			user, err = store.Update(2, update)

			if err != nil {
				t.Errorf("Error retunred from update statment %v", err.Error())
			}



		if err == nil && !reflect.DeepEqual(user, newUser) {
			fmt.Println(user)
			fmt.Println(newUser)
			t.Errorf("User returned does not match input user")
		}


	fmt.Println(newUser)
	//fmt.Println(user)


}
*/
