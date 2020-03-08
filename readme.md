## Example Go API Server

This project is to demonstrate a simple example of Go API server.

- The API is a HTTP RESTful JSON API.
- Go-Chi as the routing library.
- MySQL as the database.
- Basic authentication mechanism using token.
- There are basic units tests on the API and database.

## Running the server

The server arguments can be passed either using environment vars or flags.

The server has 4 arguments, as explained below with its type

- port: integer - the server port, default is `8001`
- db_host: integer - the database host, default is `127.0.0.1`
- db_port: integer - the database port, default is `3306`
- db_user: string - the database username, default is `root`
- db_password: string - the database password, default is `12345`
- db_name: string - the database name, default is `go_sample_api_server_structure`

If you use the default arguments, the the API is available on `http://localhost:8001`

The server will create a default user with username=`username` and password=`password`.

Below are the instruction how to run the server, either locally or using docker-compose.

### Run Locally

To build and run the server locally, you need Go on your machine.

The server expects a MySQL instance running, with root user enabled and password of `12345`.

The simplest way to start the database is by using the Docker. You can use the provided bash script to start the MySQL container.

```bash
# Start the MySQL container
[terminal 1] bash scripts/run.mysql.sh

# Run migrations
[terminal 2] bash scripts/run.migrate.sh -m

# Run the server using default arguments.
[terminal 2] make run
```

### Run Docker-Compose

You can also use docker-compose. The server will be compiled and ran on a container.

One added advantage is that you don't have to install Go.

```bash
[terminal 1] make run-docker
```

### Notes on running the tests

The MySQL store package has unit tests that expect a MySQL instance with below configurations:
- Host: 127.0.0.1
- Port: 3307
- User: root
- Password: test_password
- Database: test_database 

The unit tests will handle the migrations.

You can start the MySQL container by using `-t` flag of the `run.mysql.sh` script.

```bash
# Start the MySQL container for testing
[terminal 1] bash scripts/run.mysql.sh -t
```

## Endpoints
#### Login - POST /login

Request
```json
{
	"username": "username2",
	"password": "password2"
}
```
Response
```json
{
  "token": "1SSQteCTkmxqvFRMzeSHOCCFXjU=",
  "user_id": 2,
  "expire_time": "2020-02-20T14:11:16.398328+08:00"
}
```

#### Register - POST /register

Request
```json
{
	"username": "username",
	"password": "password"
}
```
Response
```json
{
  "user_id": 2
}
```

#### GET Profile - GET /me

Response
```json
{
  "user_id": 1,
  "username": "username"
}
```
#### Get Messages - GET /

Require Authorization Bearer header.

Response
```json
{
  "messages": [
    {
      "id": 1,
      "content": "Vanilla Toffee Bar Crunch",
      "sender": "username2",
      "sent_at": "2020-02-19T14:18:18.716031Z",
      "updated_at": "2020-02-19T14:18:18.716031Z"
    }
  ]
}
```

#### Send Message - POST /

Require Authorization Bearer header.

Request
```json
{
  "content": "Vanilla Toffee Bar Crunch",
	"recipients": [
		2
	]
}
```

#### Update Message - POST /{message_id}

Require Authorization Bearer header.

Request
```json
{
  "content": "Vanilla Toffee Bar Crunch",
	"recipients": [
		2
	]
}
```

#### Delete Message - DELETE /{message_id}

Require Authorization Bearer header.
