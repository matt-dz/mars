# Mars

Your listening history, beautifully organized. Mars automatically creates weekly and monthly playlists from your Spotify listening history.

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Development Setup](#development-setup)
- [Environment Variables](#environment-variables)

## Features

- [x] **Automatic Playlist Generation**: Weekly and monthly playlists created from your Spotify listening history
- [x] **Play Count Tracking**: See how many times you've listened to each track
- [x] **Spotify Integration**: Seamless OAuth integration with automatic token refresh
- [x] **Export to Spotify**: Add generated playlists directly to your Spotify account
- [ ] **Custom Playlists**: Create playlists for any date range

## Tech Stack

- **Backend**: Go 1.25.1
- **Frontend**: SvelteKit + Tailwind CSS
- **Database**: PostgreSQL 18
- **Deployment**: Docker Compose

## Development Setup

### Prerequisites

- Docker & Docker Compose
- Spotify Developer Account (for OAuth credentials)

### Quick Start

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
docker compose -f docker-compose.dev.yaml up -d
```

5. Access the application:
   - **Frontend**: http://localhost:8080
   - **API**: http://localhost:8080/api
   - **API Docs**: http://localhost:8080/docs

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
