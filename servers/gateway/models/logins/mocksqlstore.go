package logins

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
func (db *MockStore) Insert(login *Login) (*Login, error) {

	/*
		//Case user already exists
		if user.ID == 0 {
			return nil, fmt.Errorf("User with id %d already exists", user.ID)
			//Database error
		} else if user.ID == 1 {
			return nil, fmt.Errorf("Error inserting user into database")
		}
	*/

	return login, nil

}
