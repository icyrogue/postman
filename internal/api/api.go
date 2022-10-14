package api

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type api struct {
	r                *gin.Engine
	storage          Storage
	requestprocessor Requestprocessor
	analyzer         Analyzer
}

type Storage interface {
	Ping(ctx context.Context) error
	Get(ctx context.Context, id string) ([]byte, error)
	GetRead(ctx context.Context, id string) ([]byte, error)
}

type Requestprocessor interface {
	NewList(ctx context.Context) (id string, err error)
	AddUser(id string, body []byte) error
	AddBatch(id string, body []byte) error
}

type Analyzer interface {
	StoreTemplate(id string, data []byte) error
}

func New(storage Storage, requestprocessor Requestprocessor, analyzer Analyzer) *api {
	gin.SetMode(gin.ReleaseMode)
	return &api{r: gin.Default(),
		storage:          storage,
		requestprocessor: requestprocessor,
		analyzer:         analyzer}
}

func (a *api) Init() {
	a.r.GET("/api/ping", a.ping)
	a.r.GET("/api/pingdb", a.pingDB)

	a.r.POST("/api/list", a.newList)
	a.r.GET("/api/list/:id", a.getList)

	a.r.POST("/api/list/:id", a.addUser)
	a.r.POST("/api/list/:id/batch", a.addBatch)
	a.r.GET("/api/list/:id/stat", a.getRead)
	a.r.POST("/api/list/:id/template", a.addTemplate)

}

func (a *api) Run() {
	log.Fatal(a.r.Run())
}

func (a *api) ping(c *gin.Context) {
	c.String(http.StatusOK, "hello from postman")
}

func (a *api) pingDB(c *gin.Context) {
	if err := a.storage.Ping(c); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "DB is ok")
}

func (a *api) newList(c *gin.Context) {
	id, err := a.requestprocessor.NewList(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusCreated, "Cоздан новый список для рассылки с ID %v", id)
}

func (a *api) getList(c *gin.Context) {
	id := c.Param("id")
	c.Header("Content-Type", "application/json")

	users, err := a.storage.Get(c, id)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusOK, string(users))
}

func (a *api) addUser(c *gin.Context) {
	id := c.Param("id")
	body, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if err := a.requestprocessor.AddUser(id, body); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusAccepted, "")
}

func (a *api) addBatch(c *gin.Context) {
	id := c.Param("id")
	body, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if err = a.requestprocessor.AddBatch(id, body); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusAccepted, "")
}

func (a *api) addTemplate(c *gin.Context) {
	id := c.Param("id")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	defer c.Request.Body.Close()

	err = a.analyzer.StoreTemplate(id, body)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	c.String(http.StatusCreated, "добавлен новый шаблон для %v", id)
}

func (a *api) getRead(c *gin.Context) {
	id := c.Param("id")

	c.Header("Content-Type", "application/json")
	users, err := a.storage.GetRead(c, id)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusOK, string(users))
}
