package gdrive

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"os"
)

// OAuth is for the google drive access.
// If successful, the Google Drive service will be returned.
// Use this function to create new oauth credentials!
func OAuth(clientCredFile, tokenFile string, readonly bool) (*drive.Service, error) {
	// scope (default: read & write access)
	scope := drive.DriveScope
	if readonly {
		scope = drive.DriveReadonlyScope
	}

	// ConfigFromJSON uses a Google Developers Console client_credentials.json file to construct a config.
	// client_credentials.json can be downloaded from https://console.developers.google.com, under "Credentials".
	// Download the Web application credentials in the JSON format and provide the contents of the file as jsonKey.
	oAuthConf, err := loadOAuthConf(clientCredFile, scope)
	if err != nil {
		log.Printf("ERROR: %s/OAuth: %v", packageName, err)
		log.Printf("ERROR: %s/OAuth: This link could help: https://www.google.com/search?q=drive+client+credential", packageName)
		return nil, err
	}

	// Load oauth 2.0 token from a file. The token represents the credentials used to authorize
	// the requests to access protected resources on the OAuth 2.0 provider's backend.
	tok, err := loadToken(tokenFile)
	if err != nil {
		log.Printf("WARNING: %s/OAuth: %v", packageName, err)

		// get token with user interaction
		tok, err = reqNewToken(tokenFile, oAuthConf)
		if err != nil {
			return nil, err
		}
	}

	// new drive service with OAuth2
	ctx := context.Background()
	service, err := drive.NewService(ctx, option.WithTokenSource(oAuthConf.TokenSource(ctx, tok)))
	if err != nil {
		return nil, fmt.Errorf("gdrive/OAuth: %v", err)
	}

	return service, nil
}

//--------  HELPER  --------------------------------------------------------------------------------------------------//

// loadOAuthConf loads a valid OAuth config from a file
func loadOAuthConf(file, scope string) (*oauth2.Config, error) {
	// read the entire file
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("gdrive/loadOAuthConf: %v", err)
	}

	// parse oAuthConf
	oAuthConf, err := google.ConfigFromJSON(b, scope)
	if err != nil {
		return nil, fmt.Errorf("gdrive/loadOAuthConf: %v", err)
	}

	// success
	return oAuthConf, nil
}

// loadToken loads a valid token from a file
func loadToken(file string) (*oauth2.Token, error) {
	// open token file
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("gdrive/loadToken: %v", err)
	}
	defer f.Close()

	// parse token
	tok := new(oauth2.Token)
	err = json.NewDecoder(f).Decode(tok) // decode file
	if err != nil {
		return nil, fmt.Errorf("gdrive/loadToken: %v", err)
	}

	// success
	return tok, nil
}

// reqNewToken allow the user to request a token (user interaction).
// If successful, the valid token is written to a file and returned.
func reqNewToken(file string, oAuthConf *oauth2.Config) (*oauth2.Token, error) {
	// get authorization code from web (with user interaction)
	var authCode string
	authURL := oAuthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("\nFollow the link and create a new token file: %v\n\nEnter the authorization code here: ", authURL)
	_, _ = fmt.Scan(&authCode) // read user input

	// convert authorization code to token
	tok, err := oAuthConf.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("gdrive/reqNewToken: %v", err)
	}

	// open token file
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600) // override file
	if err != nil {
		return nil, fmt.Errorf("gdrive/reqNewToken: %v", err)
	}
	defer f.Close()

	// write token to file
	err = json.NewEncoder(f).Encode(tok)
	if err != nil {
		return nil, fmt.Errorf("gdrive/reqNewToken: %v", err)
	}

	// success
	return tok, nil
}
