package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	errconst "ticketing-service/constants"
	errconstant "ticketing-service/constants"
	errors "ticketing-service/error"
	"ticketing-service/logger"
	"ticketing-service/model"
	"ticketing-service/util"
	"time"
)

var (
	instance Data_provider
	once     sync.Once
	dbUser   = util.GetEnv(util.DB_USER, util.GetProperty(util.DB_USER))
	dbName   = util.GetEnv(util.DB_NAME, util.GetProperty(util.DB_NAME))
	// dbPassword        = os.Getenv(util.TICKETING_SERVICE_DB_PASSWORD)
	dbPassword        = ""
	dbPwdPath         = os.Getenv(util.POSTGRES_SECRET_PATH)
	tableName         = util.GetProperty(util.TICKETING_SYSTEM_DB_TABLE)
	hoststr           = util.GetEnv(util.DB_HOST, "localhost")
	portstr           = util.GetEnv(util.DB_PORT, "5432")
	connectionDetails = ""
	driver            = util.POSTGRES_DRIVER
	isPostgresEnabled = true
)

const (
	QUERY_UPDATE    = "UPDATE "
	QUERY_IS_LATEST = " and islatest = 'true';"
	VERSION         = "version"
	AND             = " and "
)

type dbProvider struct {
	DbConnection *sql.DB
}

type Data_provider interface {
	InsertReservationData(tickets []model.Ticket) (int, error)
	UpdateReservationData(tickets []model.UpdateTickets) (int, []int, error)
	CancelReservationData(tickets model.ReserveTicketResult) (int, []int, error)
	CloseDB()
	DatabasePing(timeout time.Duration) error
}

func init() {
	isPostgresEnabled, _ = strconv.ParseBool(util.GetEnv(util.IS_POSTGRES_ENABLED, util.GetProperty(util.IS_POSTGRES_ENABLED)))
	if isPostgresEnabled {
		data, err := ioutil.ReadFile(dbPwdPath)
		if err != nil {
			logger.Log().Error("Error caught while reading the postgres secret file :", err)
		}
		dbPassword = string(data)
		logger.Log().Info("Database Configured is : Postgres")
	} else {
		logger.Log().Info("Database Configured is : SQLite")
	}
}

//GetDataProviderInstance is used to create instance for dbProvider
func GetDataProviderInstance() Data_provider {
	if err := getDriverAndConnectionDetails(); err != nil {
		logger.Log().Errorf("Invalid db connection details:", err)
		instance = &dbProvider{}
		return instance
	}
	once.Do(func() {
		pdb, dbError := DBConnect(connectionDetails, driver)
		if dbError == nil {
			instance = &dbProvider{pdb}
			createTable(pdb)
		} else {
			logger.Log().Error("No db connection available")
			instance = &dbProvider{}
		}
	})
	return instance
}

//This function is to create table in Sqlite DB
func createTable(db *sql.DB) {
	if !isPostgresEnabled {
		query := util.GetProperty(util.CREATE_TABLE_QUERY)
		if _, err := db.Exec(query); err != nil {
			logger.Log().Error("Failed to create table in Sqlite DB")
		}
	}
}

func getDriverAndConnectionDetails() error {

	driver = util.POSTGRES_DRIVER
	//dbUser = "postgres"
	//dbPassword = "admin"
	//hoststr = "localhost"

	connectionDetails = "postgres://" + dbUser + ":" + dbPassword + "@" + hoststr + ":" + portstr + "/" + dbName + "?sslmode=disable&connect_timeout=" + util.GetEnv(util.DB_CONNECTION_TIMEOUT_SECOND, util.GetProperty(util.DB_CONNECTION_TIMEOUT_SECOND))
	logger.Log().Info("Connection details:" + connectionDetails)
	return nil

}

func (db *dbProvider) CloseDB() {
	if db.DbConnection == nil {
		logger.Log().Warn("No connection to db available. Nothing to close.")
		return
	}
	logger.Log().Info("Closing db connection")
	err := db.DbConnection.Close()
	if err != nil {
		logger.Log().Errorf("Error occured in closing db connection : %v", err)
	}
	logger.Log().Info("Successfully closed db connection")
}

func (db *dbProvider) InsertReservationData(tickets []model.Ticket) (int, error) {
	stackMessage := "Error occured in InsertReservationData"
	query := `insert into reservation (id,name,place_to,place_from,date,time) values `

	values := []interface{}{}
	for i, ticket := range tickets {
		values = append(values, ticket.Id, ticket.Name, ticket.To, ticket.From, ticket.Date, ticket.Time)

		numFields := 6 // the number of fields you are inserting
		n := i * numFields

		query += `(`
		for j := 0; j < numFields; j++ {
			query += `$` + strconv.Itoa(n+j+1) + `,`
		}
		query = query[:len(query)-1] + `),`
	}
	query = query[:len(query)-1] // remove the trailing comma
	query = query + " RETURNING ticket_id"
	logger.Log().Infof(query)
	fmt.Println(values)

	if err := checkBDConnection(db); err != nil {
		logger.Log().Error(err)
		return -1, err
	}
	ticket_id := 0
	err := db.DbConnection.QueryRow(query, values...).Scan(&ticket_id)
	fmt.Println("New record ID is:", ticket_id)

	if err != nil {
		logger.Log().Error(err)
		return -1, errors.AddMessageToStack(err, stackMessage)
	}
	return ticket_id, nil
}

func (db *dbProvider) UpdateReservationData(tickets []model.UpdateTickets) (int, []int, error) {
	//stackMessage := "Error occured in InsertReservationData"
	count := 0
	failedList := make([]int, 0)
	for _, ticket := range tickets {
		updStmt := []string{QUERY_UPDATE + tableName + " SET place_to = '" + ticket.To + "',place_from = '" + ticket.From + "',date = '" + ticket.Date + "',time = '" + ticket.Time + "' WHERE ticket_id =" + strconv.Itoa(ticket.TicketId) + ";"}
		query := strings.Join(updStmt, "")
		logger.Log().Infof("Query for UpdateReservationData : %s", query)
		if err := checkBDConnection(db); err != nil {
			return -1, failedList, err
		}
		count, countErr := DBExec(db.DbConnection, query)
		if countErr != nil {
			logger.Log().Errorf("Error caught in data_provider UpdateReservationData function: %v", countErr)
			//return -1, errors.AddMessageToStack(countErr, stackMessage)
		}
		if count <= 0 {
			failedList = append(failedList, ticket.TicketId)
			logger.Log().Warnf(errors.GetMessage(errconst.FAILED_TO_UPDATE_DATA)+" : given ticket Id: %s", ticket.TicketId)
			//return -1, errors.NewCustomErr(errconst.FAILED_TO_UPDATE_DATA, stackMessage)
		}
	}
	logger.Log().Info("Data updated successfully")
	return int(count), failedList, nil
}

func (db *dbProvider) CancelReservationData(tickets model.ReserveTicketResult) (int, []int, error) {
	//stackMessage := "Error occured in InsertReservationData"
	count := 0
	failedList := make([]int, 0)
	for _, ticket := range tickets.TicketNumbers {
		updStmt := []string{"Delete from " + tableName + " WHERE ticket_id =" + strconv.Itoa(ticket) + ";"}
		query := strings.Join(updStmt, "")
		logger.Log().Infof("Query for CancelReservationData : %s", query)
		if err := checkBDConnection(db); err != nil {
			return -1, failedList, err
		}
		count, countErr := DBExec(db.DbConnection, query)
		if countErr != nil {
			logger.Log().Errorf("Error caught in data_provider DeleteReservationData function: %v", countErr)
		}
		if count <= 0 {
			failedList = append(failedList, ticket)
			logger.Log().Warnf(errors.GetMessage(errconst.FAILED_TO_UPDATE_DATA)+" : given ticket Id: %s", ticket)
		}
	}
	logger.Log().Info("Data Deleted successfully")
	return int(count), failedList, nil
}

func (db *dbProvider) DatabasePing(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if db.DbConnection == nil {
		return fmt.Errorf("Database connection is nil")
	}
	return db.DbConnection.PingContext(ctx)
}

func handleError(err error) error {
	stackMessage := "[DATA_PROVIDER] In data_provider (InsertInitSchema) method"
	if err != nil {
		logger.Log().Errorf("Failed to insert the schema : %v", err)
		err := errors.WrapError(GetDBErrorCode(err), err, stackMessage)
		return err
	}
	return nil
}

func checkBDConnection(db *dbProvider) error {
	stackMsg := " [data_provider] Failure in db_util (CheckBDConnection)"
	if db.DbConnection == nil {
		pdb, dbError := DBConnect(connectionDetails, driver)
		if dbError != nil {
			err := errors.NewCustomErr(errconstant.DB_CONNECTION_ERR, stackMsg)
			logger.Log().Warnf("Error caught in data_provider in checkBDConnection function, connection not established to DB ")
			return err
		}
		db.DbConnection = pdb
	}
	return nil
}
