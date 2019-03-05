package sessions

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

//InvalidSessionID represents an empty, invalid session ID
const InvalidSessionID SessionID = ""

//idLength is the length of the ID portion
const idLength = 32

//signedLength is the full length of the signed session ID
//(ID portion plus signature)
const signedLength = idLength + sha256.Size

//SessionID represents a valid, digitally-signed session ID.
//This is a base64 URL encoded string created from a byte slice
//where the first `idLength` bytes are crytographically random
//bytes representing the unique session ID, and the remaining bytes
//are an HMAC hash of those ID bytes (i.e., a digital signature).
//The byte slice layout is like so:
//+-----------------------------------------------------+
//|...32 crypto random bytes...|HMAC hash of those bytes|
//+-----------------------------------------------------+
type SessionID string

//ErrInvalidID is returned when an invalid session id is passed to ValidateID()
var ErrInvalidID = errors.New("Invalid Session ID")

//NewSessionID creates and returns a new digitally-signed session ID,
//using `signingKey` as the HMAC signing key. An error is returned only
//if there was an error generating random bytes for the session ID
func NewSessionID(signingKey string) (SessionID, error) {
	//TODO: if `signingKey` is zero-length, return InvalidSessionID
	//and an error indicating that it may not be empty
	if len(signingKey) == 0 {
		return InvalidSessionID, errors.New("Given signing key is of length 0")
	}

	//TODO: Generate a new digitally-signed SessionID by doing the following:
	//- create a byte slice where the first `idLength` of bytes
	//  are cryptographically random bytes for the new session ID,
	//  and the remaining bytes are an HMAC hash of those ID bytes,
	//  using the provided `signingKey` as the HMAC key.
	//- encode that byte slice using base64 URL Encoding and return
	//  the result as a SessionID type

	randBytes := make([]byte, idLength)
	rand.Read(randBytes)

	hmacer := hmac.New(sha256.New, []byte(signingKey))
	hashed := make([]byte, sha256.Size)
	_, err := io.Copy(hmacer, bytes.NewReader(randBytes))
	if err != nil {
		return InvalidSessionID, err
	}
	hashed = hmacer.Sum(nil)

	combined := append(randBytes, hashed...)

	complete := base64.URLEncoding.EncodeToString(combined)

	finalSessionID := SessionID(complete)

	return finalSessionID, nil

	//the following return statement is just a placeholder
	//remove it when implementing the function
	//return InvalidSessionID, nil
}

//ValidateID validates the string in the `id` parameter
//using the `signingKey` as the HMAC signing key
//and returns an error if invalid, or a SessionID if valid
func ValidateID(id string, signingKey string) (SessionID, error) {

	//TODO: validate the `id` parameter using the provided `signingKey`.
	//base64 decode the `id` parameter, HMAC hash the
	//ID portion of the byte slice, and compare that to the
	//HMAC hash stored in the remaining bytes. If they match,
	//return the entire `id` parameter as a SessionID type.
	//If not, return InvalidSessionID and ErrInvalidID.
	decoded, err := base64.URLEncoding.DecodeString(id)
	if err != nil {
		return InvalidSessionID, err
	}

	idPortion := decoded[:idLength]
	sigPortion := decoded[idLength:]

	tester := hmac.New(sha256.New, []byte(signingKey))

	_, err = io.Copy(tester, bytes.NewReader(idPortion))
	if err != nil {
		return InvalidSessionID, err
	}

	generatedSig := tester.Sum(nil)

	if hmac.Equal(generatedSig, sigPortion) {
		return SessionID(id), nil
	}

	return InvalidSessionID, ErrInvalidID
}

//String returns a string representation of the sessionID
func (sid SessionID) String() string {
	return string(sid)
}
