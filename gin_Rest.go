/*	30.5 Практическая работа: написать HTTP-сервис (с JSON-данными)
	Обработчики:
	1. создания пользователя
	2. который по именам делает друзей из двух пользователей
	3. который удаляет пользователя по имени
	4. который по имени пользователя возвращает всех его друзей
	5. который обновляет возраст пользователя по его порядковому номеру в базе
	Дополнительные обработчики:
	1. возврата всех пользователей
	2. возврата определенного пользователя по имени
	3. возврата определенного пользователя по порядковому номеру
*/

package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10" //Пакет предлагает несколько тегов для сравнения
)

// user представляет данные о пользователе.
type User struct {
	Name    string   `json:"name" binding:"required"` //тег требует обязательное заполнение
	Age     int      `json:"age" binding:"min=18"`    //тег ограничивает минимальный возраст
	Friends []string `json:"friends"`
}

var (
	newUser User
	users   []User
	userId  int
	err     error

	validErr validator.ValidationErrors
)

func main() {
	//flag.Parse()
	router := gin.Default()
	//доверенный IP-адрес клиента (желателен для безопасности)
	router.SetTrustedProxies([]string{"127.0.0.1"})
	//создаем начальную базу пользователей (не обязательна)
	users = []User{
		{Name: "Monika", Age: 25, Friends: []string{}},
		{Name: "Barby", Age: 35, Friends: []string{}},
	}

	router.GET("/users", getUsers)                 // http://localhost:8080/users
	router.GET("/users/name/:name", getUserByName) // http://localhost:8080/users/name/Barby
	router.GET("/users/id/:id", getUserByID)       // http://localhost:8080/users/id/2
	router.GET("/friends/:name", getFriends)       // http://localhost:8080/friends/Barby

	router.POST("/users", postUsers)                       //$ curl -X POST -i http://localhost:8080/users -H "content-type: application/json" -d "{\"name\":\"Willy\",\"age\":33,\"friends\":[]}"
	router.PUT("/friends", putFriends)                     //$ curl -X PUT -i http://localhost:8080/friends -H "content-type: application/json" -d "{\"source\":\"Monika\",\"target\":\"Barby\"}"
	router.DELETE("/users/delete/:name", deleteUserByName) //$ curl -X DELETE -i http://localhost:8080/users/delete/Barby
	router.PUT("/users/:id", putAge)                       //$ curl -X PUT -H "content-type: application/json" -d "22" -i http://localhost:8080/users/2

	//По умолчанию слушает "localhost:8080")
	router.Run()

}

// ПОМОЩНИКИ:
// обработка ошибок ввода данных (валидация)
type ErrorMessage struct {
	Field   string `json:"поле"`
	Message string `json:"ошибка"`
}

// функция формирует сообщение при несоответствии поля структуры "User" тегу
func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "Это поле обязательно для заполнения"
	case "min":
		return "Должно быть больше, чем " + fe.Param()
	}
	return "Неизвестная ошибка"
}

// поиск пользователя по его имени
func repoFindUser(name string) User { //находим пользователя по его ID
	for _, user := range users {
		if user.Name == name { //если есть совпадение с ID
			return user //возвращаем пользователя
		}
	}
	return User{} //если ничего нет возвращаем пустой слайс пользователя
}

// ОБРАБОЧИКИ:
// 1. добавляет пользователя из тела запроса
func postUsers(c *gin.Context) {
	// валидация данных запроса
	if err := c.ShouldBindJSON(&newUser); err != nil { // метод получает JSON и пишет в var
		errors.As(err, &validErr)
		out := make([]ErrorMessage, len(validErr))

		for i, fe := range validErr {
			out[i] = ErrorMessage{fe.Field(), getErrorMessage(fe)} // формирование ответа при несоответствии
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": out}) //ошибка (400)
		return
	}

	//проверяем наличие пользователей в базе
	userToAdd := repoFindUser(newUser.Name)
	if userToAdd.Name == newUser.Name {
		c.String(http.StatusForbidden, "Упс! Кто-то уже в базе") //(403)
		return
	}
	// добавить нового пользователя в срез.
	users = append(users, newUser)
	//Ответ в "cmd" (c.String - формирует развернутый ответ)
	c.String(http.StatusCreated, "Создан новый пользователь: %s %d лет\n", newUser.Name, newUser.Age)
	c.IndentedJSON(http.StatusCreated, newUser) //ответ с красивым выводом структуры
	//	c.JSON(http.StatusCreated, newUser) //вывод в строку {ключ:значение,...}
	//	c.JSON(http.StatusCreated, gin.H{"пользователь:": newUser.Name, "сообщение": "создан "})
	//gin.H - это сокращение для map[string]interface{}
}

// 2. делает друзей из двух пользователей
func putFriends(c *gin.Context) {
	var (
		sourceUser User
		targetUser User
		friend     = make(map[string]string, 2)
	)
	if err := c.ShouldBindJSON(&friend); err != nil { //получаем данные из запроса
		c.AbortWithError(http.StatusBadRequest, err) //(400)
		return
	}
	// получаем из "мапы" имена друзей
	sourceName := friend["source"]
	targetName := friend["target"]

	//проверяем наличие пользователей в базе
	sourceUser = repoFindUser(sourceName)
	targetUser = repoFindUser(targetName)
	if sourceUser.Name == "" || targetUser.Name == "" {
		c.String(http.StatusNotFound, "Упс! Кого-то нет в базе") //(404)
		return
	}

	// пополняем хранилища (слайсы) друзей
	for index, user := range users {
		if user.Name == sourceName {
			for _, friend := range user.Friends {
				if friend == targetName {
					c.String(http.StatusForbidden, "Упс! Уже есть такой ДРУГ :)") //(403)
					return
				}
			}
			user.Friends = append(user.Friends, targetName) // друзья инициатора
			users[index] = user                             // обновляем структуру
		}
		if user.Name == targetName {
			for _, friend := range user.Friends {
				if friend == sourceName {
					c.String(http.StatusForbidden, "Упс! Уже есть такой ДРУГ :)") //(403)
					return
				}
			}
			user.Friends = append(user.Friends, sourceName) // друзья принявшего приглашение
			users[index] = user                             // обновляем структуру
		}
	}
	c.String(http.StatusCreated, " %v и %v теперь друзья\n", sourceName, targetName)
	//	c.IndentedJSON(http.StatusCreated, users)
}

// 3. удаляет пользователя по "Name"
func deleteUserByName(c *gin.Context) {
	//получаем имя пользователя
	name := c.Param("name")
	//проверяем наличие пользователя в базе
	userToDelete := repoFindUser(name)
	if userToDelete.Name == "" {
		c.String(http.StatusNotFound, "Упс! Кого-то нет в базе") //(404)
		return
	}

	for i, user := range users {
		for j, friend := range user.Friends { //удаляем пользователя из хранилищ друзей
			if friend == name {
				user.Friends = append(user.Friends[:j], user.Friends[j+1:]...)

			}
			users[i] = user
		}
	}
	for i, user := range users {
		if user.Name == name { //удаляем пользователя
			//удалить значение текущего индекса (пользователя) и сдвинуть все последующие влево
			users = append(users[:i], users[i+1:]...)
			c.String(http.StatusOK, "Пользователь %v удален", user.Name)
		}
	}
}

//Более эффективно:
//		users[i] = users[len(users)-1]	//копируем последний элемент в индекс "i"
//		users = users[:len(users)-1]	//усечь срез (удаляем значение последнего индекса)
//		users[i] = ""					//удалить последний элемент (записать нулевое значение)

// 4. возвращает всех друзей пользователя
func getFriends(c *gin.Context) {
	name := c.Param("name")
	for _, user := range users { // поиск пользователя по имени
		if user.Name == name {
			c.IndentedJSON(http.StatusOK, user.Friends)
			return
		}
	}
	c.String(http.StatusNotFound, "пользователь %v не найден. Введите имя", name)
}

// 5. изменяет возраст пользователя по порядковому номеру "id"
func putAge(c *gin.Context) {

	var newAge int
	id := c.Param("id")
	if userId, err = strconv.Atoi(id); err != nil { //принимаем "строковое" число - возвращем целое
		fmt.Printf("ошибка синтаксиса, получен 'ID' = %v \n", id)
	}
	if err := c.ShouldBindJSON(&newAge); err != nil { //получаем значение из JSON
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) //(400)
		return
	}
	for i, user := range users {
		if newAge < 18 {
			c.JSON(http.StatusForbidden, gin.H{"error": "ограничение в доступе для клиента"}) // (403)
			return
		} else if i == userId-1 {
			user.Age = newAge
			users[i] = user
			c.String(http.StatusOK, "Возраст пользователя: %s изменен на %d лет\n", user.Name, user.Age)
			c.IndentedJSON(http.StatusOK, user) //(200)
			return
		}
	}
	c.String(http.StatusNotFound, "пользователь с ID = %s не найден\n", id)
}

// Дополнительные обработчики:
// 1. отвечает списком всех пользователей в формате JSON
func getUsers(c *gin.Context) { //используется для получения запроса JSON
	c.IndentedJSON(http.StatusOK, users) //вывод блоками
	//c.JSON(http.StatusOK, users) //вывод в строку
}

// 2. показывает пользователя по "Name"
func getUserByName(c *gin.Context) {
	name := c.Param("name")
	for _, us := range users { // поиск пользователя по имени
		if us.Name == name {
			c.IndentedJSON(http.StatusOK, us)
			return
		}
	}
	c.String(http.StatusNotFound, "пользователь %v не найден. Введите имя", name)
}

// 3. показывает пользователя по порядковому номеру "id"
func getUserByID(c *gin.Context) {
	id := c.Param("id")
	if userId, err = strconv.Atoi(id); err != nil { //принимаем "строковое" число - возвращем целое
		c.String(http.StatusBadRequest, "ошибка синтаксиса, получен 'ID' = %v \n", id)
		return
	}
	for i, us := range users { // поиск пользователя по "id"
		if i == userId-1 {
			c.IndentedJSON(http.StatusOK, us)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"Упс": "пользователь не найден"})
}
