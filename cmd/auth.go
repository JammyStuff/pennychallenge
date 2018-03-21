// Copyright © 2018 James Wheatley <james@jammy.co>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const baseAuthURL = "https://auth.monzo.com"
const redirectURL = "http://localhost"
const tokenPath = "pennychallengetoken.json"
const tokenURL = "https://api.monzo.com/oauth2/token"

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate to Monzo",
	Long:  `Authenticates to the Monzo API.`,
	Run:   runAuth,
}

func init() {
	rootCmd.AddCommand(authCmd)
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func getAuthURL(clientID string) string {
	values := url.Values{}
	values.Add("client_id", clientID)
	values.Add("redirect_uri", redirectURL)
	values.Add("response_type", "code")
	query := values.Encode()

	return fmt.Sprintf("%s/?%s", baseAuthURL, query)
}

func getToken(clientID, clientSecret, authCode string) (*Token, error) {
	values := url.Values{}
	values.Add("grant_type", "authorization_code")
	values.Add("client_id", clientID)
	values.Add("client_secret", clientSecret)
	values.Add("redirect_uri", redirectURL)
	values.Add("code", authCode)
	body := strings.NewReader(values.Encode())

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("%d status code when getting token", resp.StatusCode)
		return nil, errors.New(msg)
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var token Token
	err = json.Unmarshal(respBody, &token)
	return &token, err
}

func readToken() (*Token, error) {
	data, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}

	var token Token
	err = json.Unmarshal(data, &token)
	return &token, err
}

func refreshToken(clientID, clientSecret string, token *Token) (*Token, error) {
	values := url.Values{}
	values.Add("grant_type", "refresh_token")
	values.Add("client_id", clientID)
	values.Add("client_secret", clientSecret)
	values.Add("refresh_token", token.RefreshToken)
	body := strings.NewReader(values.Encode())

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("%d status code when refreshing token", resp.StatusCode)
		return nil, errors.New(msg)
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var newToken Token
	err = json.Unmarshal(respBody, &newToken)
	if err != nil {
		return nil, err
	}

	err = writeToken(&newToken)

	return &newToken, err
}

func writeToken(token *Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(tokenPath, data, 0600)
}

func runAuth(cmd *cobra.Command, args []string) {
	clientID := viper.GetString("client_id")
	clientSecret := viper.GetString("client_secret")

	authURL := getAuthURL(clientID)
	fmt.Printf("Go to %s\n", authURL)

	var authCode string
	fmt.Print("Authorization code: ")
	fmt.Scanln(&authCode)

	fmt.Print("Getting access token... ")
	token, err := getToken(clientID, clientSecret, authCode)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Printf("Error getting access token: %v", err)
		os.Exit(1)
	}
	fmt.Println("OK")

	fmt.Print("Writing access token... ")
	err = writeToken(token)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Printf("Error writing access token: %v", err)
		os.Exit(1)
	}
	fmt.Println("OK")
}
