package main

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

func HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "system operational")
}

type AuthUrlResponseBody struct {
	Url string `json:"url"`
}

func PublishGoogleAuthUrl(c echo.Context) error {
	authURL := AuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	c.Logger().Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)
	return c.JSON(200, AuthUrlResponseBody{Url: authURL})
}

type CodeExchangeRequestBody struct {
	Code     string `json:"code"`
	UserExId string `json:"userExId"`
}

func CodeExchange(c echo.Context) error {
	var err error
	body := new(CodeExchangeRequestBody)
	if err = c.Bind(body); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	decodedCode, err := url.QueryUnescape(body.Code)
	if err != nil {
		return c.String(http.StatusBadRequest, "Code is invalid")
	}
	tok, err := AuthConfig.Exchange(context.TODO(), decodedCode)
	if err != nil {
		return c.String(http.StatusBadRequest, "Code is invalid")
	}
	err = DbClient.PutToken(body.UserExId, tok)
	if err != nil {
		c.Logger().Errorf("Got dynamoDb error: %v", err)
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	return c.String(http.StatusOK, "Reserved user tokens")
}

type InsertOneTaskRequestBody struct {
	UserExId string   `json:"userExId"`
	Task     OismTask `json:"task"`
}

type OismTask struct {
	ListName string `json:"listName"`
	Title    string `json:"title"`
	Notes    string `json:"notes,omitempty"`
	Duo      string `json:"duo,omitempty"`
}

type InsertOneTaskResponseBody struct {
	Ticket string `json:"ticket"`
}

func InsertOneTask(c echo.Context) error {
	requestBody := InsertOneTaskRequestBody{}
	if err := c.Bind(&requestBody); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	token, err := DbClient.FetchToken(requestBody.UserExId)
	if err != nil {
		c.Logger().Errorf("Got dynamoDb error: %v", err)
		if IsNotSuchKeyError(err) {
			return c.String(http.StatusNotFound, "No Such UserExId")
		}
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	ticket := publishInsertTaskAction(token, requestBody.Task)
	return c.JSON(http.StatusAccepted, InsertOneTaskResponseBody{Ticket: ticket.String()})
}

type InsertManyTasksRequestBody struct {
	UserExId string     `json:"userExId"`
	Tasks    []OismTask `json:"tasks"`
}

type InsertManyResponseBody struct {
	Tickets []string `json:"tickets"`
}

func InsertManyTasks(c echo.Context) error {
	requestBody := InsertManyTasksRequestBody{}
	if err := c.Bind(&requestBody); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	token, err := DbClient.FetchToken(requestBody.UserExId)
	if err != nil {
		c.Logger().Errorf("Got dynamoDb error: %v", err)
		if IsNotSuchKeyError(err) {
			return c.String(http.StatusNotFound, "No Such UserExId")
		}
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	tickets := []string{}
	for _, task := range requestBody.Tasks {
		ticket := publishInsertTaskAction(token, task)
		tickets = append(tickets, ticket.String())
	}
	return c.JSON(http.StatusAccepted, InsertManyResponseBody{Tickets: tickets})
}

type GoogleActionStatus string

const (
	Started    GoogleActionStatus = "started"
	InProgress GoogleActionStatus = "onGoing"
	Error      GoogleActionStatus = "error"
	Done       GoogleActionStatus = "done"
)

func publishInsertTaskAction(token *oauth2.Token, task OismTask) uuid.UUID {
	ticket := uuid.New()
	GoogleActionStore.Store(ticket, Started)
	go func() {
		GoogleActionStore.Store(ticket, InProgress)
		svs, err := NewTasksService(token)
		if err != nil {
			GoogleActionStore.Store(ticket, Error)
			NotiClient.NotifyInsertTaskError(task, err)
			return
		}
		err = svs.InsertTask(task.ListName, task.Title, task.Notes, task.Duo)
		if err != nil {
			GoogleActionStore.Store(ticket, Error)
			NotiClient.NotifyInsertTaskError(task, err)
			return
		}
		GoogleActionStore.Store(ticket, Done)
	}()
	return ticket
}

type Ticket struct {
	Ticket string `param:"name"`
}

type CheckGoogleActionStatusResponseBody struct {
	Status GoogleActionStatus `json:"status"`
}

func CheckGoogleActionStatus(c echo.Context) error {
	var ticket_s string
	err := echo.PathParamsBinder(c).String("ticket", &ticket_s).BindError()
	if err != nil {
		return c.String(http.StatusBadRequest, "No \"ticket\" Path Parameter")
	}
	ticket, err := uuid.Parse(ticket_s)
	if err != nil {
		return c.String(http.StatusBadRequest, "Ticket Should Be UUID")
	}
	status_i, found := GoogleActionStore.Load(ticket)
	if !found {
		return c.String(http.StatusNotFound, "No Found such ticket")
	}
	status := status_i.(GoogleActionStatus)
	return c.JSON(http.StatusOK, CheckGoogleActionStatusResponseBody{Status: status})
}
