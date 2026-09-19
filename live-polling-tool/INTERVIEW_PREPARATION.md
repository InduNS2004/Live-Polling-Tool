# Interview Preparation

## Why React?
React makes the UI component-based and state-driven. A WebSocket message can update result state, and React re-renders only the affected UI.

## Why Go + Gin?
Go gives a small, fast backend with straightforward concurrency. Gin provides routing, middleware and JSON handling without putting business logic into the router.

## Why MongoDB?
Polls and options are naturally document-shaped, while votes are kept separately for a clean audit trail and unique voting constraint.

## Why Redis?
Redis is used for live result caching and Pub/Sub. Pub/Sub allows multiple Go instances to receive the same vote event.

## Why Redis Pub/Sub?
A vote received by one backend process must reach WebSocket clients connected to any backend process. Redis is the shared event bus between those processes.

## Why WebSockets?
The server can push a new result immediately. The browser does not need to poll or refresh.

## Authentication
Passwords are hashed with bcrypt. Login returns a signed JWT. Protected Gin routes validate the signature and extract the user ID. Ownership is checked again in database filters for management actions.

## Duplicate votes
The server creates a random voter cookie. MongoDB has a unique compound index for `(poll_id, voter_id)`. The database therefore rejects a second vote for the same anonymous browser identity even if the frontend is bypassed.

## Backend validation
The server trims and checks question length, option count, option length, duplicate options, poll status and option membership. Frontend validation is only a UX aid.

## Concurrent votes
Each vote increments only its selected option using MongoDB's atomic `$inc`. The result is then read from MongoDB and published as a complete snapshot, reducing stale-count ambiguity.

## MongoDB + Redis
MongoDB owns durable state. Redis owns ephemeral delivery state. Redis never replaces the durable vote record.

## Live update flow
`POST /vote` → validate → MongoDB vote + counter → read result → Redis SET + PUBLISH → WebSocket subscriber → React state update.

## Scaling
Multiple Go instances can sit behind a load balancer. Redis Pub/Sub distributes events across instances, while MongoDB provides shared persistence. In production, WebSocket-aware load balancing and managed Redis/MongoDB are recommended.

## If Redis goes down
REST voting can still persist to MongoDB, but the live-event path is unavailable. The application reports degraded health and the WebSocket endpoint returns an explicit realtime-unavailable response instead of pretending it is live.

## If MongoDB goes down
Votes cannot be safely accepted because MongoDB is the source of truth. The health endpoint reports the database as disconnected and operations fail with structured errors.

## CORS
The Go backend allows the configured `FRONTEND_URL`, selected methods and headers. Production should use the exact deployed frontend origin.

## Deployment
Build the React app and host its static assets. Build the Go backend as a container or Go service. Supply managed MongoDB and Redis URLs as environment variables. Enable HTTPS and secure cookies in a production deployment.

## Biggest technical challenge
The important design challenge is consistency: a live UI must not drift from the persisted vote count. The implementation publishes a fresh MongoDB-derived result after each successful vote rather than incrementing an independent UI counter.

## Future improvements
- Verified voter accounts or signed participation links for stronger identity.
- Rate limiting backed by Redis.
- Poll expiry jobs.
- Rich analytics and export.
- Redis Streams or another durable event mechanism if event replay becomes a requirement.
- Automated end-to-end browser tests.
