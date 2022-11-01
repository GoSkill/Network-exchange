# Network-exchange
Использованы: пакет "net/http" и фреймворки "Горилла", "Джин", "Го-Чи"

30.5 Практическая работа

Цель практической работы: 
разработать сервис с запросами POST, GET, PUT, DELETE;
использовать localhost:8080

    1. Сделть обработчик создания пользователя с полями: имя, возраст и массив друзей.
    $ curl -i http://localhost:8080/users -H "content-type: application/json" -d "{\"name\":\"Milli\",\"age\":\"33\",\"friends\":{}}"
    //посмотреть запись в базе по ID : http://localhost:8080/user

    2. Сделать обработчик, который делает друзей из двух пользователей по их ID.
    $ curl -X PUT -H "content-type: application/json" -d "24" -i http://localhost:8080/users/2
    //посмотреть запись в базе по ID : http://localhost:8080/users/friends/1
    
    3. Сделать обработчик, который удаляет пользователя по его ID из хранилища и стирает его из массива друзей у всех его друзей.
    $ curl -X DELETE -i http://localhost:8080/users/1
    
    4. Сделать обработчик, который возвращает всех друзей пользователя.
    //посмотреть запись в базе по ID : http://localhost:8080/users/friends/1
    
    5. Сделать обработчик, который обновляет возраст пользователя.
    $ curl -X PUT -H "content-type: application/json" -d "24" -i http://localhost:8080/users/2
    //посмотреть запись в базе по ID : http://localhost:8080/user/id
