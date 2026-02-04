# Mars

Your listening history, beautifully organized. Mars (Music ARchival Software) automatically creates weekly and monthly playlists from your Spotify listening history.

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [Development Setup](#development-setup)
- [Environment Variables](#environment-variables)

## Features

- [x] **Automatic Playlist Generation**: Weekly and monthly playlists created from your Spotify listening history
  <img width="1728" height="910" alt="image" src="https://github.com/user-attachments/assets/4c69a820-d565-45fb-910c-5915ba1c40f3" />

- [x] **Play Count Tracking**: See how many times you've listened to each track
  <img width="1728" height="910" alt="image" src="https://github.com/user-attachments/assets/5f0b08de-ce62-4b8d-942f-08b1bb694432" />
  
- [x] **Top Tracks View**: View your most-played tracks with flexible time period filtering
  - Past 24 hours, 7 days, month-to-date, year-to-date
  - Custom date range selection with intuitive date picker
  <img width="1728" height="910" alt="image" src="https://github.com/user-attachments/assets/7226672f-b307-44b2-bfbf-a38e9544e787" />
    
- [x] **Spotify Integration**: Seamless OAuth integration with automatic token refresh
- [x] **Export to Spotify**: Add generated playlists directly to your Spotify account
- [ ] **Custom Playlists**: Create playlists for any date range

## Tech Stack

- **Backend**: Go 1.25.1
- **Frontend**: SvelteKit + Tailwind CSS
- **Database**: PostgreSQL 18
- **Deployment**: Docker Compose

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Spotify Developer Account (for OAuth credentials)

### Deployment

1. Download the deployment files:
```sh
wget https://raw.githubusercontent.com/matt-dz/mars/refs/heads/main/docker/docker-compose.yaml
wget https://raw.githubusercontent.com/matt-dz/mars/refs/heads/main/docker/fileserver.conf
```

2. Add your Spotify OAuth credentials for the api service:
```txt
api:
  environment:
    SPOTIFY_CLIENT_ID: your_client_id
    SPOTIFY_CLIENT_SECRET: your_client_secret
    SPOTIFY_REDIRECT_URI: http://localhost:8080/api/oauth/spotify
```

3. Set the `ADMIN_EMAIL`, `ADMIN_PASSWORD`, and `DATABASE_PASSWORD` for the api service. Ensure the passwords are secure. Don't forget to set `POSTGRES_PASSWORD` to match `DATABASE_PASSWORD`!
```txt
api:
  environment:
    DATABASE_PASSWORD: your_password
    ADMIN_EMAIL: joe@mars.com
    ADMIN_PASSWORD: secure-password # ensure this matches POSTGRES_PASSWORD in database

database:
  environment:
    POSTGRES_PASSWORD: secure-password # ensure this matches ADMIN_PASSWORD in api

```

4. Start the environment:
```bash
docker compose up -d
```

5. Access the application:
   - **Frontend**: http://localhost:8080
   - **API**: http://localhost:8080/api
   - **API Docs**: http://localhost:8080/docs

## Development Setup

For local development with hot-reloading:

1. Clone the repository:
```bash
git clone https://github.com/matt-dz/mars.git
cd mars
```

2. Create environment file:
```bash
cp .env.backend.example .env.backend
```

3. Add your Spotify OAuth credentials to `.env.backend`:
```bash
SPOTIFY_CLIENT_ID=your_client_id
SPOTIFY_CLIENT_SECRET=your_client_secret
SPOTIFY_REDIRECT_URI=http://localhost:8080/api/oauth/spotify/callback
```

4. Start the development environment:
```bash
docker compose -f docker-compose.dev.yaml up
```

The development setup includes:
- **Hot-reloading** for both frontend and backend
- **Volume mounts** for live code updates
- **Debug capabilities** with stdin/tty enabled

### Default Credentials

- **Email**: admin@example.com
- **Password**: Passw0rds!!!

## Environment Variables

Create a `.env.backend` file with the following variables:

| Variable | Description |
|----------|-------------|
| `SPOTIFY_CLIENT_ID` | Your Spotify app client ID |
| `SPOTIFY_CLIENT_SECRET` | Your Spotify app client secret |
| `SPOTIFY_REDIRECT_URI` | OAuth callback URL |

The Docker Compose setup handles all other configuration automatically.
