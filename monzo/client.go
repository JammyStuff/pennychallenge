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

package monzo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const baseURL = "https://api.monzo.com"

type Client struct {
	httpClient  *http.Client
	accessToken string
}

func NewClient(accessToken string) *Client {
	return &Client{
		httpClient:  &http.Client{},
		accessToken: accessToken,
	}
}

func (c *Client) Accounts() (*[]Account, error) {
	req, err := http.NewRequest("GET", baseURL+"/accounts", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("/accounts returned %d status code", resp.StatusCode)
		return nil, errors.New(msg)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var accountList AccountList
	err = json.Unmarshal(body, &accountList)
	if err != nil {
		return nil, err
	}

	return &accountList.Accounts, nil
}

func (c *Client) Balance(account *Account) (*Balance, error) {
	url := fmt.Sprintf("%s/balance?account_id=%s", baseURL, account.ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("/balance returned %d status code", resp.StatusCode)
		return nil, errors.New(msg)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var balance Balance
	err = json.Unmarshal(body, &balance)
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

func (c *Client) DepositToPot(pot *Pot, source *Account, amount int64, id string) error {
	values := url.Values{}
	values.Add("source_account_id", source.ID)
	values.Add("amount", strconv.FormatInt(amount, 10))
	values.Add("dedupe_id", id)
	body := strings.NewReader(values.Encode())

	url := fmt.Sprintf("%s/pots/%s/deposit", baseURL, pot.ID)
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("/pots/deposit returned %d status code", resp.StatusCode)
		return errors.New(msg)
	}

	return nil
}

func (c *Client) Pots() (*[]Pot, error) {
	req, err := http.NewRequest("GET", baseURL+"/pots", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("/pots returned %d status code", resp.StatusCode)
		return nil, errors.New(msg)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var potList PotList
	err = json.Unmarshal(body, &potList)
	if err != nil {
		return nil, err
	}

	return &potList.Pots, nil
}
