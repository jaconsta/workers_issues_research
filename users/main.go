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

type ConfigStruct struct {
	Tasks TaskConfig
	Redis RedisConfig
}

var configStruct ConfigStruct

func databaseConnection() *gorm.DB {
	dsn := "host=localhost user=johndoe password=randompassword dbname=users port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	return db
}

// exampleAddTask is integer addition task
// with named arguments
type userAddTask struct {
	// userApp   users.UserApp
	firstName string
	lastName  string
	email     string
}

// ParseArgs is not defined in CeleryTask interface, only kwargs
func (a *userAddTask) ParseArgs(args []string) error {
	a.firstName = args[0]
	a.lastName = args[1]
	a.email = args[2]
	return nil
}

// func (a *userAddTask) ParseKwargs(kwargs map[string]interface{}) error {
// 	kwargA, ok := kwargs[""]
// 	if !ok {
// 		return fmt.Errorf("undefined kwarg a")
// 	}
// 	kwargAFloat, ok := kwargA.(float64)
// 	if !ok {
// 		return fmt.Errorf("malformed kwarg a")
// 	}
// 	a.firstName = int(kwargAFloat)
// 	kwargB, ok := kwargs["b"]
// 	if !ok {
// 		return fmt.Errorf("undefined kwarg b")
// 	}
// 	kwargBFloat, ok := kwargB.(float64)
// 	if !ok {
// 		return fmt.Errorf("malformed kwarg b")
// 	}
// 	a.b = int(kwargBFloat)
// 	return nil
// }

func (a *userAddTask) RunTask() (interface{}, error) {
	fmt.Println("got task")
	result := a.firstName + " " + a.lastName + " " + a.email
	return result, nil
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
	// create redis connection pool
	brokerDb := configStruct.Redis.DbNumber * 2
	redisPoolBroker := redisPoolConnect(fmt.Sprintf("%s/%d", configStruct.Redis.Url, brokerDb))
	redisPoolBackend := redisPoolConnect(fmt.Sprintf("%s/%d", configStruct.Redis.Url, brokerDb*2))

	cli, _ := gocelery.NewCeleryClient(
		gocelery.NewRedisBroker(redisPoolBroker),
		&gocelery.RedisCeleryBackend{Pool: redisPoolBackend},
		5, // number of workers
	)

	// register task
	cli.Register(configStruct.Tasks.CreateUser, &userAddTask{})
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
