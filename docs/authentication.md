# Authentication and User Management

This document provides an overview of how users are authenticated, how their sessions are managed, and how account lifecycle actions are performed.

## Core Entities

### Users
User information is stored in the `auth.users` table. This includes:
- **Username & Password:** For local accounts (standard sign-up).
- **OAuth Data:** Google sub and email for users who sign in with Google.
- **Display Name:** A user-facing name, defaulting to their username.
- **Auth Type:** An enum (`USERNAME_PASSWORD` or `OAUTH_GOOGLE`) indicating the primary authentication method.

### Sessions
Active user sessions are tracked in the `auth.sessions` table.
- **Token:** A unique, random string used as a session identifier.
- **User ID:** Links the session to a specific user.
- **Expiry:** Sessions are set to expire after 10 days by default.

## Key Operations

### Registration and Sign-In
Authentication is primarily handled via REST endpoints:
- `POST /v0/auth/sign-up`: Creates a new user in `auth.users` and starts a session.
- `POST /v0/auth/sign-in`: Authenticates existing users and returns a session token.
- `POST /v0/auth/sign-out`: Deletes the current session token from `auth.sessions`.

Successful sign-in or sign-up sets an `auth` cookie in the user's browser, which contains the session token.

### OAuth (Google)
For Google authentication, the system uses two main routes:
- `GET /oauth/google`: Redirects the user to Google's consent screen.
- `GET /oauth/google/callback`: Processes the response from Google, creates or updates the user in `auth.users`, and establishes a session.

### Account Deletion
Users can delete their accounts via:
- `POST /v0/auth/delete-account`: This action is permanent. Users can choose whether to delete all their study sets or only their private ones. It removes the user from `auth.users`, which triggers a cascading delete of their sessions.

### Profile Updates
A user's display name can be updated via GraphQL:
- **Mutation:** `updateUser(displayName: String)`
- **Returns:** An `AuthedUser` object with the updated details.

## Middleware
The `AuthMiddleware` (defined in `auth/auth_middleware.go`) intercept requests to protected routes. It:
1. Extracts the `auth` cookie.
2. Validates the token against the `auth.sessions` table.
3. Injects the authenticated user's information into the request context, making it available to resolvers and handlers.
