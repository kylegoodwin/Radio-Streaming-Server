package users

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/assignments-zanewebbUW/servers/gateway/indexes"
)

//MySQLStore is a struct that holds a *sql.DB as Client
type MySQLStore struct {
	Client *sql.DB
	Trie   *indexes.Trie
}

/*
func main() {

	//data source name
	dsn := fmt.Sprintf("root@/blog")

	//db object
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("Error opening the database: %v", err)
		os.Exit(1)
	}

	//When comeplete, close the db
	defer db.Close()

	//create a live connection to the db
	// if err := db.Ping(); err != nil {
	// 	fmt.Printf("error pinging database: %v\n", err)
	// } else {
	// 	fmt.Printf("successfully connected!\n")
	// }
}
*/

//NewMySQLStore generates a MySQLStore struct and returns it
func NewMySQLStore(db *sql.DB) *MySQLStore {
	//Thank Mr.TA
	if db != nil {
		return &MySQLStore{
			Client: db,
			Trie:   indexes.NewTrie(),
		}
	}

	return nil
}

//PopulateTrie pulls users from the store and populates them into the Trie
func (mss *MySQLStore) PopulateTrie() error {
	//Grab these fields from all users
	rows, err := mss.Client.Query("select id,user_name,first_name,last_name from users")
	defer rows.Close()
	if err != nil {
		return err
	}

	//For each user in the DB, insert the correct key pairs for that user
	for rows.Next() {
		//Grab the user's info
		tempUser := &User{}
		if err := rows.Scan(&tempUser.ID, &tempUser.UserName, &tempUser.FirstName, &tempUser.LastName); err != nil {
			return errors.New("Failed while parsing rows")
		}

		//Add the user into the Trie
		mss.InsertUserIntoTrie(tempUser)
	}

	return nil
}

//InsertUserIntoTrie adds a user to the sqlstore trie using the ID, UserName, FirstName, and LastName fields
func (mss *MySQLStore) InsertUserIntoTrie(user *User) {
	//Declare insertion slice
	insertionSlice := []string{}
	//Break the fields up if they have spaces
	//UserName
	user.UserName = strings.ToLower(user.UserName)
	if strings.Contains(user.UserName, " ") {
		insertionSlice = append(insertionSlice, strings.Split(user.UserName, " ")...)
	} else {
		insertionSlice = append(insertionSlice, user.UserName)
	}
	//FirstName
	user.FirstName = strings.ToLower(user.FirstName)
	if strings.Contains(user.FirstName, " ") {
		insertionSlice = append(insertionSlice, strings.Split(user.FirstName, " ")...)
	} else {
		insertionSlice = append(insertionSlice, user.FirstName)
	}
	//LastName
	user.LastName = strings.ToLower(user.LastName)
	if strings.Contains(user.LastName, " ") {
		insertionSlice = append(insertionSlice, strings.Split(user.LastName, " ")...)
	} else {
		insertionSlice = append(insertionSlice, user.LastName)
	}

	//Insert into the Trie
	for _, i := range insertionSlice {
		mss.Trie.Add(i, user.ID)
	}
}

//GetByID returns the User with the given ID
func (mss *MySQLStore) GetByID(id int64) (*User, error) {
	sqlcmd := "select * from users where id=?"
	user := User{}
	err := mss.Client.QueryRow(sqlcmd, id).Scan(&user.ID, &user.Email, &user.PassHash, &user.UserName, &user.FirstName, &user.LastName, &user.PhotoURL)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//GetByEmail returns the User with the given email
func (mss *MySQLStore) GetByEmail(email string) (*User, error) {
	sqlcmd := "select * from users where email=?"
	user := User{}
	err := mss.Client.QueryRow(sqlcmd, email).Scan(&user.ID, &user.Email, &user.PassHash, &user.UserName, &user.FirstName, &user.LastName, &user.PhotoURL)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//GetByUserName returns the User with the given Username
func (mss *MySQLStore) GetByUserName(username string) (*User, error) {
	sqlcmd := "select * from users where user_name=?"
	user := User{}
	err := mss.Client.QueryRow(sqlcmd, username).Scan(&user.ID, &user.Email, &user.PassHash, &user.UserName, &user.FirstName, &user.LastName, &user.PhotoURL)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//Insert inserts the user into the database, and returns
//the newly-inserted User, complete with the DBMS-assigned ID
func (mss *MySQLStore) Insert(user *User) (*User, error) {
	insq := "insert into users(email, pass_hash, user_name, first_name, last_name, photo_URL) values (?,?,?,?,?,?)"
	res, err := mss.Client.Exec(insq, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
	if err != nil {
		return nil, err
	}

	//get generated ID from insert
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	//apply it to the user struct and return it
	user.ID = id
	return user, nil
}

//InsertSignIn logs a new successful sign in
func (mss *MySQLStore) InsertSignIn(userID int64, ip string) (int64, error) {
	insq := "insert into successful_logins(user_id, sign_in_time, login_ip) values (?,now(),?)"
	res, err := mss.Client.Exec(insq, userID, ip)
	if err != nil {
		return int64(0), err
	}

	//get generated ID from insert
	id, err := res.LastInsertId()
	if err != nil {
		return int64(0), err
	}
	return id, nil
}

//Update applies UserUpdates to the given user ID
//and returns the newly-updated user
func (mss *MySQLStore) Update(id int64, updates *Updates) (*User, error) {

	sqlcmd := "update users set first_name=? last_name=? where id=?"

	_, err := mss.Client.Exec(sqlcmd, updates.FirstName, updates.LastName, id)
	if err != nil {
		return nil, err
	}

	//Just to properly return an updated user. Probably not optimal
	user, err := mss.GetByID(id)
	user.ApplyUpdates(updates)

	return user, nil
}

//Delete deletes the user with the given ID
func (mss *MySQLStore) Delete(id int64) error {
	sqlcmd := "delete from users where id=?"
	_, err := mss.Client.Exec(sqlcmd, id)
	if err != nil {
		return err
	}
	return nil
}
