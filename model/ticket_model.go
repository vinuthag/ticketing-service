package model

import (
	"fmt"
)

type Ticket struct {
	TicketId int    `json:"ticketId,omitempty"`
	Id       string `json:"id" validate:"required,email"`
	Name     string `json:"name" validate:"required,alpha"`
	To       string `json:"to" validate:"required,alpha"`
	From     string `json:"from" validate:"required,alpha"`
	Date     string `json:"date" validate:"required"`
	Time     string `json:"time" validate:"required"`
}

func (t Ticket) String() string {
	return fmt.Sprintf("TicketId: %s, Id: %s, Name: %s, To:%s, From:%s, Date:%s, Time:%s", t.TicketId, t.Id, t.Name, t.To, t.From, t.Date, t.Time)
}

type ReserveTicketResult struct {
	TicketNumbers []int `json:"reservedTickets"`
}

type UpdateTickets struct {
	TicketId int    `json:"ticket_id" validate:required"`
	To       string `json:"to" validate:"required,alpha"`
	From     string `json:"from" validate:"required,alpha"`
	Date     string `json:"date" validate:"required"`
	Time     string `json:"time" validate:"required"`
}
