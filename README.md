# About 
Watch your favourite Tiktok and YouTube videos from Telegram directly!
All you have to do is to type a link in whatever chat bot is in

## Running bot on your Windows PC
1. Install GO runtime from the [official website](https://go.dev/).
2. Open **.env** and pass down your telegram bot token.
3. Run `go run main.go` or build `go build .`

## PostgreSQL database
This bot uses a PostgreSQL database instance. Make sure to supply a connection string over your `.env` file.
A new db instance can be created via `docker-compose` file or manual setup. All tables are automatically created by code once connected to a db.

## Proxying
This bot supports proxying. Supply credentials in **.env** file or leave them empty if not needed.

## FFmpeg
This bot has a ffmpeg dependency. Be sure to install it if running locally.

## Linux and Docker support
1. Install [Docker](https://www.docker.com/) on your main OS.
2. Build and run an image `docker-compose up -d --build`