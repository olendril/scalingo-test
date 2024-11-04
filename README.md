# Canvas for Backend Technical Test at Scalingo


## Execution

```
docker compose up
```

Application will be then running on port `5000`

The Env variable "GITHUB_ACCESS_TOKEN" can be set to use a personal access token to prevent api rate limiting issue

## Test

```
$ curl localhost:5000/ping
{ "status": "pong" }
```

The [Documentation](./api.yaml) is available to use the service.


