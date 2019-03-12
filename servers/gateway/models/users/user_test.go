package users

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

//TODO: add tests for the various functions in user.go, as described in the assignment.
//use `go test -cover` to ensure that you are covering all or nearly all of your code paths.
func TestValidateUser(t *testing.T) {

	cases := []struct {
		Case           string
		Email          string
		Password       string
		PasswordConf   string
		UserName       string
		FirstName      string
		LastName       string
		expectedOutput bool
	}{

		{"Valid user", "me@gmail.com", "password", "password", "coolguy", "Guy", "Named", false},

		{"Invalid email", "funfunfun", "fun", "fun", "coolguy", "Guy", "Named", true},

		{"Short password", "me@gmail.com", "pass", "pass", "coolguy", "Guy", "Named", true},

		{"Password mismatch", "me@gmail.com", "password", "pass", "coolguy", "Guy", "Named", true},

		{"Zero length username", "me@gmail.com", "password", "password", "", "Guy", "Named", true},

		{"Whitespace username", "me@gmail.com", "password", "password", "cool guy", "Guy", "Named", true},
	}

	for _, c := range cases {
		user := NewUser{c.Email, c.Password, c.PasswordConf, c.UserName, c.FirstName, c.LastName}

		if c.expectedOutput && user.Validate() == nil {
			t.Errorf("incorrect output for `%s`: expected `%s` but got `%s`", c.Case, "true", "No error was thrown")
		}

		if !c.expectedOutput && user.Validate() != nil {
			t.Errorf("incorrect output for `%s`: expected `%s` but got `%s`", c.Case, "true", user.Validate().Error())
		}

	}
}

func TestToUser(t *testing.T) {

	cases := []struct {
		Case         string
		Email        string
		Password     string
		PasswordConf string
		UserName     string
		FirstName    string
		LastName     string
		expectedURL  string
	}{

		{"Valid user", "me@gmail.com", "password", "password", "coolguy", "Guy", "Named", "https://www.gravatar.com/avatar/525ceb06bc8862932d853a033411e3b7"},
		{"Capital leters email", "ME@gmail.com", "password", "password", "coolguy", "Guy", "Named", "https://www.gravatar.com/avatar/525ceb06bc8862932d853a033411e3b7"},
		{"Spaces in begining of email", " me@gmail.com", "password", "password", "coolguy", "Guy", "Named", "https://www.gravatar.com/avatar/525ceb06bc8862932d853a033411e3b7"},
		//{"Email with spaces "}
	}

	for _, c := range cases {
		newUser := NewUser{c.Email, c.Password, c.PasswordConf, c.UserName, c.FirstName, c.LastName}
		user, err := newUser.ToUser()

		if err != nil {
			t.Errorf("Unexpected err while converting newuser to user. Case: `%s`", c.Case)
		}

		//Test that a password is generated and can be used to with the bycrpt validator

		err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(newUser.Password))
		if err != nil {
			t.Errorf("Unexpected err while comparing password hash and original password. Case: `%s`", c.Case)
		}

		//Test that the photourl is good, even with uppercase letters and spaces
		if user.PhotoURL != c.expectedURL {
			t.Errorf("incorrect output for `%s`: expected `%s` but got `%s`", c.Case, c.expectedURL, user.PhotoURL)
		}

	}
}

func TestFullName(t *testing.T) {

	cases := []struct {
		Case           string
		FirstName      string
		LastName       string
		expectedOutput string
	}{

		{"First and Last name", "Guy", "Dude", "Guy Dude"},
		{"First name only", "Guy", "", "Guy"},
		{"Last name only", "", "Dude", "Dude"},
		{"No name", "", "", ""},
	}

	for _, c := range cases {
		user := User{}
		user.FirstName = c.FirstName
		user.LastName = c.LastName

		if user.FullName() != c.expectedOutput {
			t.Errorf("Wrong output for `%s`: Got `%s` expected: `%s`", c.Case, user.FullName(), c.expectedOutput)
		}

	}

}

func TestUpdate(t *testing.T) {

	cases := []struct {
		Case           string
		FirstName      string
		LastName       string
		expectedOutput string
	}{

		{"Update first and last name", "New", "Name", "New Name"},
		{"Update first name", "New", "", "New"},
		{"Update last name", "", "New", "New"},
		{"Both updates blank", "", "", ""},
	}

	for _, c := range cases {
		user := User{}
		user.FirstName = "Guy"
		user.LastName = "Dude"

		update := Updates{}
		update.FirstName = c.FirstName
		update.LastName = c.LastName

		err := user.ApplyUpdates(&update)

		if err != nil {
			t.Errorf("Unexpected Error for case `%s` :`%s`", c.Case, err.Error())
		}

		if user.FullName() != c.expectedOutput {
			t.Errorf("Wrong output for `%s`: Got `%s` expected: `%s`", c.Case, user.FullName(), c.expectedOutput)
		}

	}

}

func TestAuth(t *testing.T) {

	cases := []struct {
		Case           string
		Password       string
		loginPassword  string
		expectedOutput bool
	}{

		{"Login with proper password", "password", "password", true},
		{"Login with incorrect password", "password", "badpassword", false},
		{"Login with empty password", "password", "", false},
	}

	for _, c := range cases {
		newUser := NewUser{}
		newUser.UserName = "Dude"
		newUser.Email = "me@gmail.com"
		newUser.Password = c.Password
		newUser.PasswordConf = c.Password
		user, err := newUser.ToUser()
		if err != nil {
			t.Errorf("Unexpected Error for case `%s` :`%s`", c.Case, err.Error())
		}

		//err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(c.loginPassword))

		err = user.Authenticate(c.loginPassword)

		if err != nil {
			if c.expectedOutput {
				t.Errorf("Login should succeded but failed: `%s`: Got `%s` expected: `%s`", c.Case, "true", "false")
			}
		} else if err == nil && !c.expectedOutput {
			t.Errorf("Login should have failed but did not case: `%s`: Got `%s` expected: `%s`", c.Case, "true", "false")
		}
	}
}
