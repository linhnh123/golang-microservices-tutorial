package dbclient

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/boltdb/bolt"
	"github.com/linhnh123/golang-microservices-tutorial/accountservice/model"
)

type IBoltClient interface {
	OpenBoltDb()
	QueryAccount(accountId string) (model.Account, error)
	Seed()
	Check() bool
	CloseBoltDb()
}

type BoltClient struct {
	boltDB *bolt.DB
}

func (bc *BoltClient) OpenBoltDb() {
	var err error
	bc.boltDB, err = bolt.Open("accounts.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Open DB")
}

func (bc *BoltClient) CloseBoltDb() {
	bc.boltDB.Close()
	log.Println("Close DB")
}

func (bc *BoltClient) Seed() {
	initializeBucket(bc)
	seedAccounts(bc)
}

func initializeBucket(bc *BoltClient) {
	bc.boltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("AccountBucket"))
		if err != nil {
			return fmt.Errorf("Create bucket failed: %s", err)
		}
		return nil
	})
}

func seedAccounts(bc *BoltClient) {
	total := 100
	for i := 0; i < total; i++ {
		key := strconv.Itoa(10000 + i)

		acc := model.Account{
			Id:   key,
			Name: "Person_" + strconv.Itoa(i),
		}

		jsonBytes, _ := json.Marshal(acc)

		bc.boltDB.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("AccountBucket"))
			err := b.Put([]byte(key), jsonBytes)
			return err
		})
	}
	log.Printf("Seeded %v fake accounts\n", total)
}

func (bc *BoltClient) QueryAccount(accountId string) (model.Account, error) {
	account := model.Account{}

	bc.OpenBoltDb()
	defer bc.CloseBoltDb()

	err := bc.boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("AccountBucket"))

		accountsBytes := b.Get([]byte(accountId))
		if accountsBytes == nil {
			return fmt.Errorf("No account found for " + accountId)
		}
		json.Unmarshal(accountsBytes, &account)

		return nil
	})

	if err != nil {
		logrus.Infoln(err.Error())
		return model.Account{}, nil
	}

	logrus.Infoln("Account found")
	return account, nil
}

func (bc *BoltClient) Check() bool {
	return bc.boltDB != nil
}
