# Twitter Ripper API Documentation

## Overview
This API provides endpoints to interact with Twitter data including tweets, likes, quotes, comments, and reposts. All endpoints require JWT authentication and follow a specific flow.

## Authentication Flow

### 1. User Registration
**Endpoint:** `POST /auth/signup`

**Request:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "message": "successful"
}
```

### 2. User Login
**Endpoint:** `POST /auth/login`

**Request:**
```json
{
  "email": "john@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "email": "john@example.com"
}
```

### 3. Add Twitter Account
**Endpoint:** `POST /twitter/account`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request:**
```json
{
  "username": "your_twitter_username",
  "password": "your_twitter_password"
}
```

**Response:**
```json
{
  "id": "uuid",
  "username": "your_twitter_username",
  "password": "encrypted_password",
  "user_id": "user_uuid"
}
```

## Twitter Operations Flow

### 4. Twitter Login (Required Before Any Twitter Operations)
**Endpoint:** `POST /twitter/login`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request:**
```json
{
  "username": "your_twitter_username"
}
```

**Response:**
```json
{
  "message": "Twitter login started in background"
}
```

**Status:** `202 Accepted` - Login process started in background (takes ~30 seconds)

---

## Twitter Data Endpoints

### 5. Get Tweet Data
**Endpoint:** `POST /twitter/post`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request:**
```json
{
  "username": "your_twitter_username",
  "url": "https://x.com/username/status/1234567890"
}
```

**Response:**
```json
{
  "id": "1234567890",
  "text": "Tweet content here...",
  "username": "author_username",
  "name": "Author Name",
  "timestamp": "2025-01-19T12:00:00Z",
  "likes": 100,
  "retweets": 50,
  "replies": 25,
  "photos": ["https://pbs.twimg.com/media/image1.jpg"],
  "videos": ["https://video.twimg.com/video1.mp4"],
  "gifs": [],
  "avatar": "https://pbs.twimg.com/profile_images/avatar.jpg"
}
```

### 6. Get Tweet Likes
**Endpoint:** `POST /twitter/post/likes`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request:**
```json
{
  "username": "your_twitter_username",
  "url": "https://x.com/username/status/1234567890"
}
```

**Response:**
```json
{
  "likers": [
    {
      "id": "123456789",
      "username": "user1",
      "name": "User One",
      "description": "Bio text here",
      "followers": 1000,
      "following": 500,
      "verified": false,
      "blue_verified": true
    }
  ],
  "count": 1
}
```

### 7. Get Tweet Quotes
**Endpoint:** `POST /twitter/post/quotes`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request:**
```json
{
  "username": "your_twitter_username",
  "url": "https://x.com/username/status/1234567890"
}
```

**Response:**
```json
{
  "quotes": [
    {
      "id": "987654321",
      "text": "Quote tweet with commentary...",
      "username": "quoter_username",
      "name": "Quoter Name",
      "timestamp": "2025-01-19T12:30:00Z",
      "likes": 25,
      "retweets": 5,
      "replies": 2
    }
  ],
  "count": 1
}
```

### 8. Get Tweet Comments/Replies
**Endpoint:** `POST /twitter/post/comments`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request:**
```json
{
  "username": "your_twitter_username",
  "url": "https://x.com/username/status/1234567890"
}
```

**Response:**
```json
{
  "comments": [
    {
      "id": "111222333",
      "text": "This is a reply to the tweet...",
      "username": "replier_username",
      "name": "Replier Name",
      "timestamp": "2025-01-19T13:00:00Z",
      "likes": 10,
      "retweets": 2,
      "replies": 1
    }
  ],
  "count": 1
}
```

### 9. Get Tweet Reposts/Retweets
**Endpoint:** `POST /twitter/post/reposts`

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request:**
```json
{
  "username": "your_twitter_username",
  "url": "https://x.com/username/status/1234567890"
}
```

**Response:**
```json
{
  "reposts": [
    {
      "id": "444555666",
      "username": "retweeter_username",
      "name": "Retweeter Name",
      "description": "User bio here",
      "followers": 2000,
      "following": 800,
      "verified": true,
      "blue_verified": false
    }
  ],
  "count": 1
}
```

## Error Responses

### Authentication Errors
```json
{
  "error": "Missing or invalid token"
}
```
**Status:** `401 Unauthorized`

### Twitter Session Errors
```json
{
  "error": "No valid Twitter session found. Please login first using /twitter/login"
}
```
**Status:** `401 Unauthorized`

```json
{
  "error": "Twitter session expired. Please login again using /twitter/login"
}
```
**Status:** `401 Unauthorized`

### Permission Errors
```json
{
  "error": "Twitter account not found or not owned by user"
}
```
**Status:** `403 Forbidden`

### Validation Errors
```json
{
  "error": "Invalid tweet URL"
}
```
**Status:** `400 Bad Request`

### Server Errors
```json
{
  "error": "Failed to fetch tweet data"
}
```
**Status:** `500 Internal Server Error`

## Usage Flow

1. **Register/Login** → Get JWT token
2. **Add Twitter Account** → Store Twitter credentials
3. **Twitter Login** → Authenticate with Twitter (wait ~30 seconds)
4. **Use Twitter Endpoints** → Get tweet data, likes, quotes, comments, reposts

## Notes

- All Twitter operations require a valid Twitter session
- Twitter login is asynchronous and takes ~30 seconds
- Sessions expire after 24 hours
- All endpoints return JSON responses
- Rate limiting may apply based on Twitter's API limits