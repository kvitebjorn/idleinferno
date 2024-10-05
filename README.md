# idleinferno

An IdleRPG inspired by Dante's Inferno.

Runs on port 33379 by default.

**Build commands**
- **_GO install_**
```
go install github.com/kvitebjorn/idleinferno/idleinferno-client@latest
```
```
go install github.com/kvitebjorn/idleinferno/idleinferno-server@latest
```

- **_Native Linux_**

```
go build -o ./bin/server/idleinferno-server ./idleinferno-server && go build -o ./bin/client/idleinferno-client ./idleinferno-client
```

- **_Native Windows_**

```
go build -o ./bin/server/idleinferno-server.exe ./idleinferno-server && go build -o ./bin/client/idleinferno-client.exe ./idleinferno-client
```

- **_Cross-platform (Windows example)_**

```
env GOOS=windows GOARCH=amd64 go build -o ./bin/server/idleinferno-server.exe ./idleinferno-server && env GOOS=windows GOARCH=amd64 go build -o ./bin/client/idleinferno-client.exe ./idleinferno-client
```

- **_Cross-platform (Raspberry Pi 5 example)_**

```
env GOOS=linux GOARCH=arm GOARM=5 go build -o ./bin/server/idleinferno-server-arm ./idleinferno-server && env GOOS=linux GOARCH=arm GOARM=5 go build -o ./bin/client/idleinferno-client-arm ./idleinferno-client
```
