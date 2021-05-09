package db

import (
	"database/sql"
	"net"
	"strings"
	errconstant "ticketing-service/constants"
	errors "ticketing-service/error"
	"ticketing-service/logger"
	"ticketing-service/util"

	"github.com/lib/pq"
)

//DBClose func is used to close DataBase connection
func DBClose(pdb *sql.DB) error {
	stackMsg := " [DB_UTIL] Failure in db_util (DBClose) "
	err := pdb.Close()
	if err != nil {
		err = errors.WrapError(errconstant.DB_CONNECTION_ERR, err, stackMsg)
		logger.Log().Errorf("Error caught in db_util in DBClose function : %s", err)
		return err
	}
	logger.Log().Debug("Successfully closed DB!")
	return nil
}

//DBExec func is used to execute the qurey
func DBExec(pdb *sql.DB, query string) (int64, error) {
	stackMsg := " [DB_UTIL] Failure in db_util (DBExec)"
	res, exeErr := pdb.Exec(query)
	if exeErr != nil {
		logger.Log().Errorf("Error caught in db_util DBExec function while executing query: %v", exeErr)
		err := errors.WrapError(GetDBErrorCode(exeErr), exeErr, stackMsg)
		return -1, err
	}
	count, countErr := res.RowsAffected()
	if countErr != nil {
		logger.Log().Errorf("Error caught in db_util DBExec function: %v", countErr)
		err := errors.WrapError(errconstant.FAILED_TO_ITERATE_RESULT, countErr, stackMsg)
		return -1, err
	}
	return count, nil
}

//DBGetMultipleCols func is used to executes a query that returns rows
func DBGetMultipleCols(pdb *sql.DB, query string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	rows, exeErr := pdb.Query(query)
	stackMsg := " [DB_UTIL] Failure in db_util (DBGet) "
	if exeErr != nil {
		err := errors.WrapError(errconstant.FAILED_TO_EXECUTE_QUERY, exeErr, stackMsg)
		logger.Log().Errorf("Error caught in db_util DBGetMultipleCols function: %+v\n", exeErr)
		return result, err
	}
	if rows != nil {
		cols, _ := rows.Columns()
		for rows.Next() {
			// Create a slice of interface{}'s to represent each column,
			// and a second slice to contain pointers to each item in the columns slice.
			columns := make([]interface{}, len(cols))
			columnPointers := make([]interface{}, len(cols))
			for i := range columns {
				//sql package requires pointers when scanning
				columnPointers[i] = &columns[i]
			}
			// Scan the result into the column pointers...
			if err := rows.Scan(columnPointers...); err != nil {
				err := errors.WrapError(errconstant.FAILED_TO_ITERATE_RESULT, err, stackMsg)
				logger.Log().Errorf("Error caught while iterating result set in db_util DBGetMultipleCols function: %v", exeErr)
				return result, err
			}
			// Create map, and retrieve the value for each column from the pointers slice,
			// storing it in the map with the name of the column as the key.
			for i, colName := range cols {
				val := columnPointers[i].(*interface{})
				result[colName] = *val
			}
			defer rows.Close()
		}
	}

	return result, nil
}

//DBGet func is used to executes a query that returns rows
func DBGet(pdb *sql.DB, query string) ([]byte, error) {
	var resultSet []byte
	var count int64
	var datas_ []string
	var resultStr string
	rows, exeErr := pdb.Query(query)
	stackMsg := " [DB_UTIL] Failure in db_util (DBGet) "
	if exeErr != nil {
		err := errors.WrapError(errconstant.FAILED_TO_EXECUTE_QUERY, exeErr, stackMsg)
		logger.Log().Errorf("Error caught in db_util DBGet function: %+v\n", exeErr)
		return resultSet, err
	}

	if rows != nil {
		for rows.Next() {
			var data_ string
			err := rows.Scan(&data_)
			count = count + 1
			defer rows.Close()
			if err != nil {
				err := errors.WrapError(errconstant.FAILED_TO_ITERATE_RESULT, err, stackMsg)
				logger.Log().Errorf("Error caught while iterating result set in db_util DBGet function: %v", exeErr)
				return resultSet, err
			}
			datas_ = append(datas_, data_)
		}
	}
	if datas_ != nil {
		if count == 1 {
			resultStr = strings.Join(datas_, ",")
		} else {
			resultStr = "[" + strings.Join(datas_, ",") + "]"
		}
		resultSet = append(resultSet, resultStr...)
	}
	return resultSet, nil
}

func GetDBErrorCode(exeErr error) string {
	if postgresErr, ok := exeErr.(*pq.Error); ok && postgresErr.Code == util.PG_PRIMARYKEY_CONSTRAINT_ERR_CODE {
		return errconstant.TICKET_ID_ALREADY_PRESENT
	}
	if _, ok := exeErr.(*net.OpError); ok {
		return errconstant.DB_CONNECTION_ERR
	}
	return errconstant.FAILED_TO_EXECUTE_QUERY
}
