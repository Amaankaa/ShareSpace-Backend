# ShareSpace Backend Documentation

Clean-architecture Go backend for ShareSpace. Features user authentication, posts, resources, comments, mentorship, and real-time messaging over WebSockets. MongoDB is the primary datastore. Files are uploaded to Cloudinary. Email verification and OTP delivery use external providers. HTTP is served with Gin.

## Contents
- Overview
- Architecture and Layout
- Tech Stack
- Environment Variables
- Setup and Run
- Authentication and Roles
- Routes and Endpoints
- Messaging Over WebSockets
- Data Models (High-level)
- Middleware and Security
- Integrations
- Testing
- Deployment
- Notes and Differences vs Initial Plan

---

## Overview
This service powers the ShareSpace platform, connecting students and mentors, enabling experience sharing (posts), curated resources, comments, mentorship flows, and direct messaging.

---

## Architecture and Layout
- Delivery/: HTTP handlers and router wiring
  - `controllers/`: endpoint handlers per module
  - `routers/router.go`: routes and auth guards
  - `main.go`: composition root (services, repos, usecases, controllers, WebSocket hub, router)
- Domain/: Core entities and interfaces (user, post, resource, comment, mentorship, messaging, services, utils)
- Usecases/: Business logic per module
- Repositories/: MongoDB implementations of repositories
- Infrastructure/: Integrations, middleware, and services
  - JWT service, password hashing, email verification/sending, Cloudinary, auth middleware, rate limiter (helper), WebSocket hub
- mocks/: mockery-generated interfaces used in tests
- docs/: additional docs (CI, health endpoint, this documentation)
- Dockerfile, docker-compose*.yml, render.yaml for deployment options

---

## Tech Stack
- Go (Gin HTTP framework)
- MongoDB (go.mongodb.org/mongo-driver)
- JWT (access + refresh)
- EmailListVerify (email validity check) + Brevo (SMTP API)
- Cloudinary (image storage)
- WebSockets (nhooyr/websocket)
- Testing: testify, mongo mtest, mockery

---

## Environment Variables
Required to run the service:

- Core
  - `MONGODB_URI` – Mongo connection string
  - `JWT_SECRET` – HMAC secret for JWT
- Cloudinary
  - `CLOUDINARY_CLOUD_NAME`
  - `CLOUDINARY_API_KEY`
  - `CLOUDINARY_API_SECRET`
- Email
  - `EMAILLISTVERIFY_API_KEY` – EmailListVerify API key
  - `BREVO_API_KEY` – Brevo SMTP API key
  - `FROM_EMAIL` – sender email address
  - `FROM_NAME` – sender display name
- AI (present in wiring; used if resource AI features are enabled)
  - `GEMINI_API_KEY`
  - `GEMINI_API_URL`
- Optional
  - `COOKIE_DOMAIN` – cookie domain on logout; defaults to `localhost`

---

## Setup and Run
Prerequisites: Go, MongoDB, and valid Cloudinary/Email creds.

- Install deps and start the server:

```bash
# from the repository root
go mod tidy
go run ./Delivery/main.go
```

- Health check:
  - GET http://localhost:8080/health

Docker (optional):

```bash
docker compose up --build
```

---

## Authentication and Roles
- JWT is required for protected routes.
- Send the access token as:
  - `Authorization: Bearer <access_token>`
- Claims include `_id`, `username`, `role`, `exp`.
- Roles: `user`, `admin`.
- Admin-only routes require the `AdminOnly()` middleware.

Note: A `RefreshToken` usecase exists and the controller contains a `RefreshToken` handler, but it is not currently exposed in the router. If needed, add an `/auth/refresh` endpoint.

---

## Routes and Endpoints
Base URL: `http://localhost:8080`

### Health
- GET `/health` – service status

### Auth and User
- Public
  - POST `/register` – register user and send verification OTP
  - POST `/verify-user` – verify registration (email + otp)
  - POST `/login` – login with { login: email|username, password }
  - POST `/forgot-password` – send reset OTP
  - POST `/verify-otp` – verify password reset OTP
  - POST `/reset-password` – reset password after OTP verification
  - POST `/auth/refresh` – refresh tokens
- Protected
  - POST `/logout`
  - GET `/profile`
  - PUT `/profile` – multipart form to update profile text fields and optional `profilePicture`

### Posts
- Protected
  - POST `/posts`
  - PATCH `/posts/:id`
  - DELETE `/posts/:id`
  - POST `/posts/:id/like`
  - DELETE `/posts/:id/like`
  - POST `/posts/:id/comments`
  - PATCH `/comments/:commentId`
  - DELETE `/comments/:commentId`
- Public
  - GET `/posts`
  - GET `/posts/search`
  - GET `/posts/popular`
  - GET `/posts/trending-tags`
  - GET `/posts/:id`
  - GET `/posts/:id/comments`
  - GET `/posts/category/:category`
  - GET `/users/:userId/posts`

### Resources
- Protected
  - POST `/resources`
  - PATCH `/resources/:id`
  - DELETE `/resources/:id`
  - POST `/resources/:id/like`
  - DELETE `/resources/:id/like`
  - POST `/resources/:id/bookmark`
  - DELETE `/resources/:id/bookmark`
  - GET `/resources/:id/analytics`
  - POST `/resources/:id/report`
- Public
  - GET `/resources`
  - GET `/resources/search`
  - GET `/resources/popular`
  - GET `/resources/trending`
  - GET `/resources/top-rated`
  - GET `/resources/:id`
  - GET `/users/:userId/resources`
  - GET `/users/:userId/resources/liked`
  - GET `/users/:userId/resources/bookmarked`
  - GET `/users/:userId/resources/stats`

### Mentorship
(Active when the controller is wired.)
- Protected
  - POST `/mentorship/requests`
  - GET `/mentorship/requests/incoming`
  - GET `/mentorship/requests/outgoing`
  - POST `/mentorship/requests/:id/respond`
  - DELETE `/mentorship/requests/:id`
  - GET `/mentorship/connections/:id`
  - GET `/mentorship/connections/mentor`
  - GET `/mentorship/connections/mentee`
  - GET `/mentorship/connections/active`
  - POST `/mentorship/connections/:id/interaction`
  - POST `/mentorship/connections/:id/pause`
  - POST `/mentorship/connections/:id/resume`
  - POST `/mentorship/connections/:id/end`
  - GET `/mentorship/stats`
  - GET `/mentorship/insights`

### Messaging
- Protected REST
  - POST `/conversations` – Body: `{ "participantIds": ["<hex>", ...] }`
  - GET `/conversations?limit=20&offset=0`
  - GET `/conversations/:id/messages?limit=20&offset=0`
- Protected WebSocket
  - GET `/ws` (use Authorization: Bearer token)

### Admin
- Protected + AdminOnly
  - PUT `/user/:id/promote`
  - PUT `/user/:id/demote`
  - POST `/resources/:id/verify`

---

## Messaging Over WebSockets
- Connect: `GET /ws` with `Authorization: Bearer <token>` header
- Client sends frames as JSON:
  - Send message:
    - `{ "type": "message", "conversationId": "<hex>", "content": "..." }`
  - Typing indicator:
    - `{ "type": "typing", "conversationId": "<hex>" }`
- Server broadcast payloads:
  - Message:
    - `{ "type": "message", "message": { id, conversationId, senderId, content, createdAt } }`
  - Typing:
    - `{ "type": "typing", "userId": "<hex>", "ts": "<ISO-8601>" }`
- Behavior:
  - Membership is enforced for sending and listing messages.
  - Rooms are lazily populated from conversation participants.
  - Heartbeats every 30s. Write and read timeouts help detect dead connections.

---

## Data Models (High-level)
- User: auth credentials, profile details, role; tokens and verifications managed in separate collections
- Post: text, media links, category, authorId, likes, timestamps
- Comment: id, postId, authorId, content, timestamps; usecases update post comment counts
- Resource: title, link, category, rating, likes/bookmarks, analytics, moderation state
- Mentorship: requests, connections, statuses, last interaction, stats
- Messaging:
  - Conversation: `{ id, participantIds, createdAt, updatedAt }`
  - Message: `{ id, conversationId, senderId, content, createdAt }`

---

## Middleware and Security
- `Infrastructure/auth_middleWare.go`: validates JWT and sets `user_id`, `username`, and `role` in Gin context
- `Infrastructure/jwt_service.go`: generates and validates tokens (access + refresh)
- `AdminOnly()` guard ensures `role == "admin"`
- Rate limiter helper exists (`Infrastructure/rate_limiter.go`) but is not wired by default
- CORS: not pre-configured; add a Gin CORS middleware if the frontend is on a separate origin

---

## Integrations
- Cloudinary (`Infrastructure/cloudinary_service.go`):
  - `UploadImage(ctx, file, filename)` returns a secure URL
- Email verification (`Infrastructure/email_verifier.go`):
  - Uses EmailListVerify API; returns a boolean validity check
- Email sending (`Infrastructure/email_sender.go`):
  - Uses Brevo SMTP API with sender details from env

---

## Testing
- Run all tests:

```bash
go test ./...
```

- Re-generate mocks (where go:generate is present):

```bash
go generate ./...
```

- Tests cover repositories (with mongo mtest), usecases, and controllers. New tests should use mockery-generated interfaces for isolation.

---

## Deployment
- Docker: `docker-compose.yml` (and prod/staging variants) provided
- Render: `render.yaml` provided
- Nginx: `nginx/nginx.conf` available for reverse-proxy use

---

## Notes and Differences vs Initial Plan
- Framework: Gin is used (rather than net/http + mux/chi)
- Storage: Cloudinary for images (not Firebase Storage)
- Config: environment variables (no Viper)
- Functionality extended beyond the base plan: Comments, Mentorship, Messaging, and Resource analytics/moderation
- Refresh token feature exists but not exposed as a route by default; wire `/auth/refresh` if desired

---

Maintainers can extend this doc with schema details, request/response examples per controller, and SLO/SLA notes as the project evolves.
