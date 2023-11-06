package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	BaseURL           = "https://publicapi.battlebit.cloud/"
	LeaderboardGetURL = BaseURL + "Leaderboard/Get"
)

type Category string

const (
	CategoryAll               Category = "all"
	CategoryClans             Category = "clans"
	CategoryXP                Category = "xp"
	CategoryHeals             Category = "heals"
	CategoryRevives           Category = "revives"
	CategoryVehiclesDestroyed Category = "vehiclesDestroyed"
	CategoryVehicleRepairs    Category = "vehicleRepairs"
	CategoryRoadkills         Category = "roadkills"
	CategoryLongestKill       Category = "longestKill"
	CategoryObjectives        Category = "objectives"
	CategoryKills             Category = "kills"
)

var AllCategories = []Category{
	CategoryClans,
	CategoryXP,
	CategoryHeals,
	CategoryRevives,
	CategoryVehiclesDestroyed,
	CategoryVehicleRepairs,
	CategoryRoadkills,
	CategoryLongestKill,
	CategoryObjectives,
	CategoryKills,
}

type PlayerEntry struct {
	Name  string `json:"Name" bson:"Name"`
	Value string `json:"Value" bson:"Value"`
}

type ClanEntry struct {
	Clan       string `json:"Clan" bson:"Clan"`
	Tag        string `json:"Tag" bson:"Tag"`
	XP         string `json:"XP" bson:"XP"`
	MaxPlayers string `json:"MaxPlayers" bson:"MaxPlayers"`
}

type Leaderboard struct {
	Id                    *primitive.ObjectID `bson:"_id,omitempty"`
	TopClans              *[]*ClanEntry       `json:"TopClans,omitempty" bson:"TopClans,omitempty"`
	MostXP                *[]*PlayerEntry     `json:"MostXP,omitempty" bson:"MostXP,omitempty"`
	MostHeals             *[]*PlayerEntry     `json:"MostHeals,omitempty" bson:"MostHeals,omitempty"`
	MostRevives           *[]*PlayerEntry     `json:"MostRevives,omitempty" bson:"MostRevives,omitempty"`
	MostVehiclesDestroyed *[]*PlayerEntry     `json:"MostVehiclesDestroyed,omitempty" bson:"MostVehiclesDestroyed,omitempty"`
	MostVehicleRepairs    *[]*PlayerEntry     `json:"MostVehicleRepairs,omitempty" bson:"MostVehicleRepairs,omitempty"`
	MostRoadkills         *[]*PlayerEntry     `json:"MostRoadkills,omitempty" bson:"MostRoadkills,omitempty"`
	MostKills             *[]*PlayerEntry     `json:"MostKills,omitempty" bson:"MostKills,omitempty"`
}

func StringToCategory(s string) Category {
	switch s {
	case "clans":
		return CategoryClans
	case "xp":
		return CategoryXP
	case "heals":
		return CategoryHeals
	case "revives":
		return CategoryRevives
	case "vehicles_destroyed":
		return CategoryVehiclesDestroyed
	case "vehicle_repairs":
		return CategoryVehicleRepairs
	case "roadkills":
		return CategoryRoadkills
	case "longest_kill":
		return CategoryLongestKill
	case "objectives":
		return CategoryObjectives
	case "kills":
		return CategoryKills
	default:
		return CategoryAll
	}
}

func (c Category) ToMongoField() string {
	switch c {
	case CategoryClans:
		return "TopClans"
	case CategoryXP:
		return "MostXP"
	case CategoryHeals:
		return "MostHeals"
	case CategoryRevives:
		return "MostRevives"
	case CategoryVehiclesDestroyed:
		return "MostVehiclesDestroyed"
	case CategoryVehicleRepairs:
		return "MostVehicleRepairs"
	case CategoryRoadkills:
		return "MostRoadkills"
	case CategoryLongestKill:
		return "MostLongestKill"
	case CategoryObjectives:
		return "MostObjectives"
	case CategoryKills:
		return "MostKills"
	default:
		return ""
	}
}

func FetchLeaderboard() (*Leaderboard, error) {
	resp, err := http.Get(LeaderboardGetURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching leaderboard: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request for leaderboard failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	// the fact i have to do this shit is so stupid, oki please let someone else write your api code

	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf")) // remove BOM

	var j []json.RawMessage

	if err := json.Unmarshal(body, &j); err != nil {
		return nil, fmt.Errorf("error decoding leaderboard: %s", err)
	}

	var lb Leaderboard

	for _, v := range j {
		if err := json.Unmarshal(v, &lb); err != nil {
			return nil, fmt.Errorf("error decoding leaderboard: %s", err)
		}
	}

	return &lb, err
}

func (api *API) UpdateLeaderboardJob() {
	for {
		leaderboard, err := FetchLeaderboard()
		if err != nil {
			log.Printf("non fatal error fetching leaderboard: %s", err)
			time.Sleep(10 * time.Second)
			continue
		}

		api.db.AddLeaderboard(leaderboard)

		time.Sleep(1 * time.Minute)
	}
}
