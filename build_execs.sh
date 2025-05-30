#Linux amd64
GOOS=linux GOARCH=amd64 go build -o bluosplayer-linux-amd64 bluosplayer.go
#Windows
GOOS=windows GOARCH=amd64 go build -o bluosplayer-win-amd64.exe bluosplayer.go
#Raspberry arm6 (Pi1, Zero, Zero W)
GOOS=linux GOARCH=arm go build -o bluosplayer-linux-armv6 bluosplayer.go
#Raspberry arm7 (Pi 2,3,4 Zero 2 W)
GOOS=linux GOARCH=arm go build -o bluosplayer-linux-armv7 bluosplayer.go
#Raspberry arm7 (Pi 4, 5 / arm64)
GOOS=linux GOARCH=arm64 go build -o bluosplayer-linux-arm64 bluosplayer.go
#Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o bluosplayer-apple-arm64 bluosplayer.go