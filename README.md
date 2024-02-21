# Telegram Contacts Search

## Description
This project implements a Telegram bot designed to retrieve telegram username based on provided phone number.

## Pre-requisites
- Docker and Docker Compose
- Telegram Bot API token
- Go 1.22

## Installation
1. Clone the repository
2. Ensure that you already have an active session file in the `sessions` folder. If not, you can create a new session file by running the application locally.
3. Build and start the container using Docker Compose
```bash
docker compose up --build
```

## Usage
The bot is designed to be interacted with through Telegram messages. Once deployed, it can:

- Validate phone numbers and retrieve associated Telegram usernames based on message content.
- Respond to /ping with "pong".

## Configuration
Configure the bot using environment variables specified in the .env file or directly in the compose.yaml:

- `BOT_TOKEN`: Your Telegram Bot Token.
- `APP_HASH`: Your Telegram Application Hash.
- `APP_ID`: Your Telegram Application ID.
- `PHONE`: The phone number associated with the bot account.
- `PASSWORD`: The password for the Telegram account (if applicable).
- `SESSION_FOLDER`: The local directory for storing session data. Default is ./sessions.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
