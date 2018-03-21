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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jammystuff/pennychallenge/monzo"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DateFormat = "2006-01-02"
	DaysInYear = 365
	MinBalance = 1000
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pennychallenge",
	Short: "Automated (reversed) penny challenge",
	Long: `The penny challenge involves saving 1p on the first day of the year,
2p on the second day of the year, 3p on the third day of the year, and so on.

pennychallenge implements the reversed penny challenge automatically using the
Monzo API to save the correct amount into a pot. This can then be run on a daily
schedule using a functions as a service provider (e.g. Azure Functions).

The advantage of the reversed penny challenge is that it avoids the maximum
savings occurring during December, which is when spending is highest.`,
	Run: runRoot,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pennychallenge.yaml)")

	rootCmd.Flags().StringP("source-account", "s", "", "Account ID to save from")
	viper.BindPFlag("source_account", rootCmd.Flags().Lookup("source-account"))

	rootCmd.Flags().StringP("destination-pot", "d", "", "Pot ID to save to")
	viper.BindPFlag("destination_pot", rootCmd.Flags().Lookup("destination-pot"))

	rootCmd.PersistentFlags().StringP("client-id", "I", "", "Monzo API client ID")
	viper.BindPFlag("client_id", rootCmd.PersistentFlags().Lookup("client-id"))

	rootCmd.PersistentFlags().StringP("client-secret", "S", "", "Monzo API client secret")
	viper.BindPFlag("client_secret", rootCmd.PersistentFlags().Lookup("client-secret"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".pennychallenge" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".pennychallenge")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func amountToSave(date time.Time) int64 {
	days := int64(daysInYear(date))
	yearDay := int64(date.YearDay())
	return days + 1 - yearDay
}

func checkBalance(account *monzo.Account, c *monzo.Client) (bool, error) {
	balance, err := c.Balance(account)
	if err != nil {
		return false, err
	}

	if balance.Balance < MinBalance {
		return false, nil
	}

	return true, nil
}

func daysInYear(date time.Time) int {
	year := date.Year()
	if year%4 != 0 {
		return DaysInYear
	} else if year%100 != 0 {
		return DaysInYear + 1
	} else if year%400 != 0 {
		return DaysInYear
	}
	return DaysInYear + 1
}

func getAccount(id string, c *monzo.Client) (*monzo.Account, error) {
	accounts, err := c.Accounts()
	if err != nil {
		return nil, err
	}

	for _, account := range *accounts {
		if account.ID == id {
			return &account, nil
		}
	}

	msg := fmt.Sprintf("Account %s not found", id)
	return nil, errors.New(msg)
}

func savePennies(account *monzo.Account, pot *monzo.Pot, c *monzo.Client) error {
	date := time.Now().UTC()
	amount := amountToSave(date)
	id := fmt.Sprintf("PENNY-%s", date.Format(DateFormat))

	return c.DepositToPot(pot, account, amount, id)
}

func getPot(id string, c *monzo.Client) (*monzo.Pot, error) {
	pots, err := c.Pots()
	if err != nil {
		return nil, err
	}

	for _, pot := range *pots {
		if pot.ID == id {
			return &pot, nil
		}
	}

	msg := fmt.Sprintf("Pot %s not found", id)
	return nil, errors.New(msg)
}

func runRoot(cmd *cobra.Command, args []string) {
	t, err := readToken()
	if err != nil {
		fmt.Println("Error reading access token")
		os.Exit(1)
	}

	clientID := viper.GetString("client_id")
	clientSecret := viper.GetString("client_secret")
	refresh, err := refreshToken(clientID, clientSecret, t)
	if err != nil {
		fmt.Printf("Error refreshing access token: %v\n", err)
		os.Exit(1)
	}

	token := refresh.AccessToken
	client := monzo.NewClient(token)

	fmt.Print("Getting account... ")
	accountID := viper.GetString("source_account")
	account, err := getAccount(accountID, client)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Printf("Error getting account: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK")

	fmt.Print("Checking balance... ")
	ok, err := checkBalance(account, client)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Printf("Error checking balance: %v\n", err)
		os.Exit(1)
	}
	if !ok {
		fmt.Println("FAIL")
		fmt.Println("Account balance too low")
		os.Exit(2)
	}
	fmt.Println("OK")

	fmt.Print("Getting pot... ")
	potID := viper.GetString("destination_pot")
	pot, err := getPot(potID, client)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Printf("Error getting pot: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK")

	fmt.Print("Saving... ")
	err = savePennies(account, pot, client)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Printf("Error saving: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK")
}
