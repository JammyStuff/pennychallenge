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

type AccountType int

const (
	AccTypeUnknown AccountType = iota
	AccTypePrepaid
	AccTypeRetail
)

var accountTypes = []string{
	"Unknown",
	"Prepaid",
	"Retail",
}

func (at AccountType) String() string {
	if at < AccTypeUnknown || at > AccTypeRetail {
		return "Unknown"
	}

	return accountTypes[at]
}

type Account struct {
	ID          string
	Description string
	TypeStr     string `json:"type"`
}

func (ac *Account) Type() AccountType {
	switch ac.TypeStr {
	case "uk_prepaid":
		return AccTypePrepaid
	case "uk_retail":
		return AccTypeRetail
	default:
		return AccTypeUnknown
	}
}

type AccountList struct {
	Accounts []Account
}
