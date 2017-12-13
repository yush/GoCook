# Building for Raspberry 2

GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc go build

# create directories

db
|- Images
||- import
||- original
|- gocook.db3

# Crontab
@reboot ./start.sh
Every day restart

# Update Database

# copy

public
views

# Run

env GOCOOK_BASEDIR="/home/pi/GoCook/" ./GoCook