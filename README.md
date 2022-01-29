# URL Shortening Service - GO-FIBER, REDIS, DOCKER

### How to Start Service

- Download or Clone the repository.
- Find .env.example file in the /api directory and get rid of .example from the file extension.
- Run the following command in your terminal:

```shell
docker-compose up -d
```

### Test the Service Using cURL

```shell
curl --header "Content-Type: application/json" -d "{\"url\":\"https://
en.wikipedia.org/wiki/URL_shortening\"}" http://localhost:3000/api/v1
```

If the POST request is successful, then you will get the following response:

```json
{
    "url": "https://en.wikipedia.org/wiki/URL_shortening",
    "short": "localhost:3000/454276",
    "expiry": 24,
    "rate_limit": 8,
    "rate_limit_reset": 28
}
```

- As you see in the resulting response, now you have the shortened url "localhost:3000/454276". 
- You can copy the url not including double quotes and then paste it into your web browser. 
- You will be able to be redirected to the actual webpage.
