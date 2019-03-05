package users

import (
	"testing"
)

//TODO: add tests for the various functions in user.go, as described in the assignment.
//use `go test -cover` to ensure that you are covering all or nearly all of your code paths.

func TestUserValidateAndToUser(t *testing.T) {
	cases := []struct {
		CaseName     string
		Email        string
		Password     string
		PasswordConf string
		UserName     string
		FirstName    string
		LastName     string
		ExpectedErr  bool
	}{
		{
			"Malformed Email Address 1",
			"testtest.com",
			"hunter2",
			"hunter2",
			"AzureDiamond",
			"John",
			"Johnson",
			true,
		}, {
			"Malformed Email Address 2",
			"test@testcom",
			"hunter2",
			"hunter2",
			"AzureDiamond",
			"John",
			"Johnson",
			true,
		}, {
			"Malformed Email Address 3",
			"test@",
			"hunter2",
			"hunter2",
			"AzureDiamond",
			"John",
			"Johnson",
			true,
		}, {
			"Too Short Password",
			"test@test.com",
			"pass",
			"pass",
			"AzureDiamond",
			"John",
			"Johnson",
			true,
		}, {
			"Password Mismatch",
			"test@test.com",
			"hunter2",
			"hunter1",
			"AzureDiamond",
			"John",
			"Johnson",
			true,
		}, {
			"Invalid UserName 1",
			"test@test.com",
			"hunter2",
			"hunter2",
			"",
			"John",
			"Johnson",
			true,
		}, {
			"Invalid UserName 2",
			"test@test.com",
			"hunter2",
			"hunter2",
			"Azure Diamond",
			"John",
			"Johnson",
			true,
		}, {
			"Valid New User",
			"test@test.com",
			"hunter2",
			"hunter2",
			"AzureDiamond",
			"John",
			"Johnson",
			false,
		},
	}

	for _, c := range cases {
		newUser := NewUser{c.Email, c.Password, c.PasswordConf, c.UserName, c.FirstName, c.LastName}
		err := newUser.Validate()
		if err != nil && !c.ExpectedErr {
			t.Errorf("Encountered unexpected error: %v", err)
		}
		_, err = newUser.ToUser()
		if err != nil && !c.ExpectedErr {
			t.Errorf("Encountered unexpected error: %v", err)
		}
	}
}

func TestUserFullName(t *testing.T) {
	cases := []struct {
		CaseName         string
		Email            string
		Password         string
		PasswordConf     string
		UserName         string
		FirstName        string
		LastName         string
		ExpectedFullName string
	}{
		{
			"Both Names Provided",
			"test@test.com",
			"hunter2",
			"hunter2",
			"AzureDiamond",
			"John",
			"Johnson",
			"John Johnson",
		}, {
			"First Name Provided",
			"test@test.com",
			"hunter2",
			"hunter2",
			"AzureDiamond",
			"John",
			"",
			"John",
		}, {
			"Last Name Provided",
			"test@test.com",
			"hunter2",
			"hunter2",
			"AzureDiamond",
			"",
			"Johnson",
			"Johnson",
		}, {
			"No Names Provided",
			"test@test.com",
			"hunter2",
			"hunter2",
			"AzureDiamond",
			"",
			"",
			"",
		},
	}

	for _, c := range cases {
		newUser := NewUser{c.Email, c.Password, c.PasswordConf, c.UserName, c.FirstName, c.LastName}
		user, err := newUser.ToUser()
		if err != nil {
			t.Errorf("Encountered unexpected error: %v", err)
		}
		if user.FullName() != c.ExpectedFullName {
			t.Errorf("Expected full name of: %s but instead received full name of: %s", c.ExpectedFullName, user.FullName())
		}
	}
}

func TestPasswords(t *testing.T) {
	cases := []struct {
		CaseName     string
		Email        string
		Password     string
		PasswordConf string
		UserName     string
		FirstName    string
		LastName     string
		AuthPass     string
		ExpectedErr  bool
	}{
		{
			"Incorrect Password Provided",
			"test@test.com",
			"hunter2",
			"hunter2",
			"AzureDiamond",
			"John",
			"Johnson",
			"hunter1",
			true,
		}, {
			"Correct Password Provided",
			"test@test.com",
			"hunter2",
			"hunter2",
			"AzureDiamond",
			"John",
			"Johnson",
			"hunter2",
			false,
		},
	}

	for _, c := range cases {
		newUser := NewUser{c.Email, c.Password, c.PasswordConf, c.UserName, c.FirstName, c.LastName}
		user, err := newUser.ToUser()
		if err != nil {
			t.Errorf("Encountered unexpected error: %v", err)
		}
		err = user.Authenticate(c.AuthPass)
		if err != nil && !c.ExpectedErr {
			t.Errorf("Encountered unexpected error: %v", err)
		}
	}
}

func TestUpdate(t *testing.T) {
	cases := []struct {
		CaseName     string
		Email        string
		Password     string
		PasswordConf string
		UserName     string
		FirstName    string
		LastName     string
		NewFirstName string
		NewLastName  string
		ExpectedErr  bool
	}{
		{
			"Valid Update 1",
			"test@test.com",
			"hunter2",
			"hunter2",
			"AzureDiamond",
			"John",
			"Johnson",
			"Johan",
			"Johanson",
			false,
		}, {
			"Valid Update 2",
			"test@test.com",
			"hunter2",
			"hunter2",
			"AzureDiamond",
			"John",
			"Johnson",
			"",
			"",
			false,
		},
	}

	for _, c := range cases {
		newUser := NewUser{c.Email, c.Password, c.PasswordConf, c.UserName, c.FirstName, c.LastName}
		user, err := newUser.ToUser()
		if err != nil {
			t.Errorf("Encountered unexpected error: %v", err)
		}
		testUpdate := Updates{c.NewFirstName, c.NewLastName}
		err = user.ApplyUpdates(&testUpdate)
		if err != nil && !c.ExpectedErr {
			t.Errorf("Encountered unexpected error: %v", err)
		}
		if c.NewFirstName != user.FirstName || c.NewLastName != user.LastName {
			t.Errorf("Update was not performed correctly. Given error: %v", err)
		}
	}
}
