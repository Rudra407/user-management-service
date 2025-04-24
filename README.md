# User Management Microservice

A microservice for user account management built with Go, Echo, and PostgreSQL.

## Features

- User registration
- User login with JWT authentication
- Fetch user profile by ID
- Update user profile
- Delete user account
- RESTful API design
- Clean architecture (handlers, services, repositories)
- Structured logging with request IDs
- PostgreSQL database with GORM
- Containerized with Docker

## Architecture and Flow Diagrams

The service follows a clean architecture pattern with:
- API handlers (presentation layer)
- Services (business logic layer)
- Repositories (data access layer)
- Models (domain entities)

### API Endpoint Flows

#### 1. User Registration
```
┌─────────┐      ┌─────────────┐     ┌──────────────┐     ┌────────────┐     ┌────────────┐
│  Client │──┬──>│ Register API │────>│ UserService  │────>│ UserRepo   │────>│ PostgreSQL │
└─────────┘  │   └─────────────┘     └──────────────┘     └────────────┘     └────────────┘
             │          │                    │                   │                  │
             │          │                    │                   │                  │
             │          │                    │                   │                  │
             │          │                    │<──────────────────│<─────────────────┤
             │          │<───────────────────│                   │                  │
             │<─────────│                    │                   │                  │
└─────────┘      └─────────────┘     └──────────────┘     └────────────┘     └────────────┘
```
1. Client sends user registration data (name, email, password)
2. Handler validates request format
3. Service validates business rules and checks for duplicate email
4. Repository creates new user record with hashed password
5. Response with created user details (password excluded)

#### 2. User Login
```
┌─────────┐      ┌─────────────┐     ┌──────────────┐     ┌────────────┐     ┌────────────┐
│  Client │──┬──>│  Login API  │────>│ UserService  │────>│ UserRepo   │────>│ PostgreSQL │
└─────────┘  │   └─────────────┘     └──────────────┘     └────────────┘     └────────────┘
             │          │                    │                   │                  │
             │          │                    │<──────────────────│<─────────────────┤
             │          │                    │                   │                  │
             │          │                    │                   │                  │
             │          │<───────────────────│                   │                  │
             │<─────────│                    │                   │                  │
└─────────┘      └─────────────┘     └──────────────┘     └────────────┘     └────────────┘
```
1. Client sends credentials (email, password)
2. Handler validates request format
3. Service authenticates user by finding by email and validating password
4. JWT token is generated with user ID and expiration
5. Response with JWT token

#### 3. Fetch User Profile
```
┌─────────┐      ┌─────────────┐     ┌────────────┐     ┌──────────────┐     ┌────────────┐     ┌────────────┐
│  Client │──┬──>│JWT Middleware│────>│ Profile API│────>│ UserService  │────>│ UserRepo   │────>│ PostgreSQL │
└─────────┘  │   └─────────────┘     └────────────┘     └──────────────┘     └────────────┘     └────────────┘
             │          │                   │                   │                   │                  │
             │          │                   │                   │                   │                  │
             │          │                   │                   │<──────────────────│<─────────────────┤
             │          │                   │<──────────────────│                   │                  │
             │          │<──────────────────│                   │                   │                  │
             │<─────────│                   │                   │                   │                  │
└─────────┘      └─────────────┘     └────────────┘     └──────────────┘     └────────────┘     └────────────┘
```
1. Client sends request with JWT in Authorization header
2. JWT middleware validates token and extracts user ID
3. Handler receives authorized request with user ID
4. Service fetches user by ID
5. Repository retrieves user from database
6. Response with user details (password excluded)

#### 4. Update User Profile
```
┌─────────┐      ┌─────────────┐     ┌────────────┐     ┌──────────────┐     ┌────────────┐     ┌────────────┐
│  Client │──┬──>│JWT Middleware│────>│ Update API │────>│ UserService  │────>│ UserRepo   │────>│ PostgreSQL │
└─────────┘  │   └─────────────┘     └────────────┘     └──────────────┘     └────────────┘     └────────────┘
             │          │                   │                   │                   │                  │
             │          │                   │                   │                   │                  │
             │          │                   │                   │<──────────────────│<─────────────────┤
             │          │                   │<──────────────────│                   │                  │
             │          │<──────────────────│                   │                   │                  │
             │<─────────│                   │                   │                   │                  │
└─────────┘      └─────────────┘     └────────────┘     └──────────────┘     └────────────┘     └────────────┘
```
1. Client sends update data (name, email, password) with JWT
2. JWT middleware validates token and extracts user ID
3. Handler validates request format
4. Service validates business rules (e.g., checking for duplicate email if email changed)
5. Repository updates user in database
6. Response with updated user details (password excluded)

#### 5. Delete User Account
```
┌─────────┐      ┌─────────────┐     ┌────────────┐     ┌──────────────┐     ┌────────────┐     ┌────────────┐
│  Client │──┬──>│JWT Middleware│────>│ Delete API │────>│ UserService  │────>│ UserRepo   │────>│ PostgreSQL │
└─────────┘  │   └─────────────┘     └────────────┘     └──────────────┘     └────────────┘     └────────────┘
             │          │                   │                   │                   │                  │
             │          │                   │                   │                   │                  │
             │          │                   │                   │<──────────────────│<─────────────────┤
             │          │                   │<──────────────────│                   │                  │
             │          │<──────────────────│                   │                   │                  │
             │<─────────│                   │                   │                   │                  │
└─────────┘      └─────────────┘     └────────────┘     └──────────────┘     └────────────┘     └────────────┘
```
1. Client sends request with JWT
2. JWT middleware validates token and extracts user ID
3. Service verifies user exists
4. Repository soft-deletes user in database (sets DeletedAt field)
5. Response with success message

## Project Structure

```
user-management-service/
├── api/
│   ├── handlers/        # HTTP handlers
│   └── middleware/      # Echo middleware
├── cmd/
│   └── server/          # Application entry point
├── config/              # Configuration
├── internal/
│   ├── models/          # Database models
│   ├── repositories/    # Data access layer
│   └── services/        # Business logic
├── tests/               # Tests
├── utils/               # Utility packages
├── Dockerfile           # Docker build file
├── docker-compose.yml   # Docker Compose file
├── go.mod               # Go modules file
└── go.sum               # Go modules checksums
```

## API Endpoints

| Method | URL                 | Description        | Auth Required |
|--------|---------------------|--------------------|--------------|
| POST   | /api/register       | Register new user  | No           |
| POST   | /api/login          | Login              | No           |
| GET    | /api/users/profile  | Get user profile   | Yes          |
| GET    | /api/users/:id      | Get user by ID     | Yes          |
| PUT    | /api/users          | Update user        | Yes          |
| DELETE | /api/users          | Delete user        | Yes          |
| GET    | /api/users          | List users         | Yes          |
| GET    | /health             | Health check       | No           |

## Setup Instructions

### Prerequisites

- Go 1.23 or higher
- Docker and Docker Compose (for containerized setup)
- PostgreSQL (for local development without Docker)

### Option 1: Using Docker Compose (Recommended)

1. Clone this repository:
   ```bash
   git clone https://github.com/YOUR_USERNAME/user-management-service.git
   cd user-management-service
   ```

2. Start the services using Docker Compose:
   ```bash
   docker-compose up -d
   ```

3. The API will be available at http://localhost:8000.

### Option 2: Running Locally

1. Clone this repository:
   ```bash
   git clone https://github.com/YOUR_USERNAME/user-management-service.git
   cd user-management-service
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Make sure you have PostgreSQL running and accessible.

4. Create a `.env` file with your configuration:
   ```
   SERVER_PORT=8080
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=your_db_user
   DB_PASSWORD=your_db_password
   DB_NAME=user_management
   DB_SSLMODE=disable
   JWT_SECRET=your-jwt-secret-key
   JWT_EXPIRY=24
   LOG_LEVEL=info
   ```

5. Run the application:
   ```bash
   go run cmd/server/main.go
   ```

### Running Tests

To run the tests:

```bash
go test ./tests/...
```

## Complete Flow Example

Here's a shell script that demonstrates a complete flow from user registration to deletion:

```bash
#!/bin/bash
set -e  # Exit on error

echo "1. REGISTERING A NEW USER..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8000/api/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe", "email":"john.doe@example.com", "password":"password123"}')
echo "$REGISTER_RESPONSE" | jq .
echo

echo "2. LOGGING IN..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8000/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john.doe@example.com", "password":"password123"}')
echo "$LOGIN_RESPONSE" | jq .
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r .data.token)
echo

echo "3. GETTING USER PROFILE..."
PROFILE_RESPONSE=$(curl -s -X GET http://localhost:8000/api/users/profile \
  -H "Authorization: Bearer $TOKEN")
echo "$PROFILE_RESPONSE" | jq .
USER_ID=$(echo "$PROFILE_RESPONSE" | jq -r .data.id)
echo

echo "4. GETTING USER BY ID ($USER_ID)..."
USER_BY_ID_RESPONSE=$(curl -s -X GET "http://localhost:8000/api/users/$USER_ID" \
  -H "Authorization: Bearer $TOKEN")
echo "$USER_BY_ID_RESPONSE" | jq .
echo

echo "5. UPDATING USER PROFILE..."
UPDATE_RESPONSE=$(curl -s -X PUT http://localhost:8000/api/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe Updated"}')
echo "$UPDATE_RESPONSE" | jq .
echo

echo "6. LISTING ALL USERS..."
LIST_RESPONSE=$(curl -s -X GET "http://localhost:8000/api/users?page=1&per_page=10" \
  -H "Authorization: Bearer $TOKEN")
echo "$LIST_RESPONSE" | jq .
echo

echo "7. CHECKING SERVICE HEALTH..."
HEALTH_RESPONSE=$(curl -s -X GET http://localhost:8000/health)
echo "$HEALTH_RESPONSE" | jq .
echo

echo "8. DELETING USER..."
DELETE_RESPONSE=$(curl -s -X DELETE http://localhost:8000/api/users \
  -H "Authorization: Bearer $TOKEN")
echo "$DELETE_RESPONSE" | jq .
echo

echo "9. VERIFYING DELETION BY TRYING TO LOGIN AGAIN..."
LOGIN_AFTER_DELETE_RESPONSE=$(curl -s -X POST http://localhost:8000/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john.doe@example.com", "password":"password123"}')
echo "$LOGIN_AFTER_DELETE_RESPONSE" | jq .
```

## License

This project is licensed under the MIT License. 