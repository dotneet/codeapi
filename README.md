## Usage

Prepare a docker image.

```bash
./build-image.sh
```

Edit settings.

```bash
cp .env-example .env
vim .env
```

Run the server.

```bash
go run main.go
```

Example of a request:

```bash
curl -X POST -H "Content-Type: application/json" -d @- http://localhost:8080/api/run
{"code": "print('Hello World!')"}
^D
```
