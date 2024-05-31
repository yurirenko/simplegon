package main

import (
	"fmt"
	"goweb/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// type Item = {
//   id: unsigned int; // whatever ID type our database uses
//   name: string; // max 40 characters
//   price: float; // can't be negative
//   quantity: unsigned int;
//   onSale: boolean;
//   stores?: [Store];
// }

// type Store = {
//   name: string;
//   owner: 'state' | 'private';
//   items: [Item];
// }

type StoreOwner string

const (
	PrivatelyOwned StoreOwner = "private"
	StateOwned     StoreOwner = "state"
)

// https://pkg.go.dev/github.com/go-playground/validator
type Store struct {
	gorm.Model
	ID    uint32     `json:"ID" gorm:"primary_key"`
	Name  string     `json:"name" binding:"required"`
	Owner StoreOwner `json:"owner" binding:"required,oneof=state private"`
	Items []*Item    `gorm:"many2many:stores_items;"`
}
type Item struct {
	gorm.Model
	ID       uint32   `json:"ID"`
	Name     string   `json:"name" binding:"required,max=40"`
	Price    *float32 `json:"price" binding:"required,min=0"`
	Quantity uint32   `json:"quantity" binding:"required,min=0"`
	OnSale   *bool    `json:"onSale" binding:"required" gorm:"column:onSale"`
	Stores   []*Store `gorm:"many2many:stores_items;"`
}

func getItems(c *gin.Context) {
	var items []Item
	name := c.Query("name")

	query := models.DB.Preload("Stores")

	if name != "" {
		query = query.Where("name LIKE ?", fmt.Sprintf("%%%s%%", name))
	}

	query.Find(&items)

	c.JSON(http.StatusOK, gin.H{"data": items})
}

func getStores(c *gin.Context) {
	var stores []Store
	models.DB.Model(&Store{}).Preload("Items").Find(&stores)

	c.JSON(http.StatusOK, gin.H{"data": stores})
}

func getStore(c *gin.Context) {
	var store []Store
	storeId := c.Param("id")

	if storeId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID cannot be empty"})
		return
	}

	models.DB.First(&store, storeId)

	c.JSON(http.StatusOK, gin.H{"data": store})
}

func getStoreItems(c *gin.Context) {
	var store Store

	storeId := c.Param("id")

	if storeId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID cannot be empty"})
		return
	}

	models.DB.Model(&Store{}).Preload("Items").First(&store, storeId)

	c.JSON(http.StatusOK, gin.H{"data": store.Items})
}

func postStore(c *gin.Context) {
	var json Store

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Create(&json)

	println(&json.ID)

	c.JSON(http.StatusOK, gin.H{"result": json})
}

func postItem(c *gin.Context) {
	var json Item

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Create(&json)

	c.JSON(http.StatusOK, gin.H{"result": json})
}

func main() {
	models.InitConnection()

	router := gin.Default()
	router.GET("/items", getItems)
	router.POST("/items", postItem)

	router.GET("/stores", getStores)
	router.GET("/stores/:id", getStore)
	router.GET("/stores/:id/items", getStoreItems)
	router.POST("/stores", postStore)

	router.Run("0.0.0.0:8080")
}
