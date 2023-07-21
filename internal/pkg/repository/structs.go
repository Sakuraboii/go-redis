package repository

type User struct {
	Id         int64  `db:"id"`
	Name       string `db:"name"`
	SecondName string `db:"second_name"`
	Surname    string `db:"surname"`
}
