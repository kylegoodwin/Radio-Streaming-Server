package users

import (
	"database/sql"
	"regexp"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

/* This will be used on a real connection I think
//data source name
dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/mysql441")

//db object
db, err := sql.Open("mysql", dsn)
if err != nil {
	fmt.Printf("Error opening the database: %v", err)
	os.Exit(1)
}

//When comeplete, close the db
defer db.Close()

//create a live connection to the db
if err := db.Ping(); err != nil {
	fmt.Printf("error pinging database: %v\n", err)
} else {
	fmt.Printf("successfully connected!\n")
}
*/

func TestMSSInsert(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	//Create test user
	newUser := NewUser{
		"test@test.com",
		"hunter2",
		"hunter2",
		"AzureDiamond",
		"John",
		"Johnson",
	}
	user, err := newUser.ToUser()
	if err != nil {
		t.Errorf("Failed creating the test user")
	}

	mss := NewMySQLStore(db)

	//Run success case
	mock.ExpectExec(regexp.QuoteMeta("insert into users(email, pass_hash, user_name, first_name, last_name, photo_URL) values (?,?,?,?,?,?)")).WithArgs(user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL).WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err = mss.Insert(user); err != nil {
		t.Errorf("Encountered an unexpected error during Insert: %v", err)
	}

	//Run error case

	brokenUser := User{
		0,
		"wrong@wrong.com",
		[]byte("wrong"),
		"imadummy",
		"wrong",
		"wrongson",
		"http://notcorrect.com",
	}
	mock.ExpectExec(regexp.QuoteMeta("insert into users(email, pass_hash, user_name, first_name, last_name, photo_URL) values (?,?,?,?,?,?)")).WithArgs(brokenUser.Email, brokenUser.PassHash, brokenUser.UserName, brokenUser.FirstName, brokenUser.LastName, brokenUser.PhotoURL).WillReturnError(sql.ErrNoRows)
	if _, err = mss.Insert(&brokenUser); err == nil {
		t.Errorf("Did not encounter an error during malformed Insert: %v", err)
	}

	//Final check for expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestMSSGet(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	//Create test user
	newUser := NewUser{
		"test@test.com",
		"hunter2",
		"hunter2",
		"AzureDiamond",
		"John",
		"Johnson",
	}
	user, err := newUser.ToUser()
	if err != nil {
		t.Errorf("Failed creating the test user")
	}

	//Create mysqlstore object
	mss := NewMySQLStore(db)

	//Test failed creation
	notMss := NewMySQLStore(nil)
	if notMss != nil {
		t.Errorf("NewMySQLStore did not return nil when it was supposed to fail")
	}

	//=================PASSING TESTS===================
	//Expect GetByEmail and run function
	row1 := sqlmock.NewRows([]string{"id", "email", "pass_hash", "user_name", "first_name", "last_name", "photo_URL"}).AddRow(1, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
	mock.ExpectQuery(regexp.QuoteMeta("select * from users where email=?")).WithArgs(user.Email).WillReturnRows(row1)

	//Need to grab the DB generated ID for the func test of GetByID
	_, err = mss.GetByEmail(user.Email)
	if err != nil {
		t.Errorf("Encountered an unexpected error during GetByEmail: %v", err)
	}

	//Expect GetByUserName and run function
	row2 := sqlmock.NewRows([]string{"id", "email", "pass_hash", "user_name", "first_name", "last_name", "photo_URL"}).AddRow(1, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
	mock.ExpectQuery(regexp.QuoteMeta("select * from users where user_name=?")).WithArgs(user.UserName).WillReturnRows(row2)

	_, err = mss.GetByUserName(user.UserName)
	if err != nil {
		t.Errorf("Encountered an unexpected error during GetByUserName: %v", err)
	}

	//Expect GetByID and run function
	row3 := sqlmock.NewRows([]string{"id", "email", "pass_hash", "user_name", "first_name", "last_name", "photo_URL"}).AddRow(1, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
	mock.ExpectQuery(regexp.QuoteMeta("select * from users where id=?")).WithArgs(1).WillReturnRows(row3)

	_, err = mss.GetByID(1)
	if err != nil {
		t.Errorf("Encountered an unexpected error during GetByUserName: %v", err)
	}

	//========================ERRORS=========================
	//Expect GetByEmail and run function
	//row4 := sqlmock.NewRows([]string{"id", "email", "pass_hash", "user_name", "first_name", "last_name", "photo_URL"}).AddRow(1, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
	mock.ExpectQuery(regexp.QuoteMeta("select * from users where email=?")).WithArgs("nonexistent@email.com").WillReturnError(sql.ErrNoRows)

	//Need to grab the DB generated ID for the func test of GetByID
	//malformed query args
	_, err = mss.GetByEmail("nonexistent@email.com")
	if err == nil {
		t.Errorf("Did not encounter an error during malformed GetByEmail: %v", err)
	}

	//Expect GetByUserName and run function
	//row5 := sqlmock.NewRows([]string{"id", "email", "pass_hash", "user_name", "first_name", "last_name", "photo_URL"}).AddRow(1, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
	mock.ExpectQuery(regexp.QuoteMeta("select * from users where user_name=?")).WithArgs("nonexistentusername").WillReturnError(sql.ErrNoRows)

	_, err = mss.GetByUserName("nonexistentusername")
	if err == nil {
		t.Errorf("Did not encounter an error during malformed GetByUserName: %v", err)
	}

	//Expect GetByID and run function
	//row6 := sqlmock.NewRows([]string{"id", "email", "pass_hash", "user_name", "first_name", "last_name", "photo_URL"}).AddRow(1, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
	mock.ExpectQuery(regexp.QuoteMeta("select * from users where id=?")).WithArgs(30).WillReturnError(sql.ErrNoRows)

	_, err = mss.GetByID(30)
	if err == nil {
		t.Errorf("Did not encounter an error during malformed GetByUserName: %v", err)
	}

	//Final check
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestMSSUpdate(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	//Create test user
	newUser := NewUser{
		"test@test.com",
		"hunter2",
		"hunter2",
		"AzureDiamond",
		"John",
		"Johnson",
	}
	user, err := newUser.ToUser()
	if err != nil {
		t.Errorf("Failed creating the test user")
	}

	//Create mysqlstore object
	mss := NewMySQLStore(db)

	//Create successful Updates object
	successfulUpdates := Updates{
		user.FirstName,
		user.LastName,
	}
	//Create Expectation and run test
	row4 := sqlmock.NewRows([]string{"id", "email", "pass_hash", "user_name", "first_name", "last_name", "photo_URL"}).AddRow(1, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
	mock.ExpectExec(regexp.QuoteMeta("update users set first_name=? last_name=? where id=?")).WithArgs(successfulUpdates.FirstName, successfulUpdates.LastName, 1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("select * from users where id=?")).WithArgs(1).WillReturnRows(row4)

	_, err = mss.Update(1, &successfulUpdates)
	if err != nil {
		t.Errorf("Encountered an unexpected error during Update: %v", err)
	}

	//Create Error Expectation and run test
	mock.ExpectExec(regexp.QuoteMeta("update users set first_name=? last_name=? where id=?")).WithArgs(successfulUpdates.FirstName, successfulUpdates.LastName, 0).WillReturnError(sql.ErrNoRows)

	_, err = mss.Update(0, &successfulUpdates)
	if err == nil {
		t.Errorf("Did not encounter an error during malformed Update: %v", err)
	}

	//Final check
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestMMSDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	//Success case
	mock.ExpectExec(regexp.QuoteMeta("delete from users where id=?")).WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

	//Run the actual function
	mss := NewMySQLStore(db)
	if err = mss.Delete(1); err != nil {
		t.Errorf("Encountered an unexpected error during Delete: %v", err)
	}

	//Error case
	mock.ExpectExec(regexp.QuoteMeta("delete from users where id=?")).WithArgs(0).WillReturnError(sql.ErrNoRows)

	//Run the actual function
	if err = mss.Delete(0); err == nil {
		t.Errorf("Did not encounter an error during malformed Delete: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
