package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	model "ticketing-service/model"

	"ticketing-service/logger"
	service "ticketing-service/service"

	errors "ticketing-service/error"
	"ticketing-service/util"

	errconst "ticketing-service/constants"

	"gopkg.in/go-playground/validator.v9"
)

var ticketManager = service.GetTicketManager()

func BookTicket(w http.ResponseWriter, r *http.Request) {

	payload, readErr := ioutil.ReadAll(r.Body)
	stackMessage := "Error in BookTicket"

	if readErr != nil {
		logger.Log().Infof("Error while unmarshaling data: %s\n", readErr)
		err := errors.WrapError(errconst.INVALID_JSON_DATA, readErr, stackMessage)
		writeErrorResponse(err, w)
		return
	}
	var ticketInfo []model.Ticket
	err := json.Unmarshal(payload, &ticketInfo)
	v := validator.New()
	for _, ticket := range ticketInfo {
		fmt.Println(ticket)
		validation_err := v.Struct(ticket)
		if validation_err != nil {
			out, _ := json.Marshal(model.ErrorResponse{Error_Message: validation_err.Error(), Error_Code: 5001})
			http.Error(w, string(out), 500)
			return
		}
	}

	if err != nil {
		logger.Log().Errorf("Errror caught in BookTicket : %s", err)
		writeErrorResponse(err, w)
		return
	}

	result, reserveErr := ticketManager.ReserveTicket(ticketInfo)
	if reserveErr != nil {
		writeErrorResponse(reserveErr, w)
		return
	}
	util.WriteResponseMessage(w, http.StatusOK, result)
}

func CancelTicket(w http.ResponseWriter, r *http.Request) {

	payload, readErr := ioutil.ReadAll(r.Body)
	stackMessage := "Error in CancelTicket"
	if readErr != nil {
		logger.Log().Infof("Error while unmarshaling data: %s\n", readErr)
		err := errors.WrapError(errconst.INVALID_JSON_DATA, readErr, stackMessage)
		writeErrorResponse(err, w)
		return
	}
	var cancelTickets model.ReserveTicketResult
	err := json.Unmarshal(payload, &cancelTickets)
	logger.Log().Info(cancelTickets)
	if err != nil {
		logger.Log().Errorf("Errror caught in CancelTicket : %s", err)
		writeErrorResponse(err, w)
		return
	}
	if len(cancelTickets.TicketNumbers) <= 0 {
		err := errors.NewCustomErr(errconst.INVALID_JSON_DATA, stackMessage)
		writeErrorResponse(err, w)
		return
	}
	_, reserveErr := ticketManager.CancelReservation(cancelTickets)
	if reserveErr != nil {
		writeErrorResponse(reserveErr, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func UpdateReservation(w http.ResponseWriter, r *http.Request) {
	payload, readErr := ioutil.ReadAll(r.Body)
	stackMessage := "Error in UpdateReservation"
	if readErr != nil {
		logger.Log().Infof("Error while unmarshaling data: %s\n", readErr)
		err := errors.WrapError(errconst.INVALID_JSON_DATA, readErr, stackMessage)
		writeErrorResponse(err, w)
		return
	}
	var updateTickets []model.UpdateTickets
	err := json.Unmarshal(payload, &updateTickets)

	v := validator.New()

	for _, ticket := range updateTickets {
		fmt.Println(ticket)
		validation_err := v.Struct(ticket)
		if validation_err != nil {
			out, _ := json.Marshal(model.ErrorResponse{Error_Message: validation_err.Error(), Error_Code: 5001})
			http.Error(w, string(out), 500)
			return
		}
	}
	if err != nil {
		logger.Log().Errorf("Errror caught in UpdateReservation : %s", err)
		writeErrorResponse(err, w)
		return
	}
	_, reserveErr := ticketManager.UpdateReservation(updateTickets)
	if reserveErr != nil {
		writeErrorResponse(reserveErr, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func writeErrorResponse(err error, w http.ResponseWriter) string {
	msg, code := err.(*errors.CustomError).GetErMsgAndCode()
	logger.Log().Debugf("Error message : %s Error code : %d", msg, code)
	out, marshal_err := json.Marshal(model.ErrorResponse{Error_Message: err.Error(), Error_Code: code})
	if marshal_err != nil {
		fmt.Printf("Error caught in preference_service_handler method (errorresponse): %v", marshal_err)
	}
	if code > 5000 {
		http.Error(w, string(out), 500)
	} else {
		http.Error(w, string(out), 200)
	}
	return string(out)
}
