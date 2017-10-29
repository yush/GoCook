# Building for Raspberry 2

GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc go build

# create directories

db
|- Images
||- import
||- original
|- gocook.db3

# Update Database

# copy

public
views

# Run

env GOCOOK_BASEDIR="/home/pi/GoCook/" ./GoCook