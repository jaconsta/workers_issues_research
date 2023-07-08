package users

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

func createUser(db *gorm.DB, firstName string, lastName string, email string) User {
	user := User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}
	db.Create(&user)
	return user
}

func getOneById(db *gorm.DB, id string) User {
	var user User
	db.First(&user, id)
	return user
}

type UserApp struct {
	Db *gorm.DB
}

func (this *UserApp) Create(firstName string, lastName string, email string) User {
	return createUser(this.Db, firstName, lastName, email)
}

func (this *UserApp) GetByID(id string) User {
	return getOneById(this.Db, id)
}

func UserAppBuilder(db *gorm.DB) UserApp {
	db.AutoMigrate(&User{})
	return UserApp{Db: db}
}
