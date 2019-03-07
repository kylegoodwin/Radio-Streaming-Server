package users

import (
	"context"
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"

	"github.com/Radio-Streaming-Server/servers/gateway/indexes"
	"go.mongodb.org/mongo-driver/mongo"
)

//MongoStore is a struct that holds a *mongo.Collection as Collection
type MongoStore struct {
	Collection *mongo.Collection
	Trie       *indexes.Trie
}

//NewMongoStore generates a MongoStore struct and returns it
func NewMongoStore(col *mongo.Collection) *MongoStore {
	//Thank Mr.TA
	if col != nil {
		return &MongoStore{
			Collection: col,
			Trie:       indexes.NewTrie(),
		}
	}

	return nil
}

//PopulateTrie pulls users from the store and populates them into the Trie
func (ms *MongoStore) PopulateTrie() error {
	//Grab these fields from all users
	//nil as the filter will pull all documents
	cur, err := ms.Collection.Find(context.TODO(), nil, options.Find())
	defer cur.Close(context.TODO())
	if err != nil {
		return err
	}

	//For each user in the DB, insert the correct key pairs for that user
	for cur.Next(context.TODO()) {
		//Grab the user's info
		tempUser := &User{}
		if err := cur.Decode(&tempUser); err != nil {
			return errors.New("Failed while parsing rows")
		}

		//Add the user into the Trie
		ms.InsertUserIntoTrie(tempUser)
	}

	return nil
}

//InsertUserIntoTrie adds a user to the mongo trie using the ID, UserName, FirstName, and LastName fields
func (ms *MongoStore) InsertUserIntoTrie(user *User) {
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
		ms.Trie.Add(i, user.ID)
	}
}

//GetByID returns the User with the given ID
func (ms *MongoStore) GetByID(id string) (*User, error) {
	user := User{}
	err := ms.Collection.FindOne(context.TODO(), bson.D{{"_id", id}}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//GetByEmail returns the User with the given email
func (ms *MongoStore) GetByEmail(email string) (*User, error) {
	user := User{}
	//Uncertain if this will work or if the email needs capitalization in the document key
	err := ms.Collection.FindOne(context.TODO(), bson.D{{"email", email}}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//GetByUserName returns the User with the given Username
func (ms *MongoStore) GetByUserName(username string) (*User, error) {
	user := User{}
	//Uncertain if this will work or if the username needs capitalization in the document key
	err := ms.Collection.FindOne(context.TODO(), bson.D{{"userName", username}}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//Insert inserts the user into the database, and returns
//the newly-inserted User, complete with the DBMS-assigned ID
func (ms *MongoStore) Insert(user *User) (*User, error) {
	insq := "insert into users(email, pass_hash, user_name, first_name, last_name, photo_URL) values (?,?,?,?,?,?)"
	res, err := ms.Collection.Exec(insq, user.Email, user.PassHash, user.UserName, user.FirstName, user.LastName, user.PhotoURL)
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
func (ms *MongoStore) InsertSignIn(userID int64, ip string) (int64, error) {
	insq := "insert into successful_logins(user_id, sign_in_time, login_ip) values (?,now(),?)"
	res, err := ms.Collection.Exec(insq, userID, ip)
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
func (ms *MongoStore) Update(id int64, updates *Updates) (*User, error) {

	mongocmd := "update users set first_name=? last_name=? where id=?"

	_, err := ms.Collection.Exec(mongocmd, updates.FirstName, updates.LastName, id)
	if err != nil {
		return nil, err
	}

	//Just to properly return an updated user. Probably not optimal
	user, err := ms.GetByID(id)
	user.ApplyUpdates(updates)

	return user, nil
}

//Delete deletes the user with the given ID
func (ms *MongoStore) Delete(id int64) error {
	mongocmd := "delete from users where id=?"
	_, err := ms.Collection.Exec(mongocmd, id)
	if err != nil {
		return err
	}
	return nil
}
