package main

/*
/ This feature's purpose is to support "TURN Server REST API", see
/ "TURN REST API" link in the project's page
/ https://github.com/coturn/coturn/
/
/ This option is used with timestamp:
/
/ usercombo -> "timestamp:userid"
/ turn user -> usercombo
/ turn password -> base64(hmac(secret key, usercombo))
*/

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"time"
)

type Credential struct {
	Username string
	Password string
}

// newCredential creates a new Credential object
func NewCredential(secretkey string) *Credential {
	if len(secretkey) == 0 {
		// Default for test case
		secretkey = "north"
	}
	usercombo := generateUser()

	return &Credential{
		Username: usercombo,
		Password: generatePassword(secretkey, usercombo),
	}
}

// usercombo -> "timestamp:userid"
// turn user -> usercombo
// generaUser according to what is expected by coturn
// the user is available only for 60 second
func generateUser() string {
	now := time.Now()
	timestamp := now.Unix()
	timestamp += 60
	return fmt.Sprintf("%d:%s", timestamp, "lk")
}

// generate a password
func generatePassword(secretkey string, usercombo string) string {
	mac := hmac.New(sha1.New, []byte(secretkey))
	mac.Write([]byte(usercombo))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
