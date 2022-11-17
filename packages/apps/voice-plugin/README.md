# Set up Development Environment

```
go install
go generate
go build
```

## Environment Variables

| param                    | meaning                                             |
|--------------------------|-----------------------------------------------------|
| VOICE_API_SERVER_ADDRESS | port for the voice rest api, should be set to 11010 |
| DB_HOST                  | hostname of postgres db                             |
| DB_PORT                  | port of postgres db                                 |
| DB_NAME                  | database name                                       |
| DB_USER                  | user to log into db as                              |
| DB_PASS                  | the database password                               |