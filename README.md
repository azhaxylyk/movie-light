# Movie Light

**Movie Light** is a web application for searching, selecting, and watching movies together.

## Features

- 🔍 Movie search with TMDB API  
- 🎴 Swipe-style movie selection for groups  
- 🎬 Synchronized movie trailer viewing  
- 💬 Real-time chat during movie sessions

## Getting Started

### 1. Clone the Repository

```bash
git clone git@github.com:azhaxylyk/movie-light.git
cd movie-light
```

### 2. Set Up Environment Variables

Create a `.env` file in the root directory and fill it according to the example:

```env
# Google OAuth Credentials
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret

# General API Key
API_KEY=your_api_key

# YouTube API Key
YOU_TUBE_API_KEY=your_youtube_api_key
```

### 3. Run the Application

Use the following command to start the application:

```bash
go run ./cmd
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.