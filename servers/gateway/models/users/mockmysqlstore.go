package users

import "github.com/Radio-Streaming-Server/servers/gateway/indexes"

type MockStore struct {
}

//NewFakeConnection creates a new DB connection
func NewFakeConnection() *MockStore {

	//create the data source name, which identifies the
	//user, password, server address, and default database
	//os.Getenv("MYSQL_ROOT_PASSWORD")

	store := MockStore{}

	return &store

}

//Insert puts a user into the database
func (db *MockStore) Insert(user *User) (*User, error) {

	/*
		//Case user already exists
		if user.ID == 0 {
			return nil, fmt.Errorf("User with id %d already exists", user.ID)
			//Database error
		} else if user.ID == 1 {
			return nil, fmt.Errorf("Error inserting user into database")
		}
	*/

	return user, nil

}

//GetByID returns the User with the given ID
func (db *MockStore) GetByID(id int64) (*User, error) {

	validNewUser := NewUser{
		Email:        "coolguy@gmail.com",
		Password:     "password",
		PasswordConf: "password",
		UserName:     "coolguy",
		FirstName:    "Cool",
		LastName:     "Guy",
	}

	validUser, _ := validNewUser.ToUser()

	return validUser, nil

}

//GetByEmail returns the User with the given email
func (db *MockStore) GetByEmail(email string) (*User, error) {

	validNewUser := NewUser{
		Email:        "coolguy@gmail.com",
		Password:     "password",
		PasswordConf: "password",
		UserName:     "coolguy",
		FirstName:    "Cool",
		LastName:     "Guy",
	}

	validUser, _ := validNewUser.ToUser()

	return validUser, nil

}

//GetByUserName returns the User with the given Username
func (db *MockStore) GetByUserName(username string) (*User, error) {
	validUser := User{
		ID:        1,
		Email:     "Anything you want",
		PassHash:  []byte("Anything you want"),
		FirstName: "Anything you want",
		LastName:  "Anything you want",
		PhotoURL:  "Anything you want",
	}

	return &validUser, nil

}

//Update applies UserUpdates to the given user ID
//and returns the newly-updated user
func (db *MockStore) Update(id int64, updates *Updates) (*User, error) {
	validUser := User{
		ID:        1,
		Email:     "coolguy@gmail.com",
		UserName:  "coolguy",
		FirstName: "New",
		LastName:  "Name",
	}
	validUser.SetPassword("password")

	return &validUser, nil

}

//Delete deletes the user with the given ID
func (db *MockStore) Delete(id int64) error {

	return nil

}

func (db *MockStore) BuildTrie() (*indexes.Trie, error) {
	return indexes.NewTrie(), nil
}
