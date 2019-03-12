package logins

import (
	"database/sql"
	"fmt"
)

type MySQLStore struct {
	Client *sql.DB
}

func NewDBConnection(db *sql.DB) *MySQLStore {

	//create the data source name, which identifies the
	//user, password, server address, and default database
	//os.Getenv("MYSQL_ROOT_PASSWORD")

	store := MySQLStore{}
	store.Client = db

	return &store

}

//Insert puts a user into the database
func (db *MySQLStore) Insert(login *Login) (*Login, error) {

	insq := "insert into logins( user_id, login_time, user_ip) values (?,?,?)"
	res, err := db.Client.Exec(insq, login.UserID, login.LoginTime, login.UserIP)

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

	login.Key = id
	return login, nil

}
