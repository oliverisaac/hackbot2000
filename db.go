package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
)

type Hack struct {
	gorm.Model
	Victim string `gorm:"type:varchar(20)"`
	Hacker string `gorm:"type:varchar(20)"`
	Team   string `gorm:"type:varchar(50)"`
}

type dbConnection struct {
	host     string
	port     int
	username string
	password string
	name     string
	options  string
}

var sharedDBConfig dbConnection
var sharedDB *gorm.DB

func (c dbConnection) toDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", c.username, c.password, c.host, c.port, c.name, c.options)
}

func dbInit(config dbConnection) *gorm.DB {
	sharedDBConfig = config
	db().Debug().AutoMigrate(&Hack{})
	return db()
}

func db() *gorm.DB {
	var err error
	if sharedDB == nil || sharedDB.DB().Ping() != nil {
		sharedDB, err = gorm.Open("mysql", sharedDBConfig.toDSN())
		log.Println("Connection Established")
	}
	if err != nil {
		log.Panic(err)
	}

	return sharedDB
}

func addHack(victim string, hacker string, team string) {
	db().Create(&Hack{Victim: victim, Hacker: hacker, Team: team})
}

func getTimesHacker(user string, team string) int {
	var count int
	db().Model(&Hack{}).Where("hacker = ? and team = ?", user, team).Count(&count)
	return count
}

func getTimesVictim(user string, team string) int {
	var count int
	db().Model(&Hack{}).Where("victim = ? and team = ?", user, team).Count(&count)
	return count
}

func recentlyHacked(user string, team string) bool {
	var count int
	db().Model(&Hack{}).Where("victim = ? and team = ? and created_at > TIMESTAMP( DATE( DATE_SUB(NOW(), INTERVAL 5 MINUTE) ) )", user, team).Count(&count)
	return count >= 1
}

type LeaderEntry struct {
	User  string
	Score int
}

func getLeaders(team string) []LeaderEntry {
	numResults := 5
	result := make([]LeaderEntry, 0, numResults)
	db().Raw("SELECT IFNULL( hacker, victim ) AS `user`, IFNULL(timesHacker, 0) - IFNULL( timesVictim, 0) AS score FROM ( SELECT * FROM ( SELECT hacker, COUNT(hacker) AS timesHacker FROM hacks WHERE team = ? GROUP BY hacker) s1 LEFT JOIN ( SELECT victim, COUNT(victim) AS timesVictim FROM hacks WHERE team = ? GROUP BY victim) s2 ON s1.hacker = s2.victim UNION SELECT * FROM ( SELECT hacker, COUNT(hacker) AS timesHacker FROM hacks WHERE team = ? GROUP BY hacker) s1 RIGHT JOIN ( SELECT victim, COUNT(victim) AS timesVictim FROM hacks WHERE team = ? GROUP BY victim) s2 ON s1.hacker = s2.victim ) s3 ORDER BY score DESC LIMIT ?", team, team, team, team, numResults).Scan(&result)

	return result
}
