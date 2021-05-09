package service

import (
	"encoding/json"
	"fmt"
	"strings"
	errconstant "ticketing-service/constants"
	"ticketing-service/db"
	errors "ticketing-service/error"
	"ticketing-service/logger"
	model "ticketing-service/model"
	notification "ticketing-service/notification"
)

var dbprovider db.Data_provider
var notification_manager notification.NotificationManager = notification.GetNotificationManagerInstance()

type TicketingServiceManager interface {
	ReserveTicket(ticket []model.Ticket) ([]byte, error)
	CancelReservation(ticket model.ReserveTicketResult) ([]byte, error)
	UpdateReservation(ticket []model.UpdateTickets) ([]byte, error)
}

type TicketManager struct {
}

func GetTicketManager() TicketingServiceManager {
	dbprovider = db.GetDataProviderInstance()
	return &TicketManager{}
}

func (ticketManager TicketManager) ReserveTicket(ticket []model.Ticket) ([]byte, error) {
	stackMessage := "Error in ReserveTicket"
	logger.Log().Info("ReserveTicket")
	ticket_number, err := dbprovider.InsertReservationData(ticket)
	if err != nil {
		return nil, errors.NewCustomErr(errconstant.INVALID_TICKET_ID, stackMessage)
	}
	ticketNumber := make([]int, 0)
	ticketNumber = append(ticketNumber, ticket_number)
	for i := 1; i < len(ticket); i++ {
		ticket_number = ticket_number + 1
		ticketNumber = append(ticketNumber, ticket_number)
	}
	var ticketResults model.ReserveTicketResult
	ticketResults.TicketNumbers = ticketNumber
	result, marshalErr := json.Marshal(ticketResults)
	if marshalErr != nil {
		return nil, errors.NewCustomErr(errconstant.JSON_MARSHAL_ERR, stackMessage)
	}
	logger.Log().Info(result)

	messageJson, messageErr := json.Marshal(ticket)
	if messageErr != nil {
		logger.Log().Errorf(errors.GetMessage(errconstant.JSON_MARSHAL_ERR)+" : %v", messageErr)
		return nil, errors.NewCustomErr(errconstant.JSON_MARSHAL_ERR, stackMessage)
	}

	_, notifyErr := notification_manager.NotifyRMQSubscribers(strings.Trim(strings.Replace(fmt.Sprint(ticketNumber), " ", ",", -1), "[]"), notification.ReserveTicketEvent, string(messageJson))
	if notifyErr != nil {
		logger.Log().Errorf(errors.GetMessage(errconstant.FAILED_TO_NOTIFY_QUEUE)+" : %v", notifyErr)
		return nil, errors.NewCustomErr(errconstant.FAILED_TO_NOTIFY_QUEUE, stackMessage)
	}

	return result, nil
}

func (ticketManager TicketManager) CancelReservation(ticket model.ReserveTicketResult) ([]byte, error) {
	logger.Log().Infof("CancelReservation")
	stackMessage := "Error in CancelReservation"
	_, failedList, _ := dbprovider.CancelReservationData(ticket)
	if len(failedList) > 0 {
		logger.Log().Info(failedList)
		return nil, errors.NewCustomErrWithMsg("Tickets not found :"+strings.Trim(strings.Replace(fmt.Sprint(failedList), " ", ",", -1), "[]"), errconstant.INVALID_TICKET_ID, stackMessage)
	}
	messageJson := "{\"msg\":\"Successfully cancelled reservation for tickets :" + strings.Trim(strings.Replace(fmt.Sprint(ticket.TicketNumbers), " ", ",", -1), "[]") + "}"
	_, notifyErr := notification_manager.NotifyRMQSubscribers(strings.Trim(strings.Replace(fmt.Sprint(ticket.TicketNumbers), " ", ",", -1), "[]"), notification.CancelTicketEvent, string(messageJson))
	if notifyErr != nil {
		logger.Log().Errorf(errors.GetMessage(errconstant.FAILED_TO_NOTIFY_QUEUE)+" : %v", notifyErr)
		return nil, errors.NewCustomErr(errconstant.FAILED_TO_NOTIFY_QUEUE, stackMessage)
	}
	return nil, nil
}

func (ticketManager TicketManager) UpdateReservation(ticket []model.UpdateTickets) ([]byte, error) {
	stackMessage := "Error in UpdateReservation"
	logger.Log().Info("UpdateReservation")
	_, failedList, _ := dbprovider.UpdateReservationData(ticket)
	if len(failedList) > 0 {
		logger.Log().Info(failedList)
		var ticketResults model.ReserveTicketResult
		ticketResults.TicketNumbers = failedList
		result, marshalErr := json.Marshal(ticketResults)
		if marshalErr != nil {
			return nil, errors.NewCustomErr(errconstant.JSON_MARSHAL_ERR, stackMessage)
		}
		logger.Log().Info(result)
		return nil, errors.NewCustomErrWithMsg("Tickets not found :"+strings.Trim(strings.Replace(fmt.Sprint(failedList), " ", ",", -1), "[]"), errconstant.INVALID_TICKET_ID, stackMessage)
	}

	ticketNumber := make([]int, 0)
	for _, ticket_id := range ticket {
		ticketNumber = append(ticketNumber, ticket_id.TicketId)
	}

	messageJson, messageErr := json.Marshal(ticket)
	if messageErr != nil {
		logger.Log().Errorf(errors.GetMessage(errconstant.JSON_MARSHAL_ERR)+" : %v", messageErr)
		return nil, errors.NewCustomErr(errconstant.JSON_MARSHAL_ERR, stackMessage)
	}

	_, notifyErr := notification_manager.NotifyRMQSubscribers(strings.Trim(strings.Replace(fmt.Sprint(ticketNumber), " ", ",", -1), "[]"), notification.UpdateTicketEvent, string(messageJson))
	if notifyErr != nil {
		logger.Log().Errorf(errors.GetMessage(errconstant.FAILED_TO_NOTIFY_QUEUE)+" : %v", notifyErr)
		return nil, errors.NewCustomErr(errconstant.FAILED_TO_NOTIFY_QUEUE, stackMessage)
	}

	return nil, nil
}
