package logins

//Store represents a store for Logins
type Store interface {
	//GetByID returns the User with the given ID
	Insert(*Login) (*Login, error)
}
