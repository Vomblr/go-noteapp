# go-noteapp

Test assignment for SOAX

### Project details

To make simple REST app for notes management.

HOST: ```localhost:8888```

Two endpoints are mandatory:

- ```POST /api/addNote``` Creates note item using mandatory JSON fields ```name``` and ```body```

- ```GET /api/note?id=<note_id>``` Gets note item from storage by ```id``` parameter

### Instalation

Clone the repo with ```git clone https://github.com/Vomblr/go-noteapp.git``` then ```cd go-noteapp```

You should have Golang installed in your system. As bonus part for this assignment mysql db was used as storage, so you should have mysql running on your local machine. If you don't have it yet you can simply run it with Docker: ```docker run -d -p 3306:3306 --name mysql -e MYSQL_ROOT_PASSWORD=root mysql```

Create database for your app with command ```mysql -uroot -proot -e 'CREATE DATABASE noteapp'``` or
```docker exec -it mysql mysql -uroot -proot -e 'CREATE DATABASE noteapp'``` if you used Docker.

Then install all modules needed for the app with command ```go get -d ./...```

### Run

Compile the binary with command ```go build``` and execute ```./go-noteapp```

### Extra

As extra part making that assingment I studied basics of building secure REST app using Mux Router and GORM. Tried to handle all possible errors like JSON syntax errors or unsupported media.

### Enjoy


