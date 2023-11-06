package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type API struct {
	*iris.Application

	db *Database
}

func NewAPI(config *Config) *API {
	app := iris.New()
	db := connectDatabase(config.MongoURI)

	api := &API{app, db}

	api.registerErrors()
	api.registerRoutes()

	return api
}

func (api *API) StartCron() {
	go api.db.RemoveOldEntriesJob()
	go api.UpdateLeaderboardJob()
}

func (api *API) registerErrors() {
	api.OnErrorCode(iris.StatusNotFound, func(ctx iris.Context) {
		ctx.JSON(iris.Map{"message": "not found"})
	})
}

func (api *API) registerRoutes() {
	api.Get("/", func(ctx iris.Context) {
		// serve the rendered openapi specification

		ctx.JSON(iris.Map{"wys g": "wagwan g"})
	})

	apiRoot := api.Party("/api", func(ctx iris.Context) {
		log.Printf("Request: %s %s", ctx.Method(), ctx.Path())
		ctx.Next()
	})

	v1 := apiRoot.Party("/v1")

	v1.Get("/leaderboard", api.GetLeaderboard)
}

func (api *API) GetLeaderboard(ctx iris.Context) {
	params := ctx.URLParams()

	category, ok := params["category"]
	if !ok {
		category = "all"
	}

	var categories []Category

	if category == "all" {
		categories = AllCategories
	} else {
		categories = make([]Category, 0)

		for _, c := range strings.Split(category, ",") {
			category := StringToCategory(c)
			switch category {
			case CategoryAll:
				categories = append(categories, CategoryAll)
				break
			case CategoryClans:
				categories = append(categories, CategoryClans)
			case CategoryXP:
				categories = append(categories, CategoryXP)
			case CategoryHeals:
				categories = append(categories, CategoryHeals)
			case CategoryRevives:
				categories = append(categories, CategoryRevives)
			case CategoryVehiclesDestroyed:
				categories = append(categories, CategoryVehiclesDestroyed)
			case CategoryVehicleRepairs:
				categories = append(categories, CategoryVehicleRepairs)
			case CategoryKills:
				categories = append(categories, CategoryKills)
			default:
				ctx.StatusCode(iris.StatusBadRequest)
				ctx.JSON(iris.Map{"error": fmt.Sprintf("invalid category: %s", c)})
				return
			}
		}
	}

	page, ok := params["page"]
	if !ok {
		page = "1"
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "invalid page, must be an integer"})
		return
	}

	if pageInt < 1 || pageInt > 249 {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "invalid page"})
		return
	}

	var cursorID *primitive.ObjectID
	cursor := params["cursor"]
	if cursor != "" {
		cursorObjID, err := primitive.ObjectIDFromHex(cursor)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.JSON(iris.Map{"error": "invalid cursor"})
			return
		}
		cursorID = &cursorObjID
	}

	lb, err := api.db.GetLeaderboard(categories, cursorID, pageInt)
	if err != nil {
		log.Println(err)
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "failed to get leaderboard"})
		return
	}

	ctx.JSON(lb)
}
