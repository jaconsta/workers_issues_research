package main

import (
	"fmt"
	"time"

	"github.com/gocelery/gocelery"
	"github.com/gomodule/redigo/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	users "jaconsta/tickets_demo/users/lib"
)

type TaskConfig struct {
	CreateUser    string
	MessageNotify string
}
type RedisConfig struct {
	DbNumber int
	IpFamily int
	Url      string
}
type DatabaseConfig struct {
	Host     string
	Username string
	Password string
	Dbname   string
	Port     string
}

type ConfigStruct struct {
	Tasks    TaskConfig
	Redis    RedisConfig
	Database DatabaseConfig
}

var configStruct ConfigStruct

func databaseConnection() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", configStruct.Database.Host, configStruct.Database.Username, configStruct.Database.Password, configStruct.Database.Dbname, configStruct.Database.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	return db
}

func addUserTask(usersMod *users.UserApp) func(string, string, string) string {
	return func(first string, last string, email string) string {
		usersMod.Create(first, last, email)
		fmt.Printf("%s, %s, %s \n", first, last, email)
		return "got_cha"
	}
}

func redisPoolConnect(url string) *redis.Pool {
	redisPoolBroker := &redis.Pool{
		MaxIdle:     3,                 // maximum number of idle connections in the pool
		MaxActive:   0,                 // maximum number of connections allocated by the pool at a given time
		IdleTimeout: 240 * time.Second, // close connections after remaining idle for this duration
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(url)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	return redisPoolBroker
}

func workersConfig(usersMod users.UserApp) *gocelery.CeleryClient {
	brokerDb := configStruct.Redis.DbNumber * 2
	redisPoolBroker := redisPoolConnect(fmt.Sprintf("%s/%d", configStruct.Redis.Url, brokerDb))
	redisPoolBackend := redisPoolConnect(fmt.Sprintf("%s/%d", configStruct.Redis.Url, brokerDb*2))

	cli, _ := gocelery.NewCeleryClient(
		gocelery.NewRedisBroker(redisPoolBroker),
		&gocelery.RedisCeleryBackend{Pool: redisPoolBackend},
		5, // number of workers
	)

	cli.Register(configStruct.Tasks.CreateUser, addUserTask(&usersMod))
	return cli
}

func main() {
	configStruct = ConfigStruct{
		Tasks: TaskConfig{
			CreateUser:    "users.create",
			MessageNotify: "messages.Notify",
		},
		Redis: RedisConfig{
			DbNumber: 1,
			IpFamily: 4,
			Url:      "redis://localhost:6379",
		},
		Database: DatabaseConfig{
			Host:     "localhost ",
			Username: "johndoe ",
			Password: "randompassword ",
			Dbname:   "users ",
			Port:     "5432 ",
		},
	}

	db := databaseConnection()
	userApp := users.UserAppBuilder(db)

	worker := workersConfig(userApp)

	defer worker.StopWorker()
	fmt.Println("Working?")
	for {
		worker.StartWorker()

		// wait for client request
		time.Sleep(10 * time.Second)
	}
}
