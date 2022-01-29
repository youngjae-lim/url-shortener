# URL Shortening Service - GO-FIBER, REDIS, DOCKER

### How to start Service

- Download or Clone the repository.
- Find .env.example file in the /api directory and get rid of .example from the file extension.
- Run the following command in your terminal:

```shell
docker-compose up -d
```

### Test using cURL

```shell
curl --header "Content-Type: application/json" -d "{\"url\":\"https://
en.wikipedia.org/wiki/URL_shortening\"}" http://localhost:3000/api/v1
```
