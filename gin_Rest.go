/*1.  вызов обработчика создания пользователя
$  curl http://localhost:8080/users/3  --include --header "Content-Type: application/json" --request "POST" --data  `{ID: "3", Name: "Петров", Age: 35, Friends: []string{"Сидоров"}}`
							{ID: "4", Name: "Петров", Age: 35, Friends: []string{"Сидоров"}}
$ curl -h "Content-Type: application/json" -d `{id: "3", name: "Петров", age: 35, friends: []string{"Сидоров"}}` http://localhost:8080/users

2. вызов обработчика, который делает друзей из двух пользователей

3. вызов обработчика, который удаляет пользователя

4. вызов обработчика, который возвращает всех друзей пользователя

5. вызов обработчика, который обновляет возраст пользователя

6. вызов обработчик для возврата всех пользователей
$  curl http://localhost:8000/users

7. вызов обработчика для возврата определенного пользователя
$  curl http://localhost:8000/users/1
*/

package main

import (
	_ "encoding/json"
	"flag"
	"net/http"
	_ "os"

	"github.com/gin-gonic/gin"
)

// user представляет данные о пользователе.
type User struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Friends []string `json:"friends"`
}

var users = []User{
	{ID: "1", Name: "Иванов", Age: 35, Friends: []string{"Петров"}},
	{ID: "4", Name: "Петров", Age: 35, Friends: []string{"Сидоров"}},
}

func main() {
	flag.Parse()

	router := gin.Default()

	router.SetTrustedProxies([]string{"127.0.0.1"})

	router.GET("/users", getUsers)
	router.GET("/users/:id", getUserByID)
	router.POST("/users", postUsers)

	router.Run("localhost:8000")
}

// отвечает списком всех пользователей в формате JSON.
func getUsers(c *gin.Context) { //используется для получения запроса JSON
	c.IndentedJSON(http.StatusOK, users)
	//c.Context.JSON(http.StatusOK, users)
}

// добавляет пользователя из JSON, полученного в теле запроса.
func postUsers(c *gin.Context) {

	var newUser User

	// Вызов BindJSON, чтобы связать полученный JSON с newUser.
	if err := c.BindJSON(&newUser); err != nil { //BindJSON(&var) функция получает JSON и пишет в var
		return
	}
	// Добавить нового пользователя в срез.
	users = append(users, newUser)
	c.IndentedJSON(http.StatusCreated, newUser)
}

// находит пользователя по параметру id,
// и возвращает этотого пользователя в качестве ответа.
func getUserByID(c *gin.Context) {
	id := c.Param("id")

	// поиск пользователя, значение ID которого соответствует параметру
	for _, a := range users {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"Упс": "пользователь не найден"})
}

/*	ПРИМЕРЫ
type StructA struct {
    FieldA string `form:"field_a"`
}
type StructB struct {
    NestedStruct StructA	//поле вложенной структуры StructA
    FieldB string `form:"field_b"`	поле структуры StructB
}
функция Bind(&var) получает из var и пишет в JSON
func GetDataB(c *gin.Context) {
    var b StructB
    c.Bind(&b)
    c.JSON(200, gin.H{
        "a": b.NestedStruct, //поля структуры "StructB"
        "b": b.FieldB,
    })
}
func main() {
    r := gin.Default() //запуск роутера
    r.GET("/getb", GetDataB) //обработчик GET

    r.Run(:8000) //слушаем URL
	запуск $ 	curl "http://localhost:8080/getb?field_a=hello&field_b=world"
	результат	{"a":{"FieldA":"hello"},"b":"world"}
}


*/
