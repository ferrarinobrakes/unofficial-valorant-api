# Unofficial Valorant API

my very w.i.p and todo attempt at a drop-in replacement for HenrikDev's valorant api using local riot client endpoints. this implementation users a master-client architecture where machines running riot client communicate with a master server via custom TCP protocol  
  
big thanks to [@techchrism](https://github.com/techchrism) for his work on the api documentation

## architecture

- **master server**: exposes connect api for end user, manages clients, implements caching
- **client nodes**: run on machines with riot client, execute LCU requests, report back to master
- **protocol**: custom TCP protocol with length-prefixed protobuf messages for internal communication
- **database**: SQLite with SQLC

## features

- account lookup by name and tag

## api flow

1. **friend request** -> get puuid + region from lcu
2. **map region to shard** -> convert region (e.g., "eu2") to shard (e.g., "eu")
3. **match history** -> fetch player's matches using PUUID
4. **match details** -> extract account level, card, title from match data
5. **cache & return** -> store results and return to client

this gives us the same response for account request that the v1 of henriks api does

## setup

### prerequisites

- go 1.23+
- buf cli
- sqlc
- riot client on client nodes

### build

```bash
buf generate
sqlc generate
go build -o bin/master.exe ./cmd/master
go build -o bin/client.exe ./cmd/client
```

## running

follow `env.example` to set up environment variables, clients will automatically detect riot client and connect to master server

## api usage

### get account

```bash
curl -X POST http://localhost:8081/api.v1.ValorantAPI/GetAccount -H "Content-Type: application/json" -d '{"name":"abcd","tag":"1234"}'
```

response:
```json
{
  "status": 200,
  "data": {
    "puuid": "e83573ec-ec6f-5034-9a38-ed0ccf8dbb1b",
    "region": "eu",
    "account_level": 50,
    "name": "zcerno",
    "tag": "3137",
    "card": "a372c539-41c7-6a41-7064-80a61ccd597b",
    "title": "ed96f7bd-4eed-de28-8c93-40bd313a3157",
    "updated_at": "2025-11-23T12:24:13.848Z"
  }
}
```

### health check

```bash
curl http://localhost:8081/health
```

### testing

not yet 😇😇