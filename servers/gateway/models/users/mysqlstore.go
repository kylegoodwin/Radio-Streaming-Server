package users

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Radio-Streaming-Server/servers/gateway/indexes"
	_ "github.com/go-sql-driver/mysql"
)

//MySQLStore is an abstraction of the MySQL database
type MySQLStore struct {
	Client *sql.DB
}

//NewDBConnection creates a new DB connection
func NewDBConnection(db *sql.DB) *MySQLStore {

	//create the data source name, which identifies the
	//user, password, server address, and default database
	//os.Getenv("MYSQL_ROOT_PASSWORD")

	store := MySQLStore{}
	store.Client = db

	return &store

}

//Insert puts a user into the database
func (db *MySQLStore) Insert(user *User) (*User, error) {

	insq := "insert into users( email,password_hash, username, first_name, last_name, photo_url) values (?,?,?,?,?,?)"
	res, err := db.Client.Exec(insq, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)

	if err != nil {
		fmt.Printf("error inserting new row: %v\n", err)
		return nil, err
	}
	//get the auto-assigned ID for the new row
	id, err := res.LastInsertId()

	if err != nil {
		fmt.Printf("error getting new ID: %v\n", id)
		return nil, err
	}

	user.ID = id
	return user, nil

}

//Note that you can probably refactor this code to only have one method that sends the query

//GetByID returns the user with the associated id from the database
func (db *MySQLStore) GetByID(id int64) (*User, error) {

	selectStatement := "Select * From users Where id=?"
	row := db.Client.QueryRow(selectStatement, id)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.PassHash, &user.UserName, &user.FirstName, &user.LastName, &user.PhotoURL); err != nil {
		fmt.Printf("error scanning row: %v\n", err)
		return nil, fmt.Errorf("DBMS error when querying")
	}

	return &user, nil

}

//GetByEmail returns the User with the given email
func (db *MySQLStore) GetByEmail(email string) (*User, error) {

	selectStatement := "select * from users where email = ?"
	rows, err := db.Client.Query(selectStatement, email)

	if err != nil {
		fmt.Printf("error with db query: %v\n", err.Error())
		return nil, err

	}

	user := User{}

	for rows.Next() {
		if err := rows.Scan(&user.ID, &user.Email, &user.PassHash, &user.UserName, &user.FirstName, &user.LastName, &user.PhotoURL); err != nil {
			fmt.Printf("error scanning row: %v\n", err)
			return nil, err
		}
	}

	return &user, nil

}

//GetByUserName returns the User with the given Username
func (db *MySQLStore) GetByUserName(username string) (*User, error) {

	selectStatement := "select * from users where username= ?"
	row := db.Client.QueryRow(selectStatement, username)

	user := User{}

	if err := row.Scan(&user.ID, &user.Email, &user.PassHash, &user.UserName, &user.FirstName, &user.LastName, &user.PhotoURL); err != nil {
		fmt.Printf("error scanning row: %v\n", err)
		return nil, fmt.Errorf("DBMS error when querying")
	}

	return &user, nil

}

//Update applies UserUpdates to the given user ID
//and returns the newly-updated user
func (db *MySQLStore) Update(id int64, updates *Updates) (*User, error) {

	_, err := db.Client.Exec("update users set first_name=? , last_name=?, photo_url=? where id =?", updates.FirstName, updates.LastName, updates.PhotoURL, id)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	var user *User
	user, err = db.GetByID(id)

	if err != nil {
		err = fmt.Errorf("Issue getting user after update from db")
		return nil, err
	}

	return user, nil

}

//Delete deletes the user with the given ID
func (db *MySQLStore) Delete(id int64) error {

	deleteStatement := "delete from users where id= ?"
	_, err := db.Client.Exec(deleteStatement, id)

	if err != nil {
		return err
	}

	return nil

}

func (db *MySQLStore) GetAllUsers() ([]*User, error) {

	selectStatement := "Select * From users"
	rows, err := db.Client.Query(selectStatement)

	if err != nil {
		fmt.Printf("error querying database %v\n", err)
		return nil, fmt.Errorf("DBMS error when querying")
	}

	var users []*User = []*User{}

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.PassHash, &user.UserName, &user.FirstName, &user.LastName, &user.PhotoURL); err != nil {
			fmt.Printf("error scanning row: %v\n", err)
			return nil, fmt.Errorf("DBMS error when querying")
		}
		users = append(users, &user)

	}

	return users, nil

}

func (db *MySQLStore) BuildTrie() (*indexes.Trie, error) {
	trie := indexes.NewTrie()

	selectStatement := "select * from users"
	rows, err := db.Client.Query(selectStatement)

	if err != nil {
		return nil, fmt.Errorf("Error querying database for users")
	}

	for rows.Next() {
		user := User{}
		if err := rows.Scan(&user.ID, &user.Email, &user.PassHash, &user.UserName, &user.FirstName, &user.LastName, &user.PhotoURL); err != nil {
			fmt.Printf("error scanning row: %v\n", err)
			return nil, fmt.Errorf("DBMS error when querying")
		}

		names := []string{}
		firstNames := strings.Split(strings.ToLower(user.FirstName), " ")
		lastNames := strings.Split(strings.ToLower(user.LastName), " ")
		names = append(names, firstNames...)
		names = append(names, lastNames...)

		for _, name := range names {
			trie.Add(name, user.ID)
		}

	}

	return trie, nil

}
