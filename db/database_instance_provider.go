package db

import (
	"database/sql"
	"fmt"
	errconstant "ticketing-service/constants"
	errors "ticketing-service/error"
	"time"

	_ "github.com/lib/pq"
)

var SQLOpen = sql.Open

const DB_CONNECTION_ERR_MSG = "Error caught in db_util in DBConnect function, unable to ping DB :: "

//DBConnect func is used to establish DataBase connection
func DBConnect(connectionDetails string, driver string) (pdb *sql.DB, err error) {
	fmt.Println("Creating DB instance")
	stackMsg := " [DB_INSTANCE_PROVIDER] Failure in instance provider (DBConnect) "
	//fmt.Printlnf("db connection detail :: %s", connectionDetails)
	startTime := time.Now()
	pdb, err = SQLOpen(driver, connectionDetails)
	if err == nil {
		err = pdb.Ping()
		if err != nil {
			err = errors.WrapError(errconstant.DB_CONNECTION_ERR, err, stackMsg)
			fmt.Printf(DB_CONNECTION_ERR_MSG+"%+s\n", err)
		}
	} else {
		err = errors.WrapError(errconstant.DB_CONNECTION_ERR, err, stackMsg)
		fmt.Printf(DB_CONNECTION_ERR_MSG+"%+s\n", err)
		return nil, err
	}
	fmt.Println("[TIME TAKEN] to try to establish connection to DB :: %s", time.Since(startTime))
	if err != nil {
		fmt.Printf(DB_CONNECTION_ERR_MSG+"%+s\n", err)
		return nil, err
	}
	pdb.SetMaxOpenConns(3)
	fmt.Println("Successfully connected to DB!")
	return pdb, err
}
