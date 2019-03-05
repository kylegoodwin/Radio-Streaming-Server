package users

import (
	"fmt"
)

//MyMockStore is a fake MySqlStore for testing purposes
type MyMockStore struct {
}

// GetByID will work like i want it to
func (ms *MyMockStore) GetByID(id int64) (*User, error) {
	if id == int64(2) {
		return nil, fmt.Errorf("Error getting user with id: %d", id)
	}
	newUser := &NewUser{
		Email:        "test@test.com",
		Password:     "hunter2",
		PasswordConf: "hunter2",
		UserName:     "AzureDiamond",
		FirstName:    "John",
		LastName:     "Johnson",
	}
	user, err := newUser.ToUser()
	if err != nil {
		return nil, fmt.Errorf("Error converting to user in successful test case")
	}
	user.ID = int64(1)

	return user, nil
}

// GetByEmail will work like i want it to
func (ms *MyMockStore) GetByEmail(email string) (*User, error) {
	if email == "wrong@test.com" {
		return nil, fmt.Errorf("Error getting user with id: %v", email)
	}
	newUser := NewUser{
		Email:        "test@test.com",
		Password:     "hunter2",
		PasswordConf: "hunter2",
		UserName:     "AzureDiamond",
		FirstName:    "John",
		LastName:     "Johnson",
	}
	user, err := newUser.ToUser()
	if err != nil {
		return nil, fmt.Errorf("Error  converting newuser to user")
	}
	return user, nil
}

// Insert will work like i want it to
func (ms *MyMockStore) Insert(user *User) (*User, error) {
	if user.FirstName == "Error" {
		return nil, fmt.Errorf("Error Inserting New User")
	}
	if user.LastName != "Error" {
		user.ID = int64(1)
	}
	// Assumes that if user without the FirstName field being equal to "Error" will
	// always result in a successful insert
	return user, nil
}

// Update will work like i want it to
func (ms *MyMockStore) Update(id int64, updates *Updates) (*User, error) {

	if updates.FirstName == "Error" {
		return nil, fmt.Errorf("Error updating user")
	}
	newUser := NewUser{
		Email:        "test@test.com",
		Password:     "hunter2",
		PasswordConf: "hunter2",
		UserName:     "AzureDiamond",
		FirstName:    "Test",
		LastName:     "Johnson",
	}
	user, err := newUser.ToUser()
	if err != nil {
		return nil, fmt.Errorf("Error  converting newuser to user")
	}

	//Success case
	return user, nil
}
