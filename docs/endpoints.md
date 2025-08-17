# ShareSpace API Endpoints

Base URL: http://localhost:8080
Authentication: Bearer access token required for protected routes (Authorization: Bearer <token>)

Note: IDs are MongoDB ObjectIDs in hex. Errors return JSON: { "error": string } with appropriate HTTP status.

## Health
- GET /health
  - Response 200: { status, timestamp, version, service }

## Auth & User
- POST /register
  - Body: user { username, fullname, email, password }
  - 201: { message, user, note }
  - 400: { error }
- POST /verify-user
  - Body: { email, otp }
  - 200: { message }
  - 400: { error }
- POST /login
  - Body: { login, password }
  - 200: { user, access_token, refresh_token }
  - 401|400: { error }
- POST /auth/refresh
  - Body: { refresh_token }
  - 200: { accessToken, refreshToken, accessExpiresAt, refreshExpiresAt }
  - 400|401: { error }
- POST /forgot-password
  - Body: { email }
  - 200: { message: "OTP sent" }
  - 400: { error }
- POST /verify-otp
  - Body: { email, otp }
  - 200: { message }
  - 400: { error }
- POST /reset-password
  - Body: { email, new_password }
  - 200: { message }
  - 400: { error }

Protected
- POST /logout
  - 200: { message }
  - 401|500: { error|message }
- GET /profile
  - 200: User
  - 401|404: { error }
- PUT /profile (multipart/form-data)
  - Fields: fullname, bio, phone, website, twitter, linkedin, profilePicture (file)
  - 200: User
  - 400|401: { error }

Admin (Protected + AdminOnly)
- PUT /user/:id/promote
  - 200: { message }
  - 400|401: { error }
- PUT /user/:id/demote
  - 200: { message }
  - 400|401: { error }

## Posts
Protected
- POST /posts
  - Body: CreatePostRequest
  - 201: { message, post: PostResponse }
  - 400|401: { error }
- PATCH /posts/:id
  - Body: UpdatePostRequest
  - 200: { message, post: PostResponse }
  - 400|401|403|404: { error }
- DELETE /posts/:id
  - 200: { message }
  - 400|401|403|404|500: { error }
- POST /posts/:id/like
  - 200: { message }
  - 400|401|404|409|500: { error }
- DELETE /posts/:id/like
  - 200: { message }
  - 400|401|404|409|500: { error }
- POST /posts/:id/comments
  - Body: { content }
  - 201: { message, comment: CommentResponse }
  - 400|401|403|404: { error }
- PATCH /comments/:commentId
  - Body: { content }
  - 200: { message, comment: CommentResponse }
  - 400|401|403|404: { error }
- DELETE /comments/:commentId
  - 200: { message }
  - 400|401|403|404|500: { error }

Public
- GET /posts
  - Query: category, authorId, tag, year, isAnonymous, page, pageSize, sortBy, sortOrder
  - 200: PostListResponse
  - 500: { error }
- GET /posts/search?q=...
  - Query: category, authorId, page, pageSize, sortBy, sortOrder
  - 200: PostListResponse
  - 400|500: { error }
- GET /posts/popular
  - Query: limit, timeframe
  - 200: PostListResponse
  - 500: { error }
- GET /posts/trending-tags
  - Query: limit
  - 200: { tags: string[] }
  - 500: { error }
- GET /posts/:id
  - 200: PostResponse
  - 400|404|500: { error }
- GET /posts/:id/comments
  - Query: page, pageSize, sort
  - 200: CommentListResponse
  - 400|500: { error }
- GET /posts/category/:category
  - Query: page, pageSize, sortBy, sortOrder
  - 200: PostListResponse
  - 400|500: { error }
- GET /users/:userId/posts
  - Query: page, pageSize, sortBy, sortOrder
  - 200: PostListResponse
  - 400|500: { error }

## Resources
Protected
- POST /resources
  - Body: CreateResourceRequest
  - 201: { message, resource: ResourceResponse }
  - 400|401: { error }
- PATCH /resources/:id
  - Body: UpdateResourceRequest
  - 200: { message, resource: ResourceResponse }
  - 400|401|403|404: { error }
- DELETE /resources/:id
  - 200: { message }
  - 400|401|403|404|500: { error }
- POST /resources/:id/like
  - 200: { message }
  - 400|401|404|409|500: { error }
- DELETE /resources/:id/like
  - 200: { message }
  - 400|401|404|409|500: { error }
- POST /resources/:id/bookmark
  - 200: { message }
  - 400|401|404|409|500: { error }
- DELETE /resources/:id/bookmark
  - 200: { message }
  - 400|401|404|409|500: { error }
- GET /resources/:id/analytics
  - 200: ResourceAnalytics
  - 400|401|403|404|500: { error }
- POST /resources/:id/report
  - Body: { reason }
  - 200: { message }
  - 400|401|404|500: { error }
- POST /resources/:id/verify (Admin)
  - 200: { message }
  - 400|401|404|500: { error }

Public
- GET /resources
  - Query: type, category, creatorId, tag, difficulty, isVerified, hasDeadline, page, pageSize, sortBy, sortOrder
  - 200: ResourceListResponse
  - 500: { error }
- GET /resources/search?q=...
  - Query: category, type, page, pageSize, sortBy, sortOrder
  - 200: ResourceListResponse
  - 400|500: { error }
- GET /resources/popular
  - Query: limit, timeframe
  - 200: ResourceListResponse
  - 500: { error }
- GET /resources/trending
  - Query: limit
  - 200: ResourceListResponse
  - 500: { error }
- GET /resources/top-rated
  - Query: limit, category
  - 200: ResourceListResponse
  - 500: { error }
- GET /resources/:id
  - 200: ResourceResponse
  - 400|404|500: { error }
- GET /users/:userId/resources
  - Query: page, pageSize, sortBy, sortOrder
  - 200: ResourceListResponse
  - 400|500: { error }
- GET /users/:userId/resources/liked
  - Query: page, pageSize, sortBy, sortOrder
  - 200: ResourceListResponse
  - 400|500: { error }
- GET /users/:userId/resources/bookmarked
  - Query: page, pageSize, sortBy, sortOrder
  - 200: ResourceListResponse
  - 400|500: { error }
- GET /users/:userId/resources/stats
  - 200: UserResourceStats
  - 400|500: { error }

## Mentorship (Protected)
- POST /mentorship/requests
  - Body: { mentorId, message, topics[] }
  - 201: MentorshipRequest
  - 400|401: { error }
- GET /mentorship/requests/incoming
  - Query: limit, offset
  - 200: { requests: [], total, limit, offset }
  - 401|500: { error }
- GET /mentorship/requests/outgoing
  - Query: limit, offset
  - 200: { requests: [], total, limit, offset }
  - 401|500: { error }
- POST /mentorship/requests/:id/respond
  - Body: { action: "accept"|"decline", message? }
  - 200: MentorshipRequest
  - 400|401: { error }
- DELETE /mentorship/requests/:id
  - 200: { message }
  - 400|401: { error }
- GET /mentorship/connections/:id
  - 200: MentorshipConnection
  - 400|401: { error }
- GET /mentorship/connections/mentor
  - Query: limit, offset
  - 200: { connections: [], total, limit, offset }
  - 401|500: { error }
- GET /mentorship/connections/mentee
  - Query: limit, offset
  - 200: { connections: [], total, limit, offset }
  - 401|500: { error }
- GET /mentorship/connections/active
  - 200: { connections: [] }
  - 401|500: { error }
- POST /mentorship/connections/:id/interaction
  - 200: { message }
  - 400|401: { error }
- POST /mentorship/connections/:id/pause
  - 200: { message }
  - 400|401: { error }
- POST /mentorship/connections/:id/resume
  - 200: { message }
  - 400|401: { error }
- POST /mentorship/connections/:id/end
  - Body: { reason, feedback? }
  - 200: { message }
  - 400|401: { error }
- GET /mentorship/stats
  - 200: MentorshipStats
  - 401|500: { error }
- GET /mentorship/insights
  - 200: MentorshipInsights
  - 401|500: { error }

## Messaging
Protected REST
- POST /conversations
  - Body: { participantIds: string[] }
  - 201: Conversation
  - 400|401: { error }
- GET /conversations
  - Query: limit, offset
  - 200: { conversations: [] } | []
  - 401|500: { error }
- GET /conversations/:id/messages
  - Query: limit, offset
  - 200: { messages: [] } | []
  - 400|401|500: { error }

WebSocket
- GET /ws
  - Auth via Authorization header
  - Client frames:
    - { type: "message", conversationId, content }
    - { type: "typing", conversationId }
  - Server frames:
    - { type: "message", message }
    - { type: "typing", userId, ts }
