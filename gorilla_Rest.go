package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type User struct { // Структура пользователя
	Name    string         `json:"name"`
	Age     int            `json:"age"`
	Friends map[int]string `json:"friends"`
}
type (
	Users   (map[int]User)
	Friends (map[int]string)
)

var (
	//	user  User                 //хранилище для одного пользователя
	users = make(map[int]User) //хранилище для всех пользователей
)

type Friendship struct { // Структура для запроса дружбы (имена переменых)
	SourceId int `json:"sourceId"` //ID инициатора дружбы
	TargetId int `json:"targetId"` //ID принявшего запрос
}

func main() {
	router := mux.NewRouter().StrictSlash(true) //создаем новый маршрутизатор
	//регистрируем иаршруты
	router.HandleFunc("/", Index).Methods("GET")                                 //начальная страница
	router.HandleFunc("/users", userIndex).Methods("GET")                        //получаем всех пользователей
	router.HandleFunc("/users/{userId}", userShow).Methods("GET")                //получаем пользователя по его ID
	router.HandleFunc("/users/friends/{userId}", friendsUserShow).Methods("GET") //получаем друзей пользователя по его ID

	router.HandleFunc("/users", userCreate).Methods("POST") //создаем нового пользователя
	//$ curl -i http://localhost:8080/users -H "content-type: application/json" -d "{\"name\":\"Milli\",\"age\":\"33\",\"friends\":{}}"

	router.HandleFunc("/friends", makeFriends).Methods("POST") //создаем дружеский союз из двух пользователей
	//$ curl -i http://localhost:8080/friends -H "content-type: application/json" -d "{\"sourceId\":1,\"targetId\":2}"

	router.HandleFunc("/users/{userId}", updateAge).Methods("PUT") //изменяем возраст пользователя
	//$ curl -X PUT -H "content-type: application/json" -d "24" -i http://localhost:8080/users/2

	router.HandleFunc("/users/{userId}", deleteUser).Methods("DELETE") //удаляем пользователя по его ID
	//$ curl -X DELETE -i http://localhost:8080/users/1

	log.Println("Слушаем порт :8080")
	log.Fatal(http.ListenAndServe(":8080", router)) //передаем роутер в функцию ListenAndServe
}

//ПОМОЩНИКИ:

// 1. Показываем начальную Index-страницу по URL  http://localhost:8080
func Index(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "         Привет!\n HTTP-сервис ждет команду")
}

// 2. Создаем начальную базу пользователей
func init() {
	repoCreateUser(User{ //пользователь без друзей
		"Adell",
		21,
		Friends{},
	})
	repoCreateUser(User{"Barbora", 22, map[int]string{999: "Gloria"}}) //у пользователя есть друг
}

// 3. Получаем всех пользователей по URL  http://localhost:8080/users
func userIndex(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(users); err != nil { //показываем всех пользователей в хранилище
		http.Error(w, err.Error(), 500)
		return
	}
}

// 4. Присвоение уникального ID новому пользователю
var currentId int //начальный ID регистрации пользователя

func repoCreateUser(user User) Users {
	currentId += 1          // увеличиваем текущий ID для созданного пользователя на "1"
	users[currentId] = user //присваиваем ID новому пользователю
	return users            //возвращаем обновленный список пользователей
}

// 5. Находим пользователя по его ID
func repoFindUser(id int) User { //находим пользователя по его ID
	for key, user := range users {
		if key == id { //если есть совпадение с ID
			return user //возвращаем пользователя
		}
	}
	return User{} //если ничего нет возвращаем пустой слайс пользователя
}

// 6. Получаем пользователя по его ID
func userShow(w http.ResponseWriter, r *http.Request) {
	var (
		userId int
		err    error
	)
	vars := mux.Vars(r) //получаем ID пользователя из запроса

	if userId, err = strconv.Atoi(vars["userId"]); err != nil { //принимаем "строковое" число - возвращем целое
		fmt.Printf("ошибка синтаксиса, получен 'ID' = %v \n", vars["userId"])
	}
	user := repoFindUser(userId) //полученный Id отправляем в хранилище для поиска пользователя

	if user.Name != "" { //если с таким ID пользователь существует, то:
		//показываем ответ в окне браузера по URL  http://localhost:8080/users/id
		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, err.Error(), 400)
		}
		return
	}
	// Если не нашли пользователя, то ошибка 404 (не найдено)
	//формируем ответ
	type jsonErr struct {
		Code int    `json:"code"`
		Text string `json:"text"`
		ID   string `json:"id"`
	}
	json.NewEncoder(w).Encode(jsonErr{Code: http.StatusNotFound, Text: "Нет пользователя:", ID: vars["userId"]})
}

//ОБРАБОТЧИКИ:

// 1. Создаем нового пользователя и присваиваем ему ID
func userCreate(w http.ResponseWriter, r *http.Request) {

	var user User                                                     //хранилище для одного пользователя
	w.Header().Set("Content-Type", "application/json; charset=UTF-8") //формируем заголовок ответа
	err := json.NewDecoder(r.Body).Decode(&user)                      //декодируем запрос JSON
	if err != nil {                                                   //Если при декодировании JSON возникла ошибка,
		w.WriteHeader(http.StatusBadRequest) //возвращается код 400 Bad Request
		http.Error(w, "неправильный, некорректный запрос 'cURL'\n", http.StatusBadRequest)
		return
	}
	defer r.Body.Close() //отложенное закрытие запроса
	//если запрос не полный
	if user.Name == "" && user.Age == 0 {
		w.WriteHeader(http.StatusNoContent) //ошибка "Нет содержимого"
		return
	} else if user.Name == "" {
		w.WriteHeader(http.StatusPartialContent) //ошибка "Частичное содержимое"
		w.Write([]byte(" 206 новая запись не создана Имя не указано\n"))
		return
	}
	newUser := repoCreateUser(user) //получаем из функции нового пользователя с присвоенным ему ID

	//удачное завершение
	//ответ в командной строке
	w.WriteHeader(http.StatusCreated)
	if user.Age == 0 {
		w.Write([]byte(" Создан новый пользователь -> " + user.Name + " (без указания возраста)\n"))
		return
	} else {
		age := strconv.Itoa(user.Age)
		newAge := " Создан новый пользователь -> " + user.Name + " " + age + " лет\n В хранилище:\n"
		w.Write([]byte(newAge))
	}
	//ответ в окне браузера по указанному URL  http://localhost:8080/users
	if err := json.NewEncoder(w).Encode(newUser); err != nil { //показывает список пользователей
		w.WriteHeader(http.StatusInternalServerError) //возвращается код 500 Internal Server Error
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

// 2. Создаем друзей из двух пользователей по их ID
func makeFriends(w http.ResponseWriter, r *http.Request) { //обработчик запроса
	//инициализация переменных
	var union Friendship

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewDecoder(r.Body).Decode(&union); err != nil { //Если при декодировании JSON возникла ошибка,
		w.WriteHeader(http.StatusBadRequest) //возвращается код 400 Bad Request
		http.Error(w, "неправильный, некорректный запрос 'cURL'\n", http.StatusBadRequest)
		return
	}
	defer r.Body.Close() //отложенное закрытие запроса

	source := repoFindUser(union.SourceId) //получаем пользователя инициатора дружбы по его ID
	target := repoFindUser(union.TargetId) //получаем пользователя который примет инициатора в друзья

	if source.Name != "" && target.Name != "" { //проверяем наличие пользователей
		source.Friends[union.TargetId] = target.Name //пополняем карту друзей инициатора
		target.Friends[union.SourceId] = source.Name //пополняем карту друзей принявшего приглашение

		//ответ в окне браузера по URL  http://localhost:8080/users
		w.WriteHeader(http.StatusOK)                             //формируем заголовок ответа
		if err := json.NewEncoder(w).Encode(users); err != nil { //показывает список пользователей
			w.WriteHeader(http.StatusInternalServerError) //возвращается код 500 (внутренняя ошибка сервера)
			http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
		//ответ в командной строке
		makeFriend := source.Name + " и " + target.Name + " теперь друзья\n"
		w.Write([]byte(makeFriend))
		return
	} else {
		w.WriteHeader(http.StatusNotFound) //возвращается код 404 (не найдено)
		w.Write([]byte("Упс! Проверьте ID пользователей \n"))
	}
}

// 3. Удаляем пользователя по его ID
func deleteUser(w http.ResponseWriter, r *http.Request) {
	var userId int

	vars := mux.Vars(r) //получаем map[key:value] с Id пользователя из маршрута key => {userId}:1

	userId, _ = strconv.Atoi(vars["userId"]) //принимаем "строковое" число - возвращем целое

	_, ok := users[userId] //получаем true, если пользователь с таким Id существует
	if ok {                //если true
		for key, user := range users {
			if key == userId { //если ID пользователя совпадает с запрошенным
				delete(users, key) //удаляем пользователя из хранилища
				deleteId := " пользователь " + user.Name + " удален\n В хранилище:\n"
				w.Write([]byte(deleteId))
			}
			for id := range user.Friends { //проверяем хранилище друзей каждого пользователя
				if id == userId { //если Id пользователя совпадает с запрошенным
					delete(user.Friends, id) //удаляем его из друзей оставшихся пользователей
				}
			}
		}
	} else { // Если мы не нашли пользователя, то ошибка 404 (не найдено)
		w.WriteHeader(http.StatusNotFound)
		//ответ в командной строке
		errID := vars["userId"]
		deleteId := "не удается найти пользователя c ID = " + errID + " для удаления\n"
		w.Write([]byte(deleteId))
		return
	}

	json.NewEncoder(w).Encode(users) //показывает список оставшихся пользователей
}

// 4. Показываем друзей пользователя по его ID
func friendsUserShow(w http.ResponseWriter, r *http.Request) {

	var (
		userId  int
		friends string
	)
	vars := mux.Vars(r)

	userId, _ = strconv.Atoi(vars["userId"]) //принимаем "строковое" число - возвращем целое

	user := repoFindUser(userId) //полученный Id отправляем в хранилище для поиска пользователя
	if user.Name != "" {         //если под  таким ID пользователь существует, то:
		for _, friend := range user.Friends { //проверяем хранилище друзей пользователя
			friends = friends + " " + friend + " *"
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8") //формируем заголовок ответа
		//показываем ответ в окне браузера по URL  http://localhost:8080/friends/id
		if err := json.NewEncoder(w).Encode(user.Friends); err != nil {
			http.Error(w, err.Error(), 400)
		}
		show := "пользователь " + user.Name + " дружит с:" + friends + "\n"
		w.Write([]byte(show))
		return
	}
	// Если мы не нашли пользователя
	w.Header().Set("Content-Type", "application/json; charset=UTF-8") //формируем заголовок ответа
	w.WriteHeader(http.StatusNotFound)                                //ошибка 404 (не найдено)
	//формируем ответ
	type jsonErr struct {
		Code int    `json:"code"`
		Text string `json:"text"`
	}
	if err := json.NewEncoder(w).Encode(jsonErr{Code: http.StatusNotFound, Text: "Пользователь не найден"}); err != nil {
		panic(err)
	}
}

// 5. Изменяем возраст пользователя
func updateAge(w http.ResponseWriter, r *http.Request) {
	var (
		userId int
		newAge int
	)
	vars := mux.Vars(r) //получаем map[key:value] с Id пользователя

	userId, _ = strconv.Atoi(vars["userId"]) //принимаем "строковое" число - возвращем целое

	user := repoFindUser(userId) //полученный Id отправляем в хранилище для поиска пользователя

	if user.Name != "" { //если под таким ID пользователь существует, то:
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewDecoder(r.Body).Decode(&newAge); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		defer r.Body.Close() //отложенное закрытие запроса

		for key, user := range users {
			if key == userId { //если Id совпадает с запрошенным
				user.Age = newAge //обновляем возраст
				users[key] = user //вносим обновление в хранилище пользователей
				//формируем ответ в командной строке
				update := "возраст пользователя " + user.Name + " успешно обновлён на " + strconv.Itoa(user.Age) + "\n"
				w.Write([]byte(update))
				return
			}
		}
	}
	// Если мы не нашли пользователя
	w.Header().Set("Content-Type", "application/json; charset=UTF-8") //формируем заголовок ответа
	w.WriteHeader(http.StatusNotFound)                                //то ошибка 404 (не найдено)
	//формируем ответ
	updateFails := "не удается найти пользователя c ID = " + strconv.Itoa(userId) + " для изменения возраста\n"
	w.Write([]byte(updateFails))
	json.NewEncoder(w).Encode(users) //показывает список всех пользователей
}
