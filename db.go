package main

import (
	"fmt"
	// "github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const EmailAddressKey = "email_address"

func getEmailAddressKey(blockchainAddress string) string {
	return fmt.Sprintf("%s/%s", EmailAddressKey, blockchainAddress)
}

func (notificator Notificator) GetEmailAddress(blockchainAddress string) string {
	key := []byte(getEmailAddressKey(blockchainAddress))
	bz, err := notificator.DB.Get(key, &opt.ReadOptions{})
	if err != nil {
		return ""
	}

	return string(bz)
}

func (notificator Notificator) SetEmailAddress(blockchainAddress string, emailAddress string) error {
	key := []byte(getEmailAddressKey(blockchainAddress))
	bz := []byte(emailAddress)

	err := notificator.DB.Put(key, bz, &opt.WriteOptions{})
	if err != nil {
		return err
	}

	return nil
}
