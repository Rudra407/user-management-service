# User Management Microservice

A microservice for user account management with multi-tenant (organization) support built with Go, Echo, and PostgreSQL.

## Features

- Organization management (multi-tenant support)
  - Create organizations with admin users
  - Update organization details
  - Delete organizations
  - List organizations and their users
- User management
  - User registration within organizations
  - User login with organization context
  - JWT authentication with organization scope
  - Fetch user profile by ID
  - Update user profile
  - Update user roles (admin, user)
  - Delete user account
  - List users within an organization
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

#### 1. Organization Creation
```
┌─────────┐      ┌─────────────────────┐     ┌─────────────────────┐     ┌────────────────────┐     ┌────────────┐
│  Client │──┬──>│ Create Organization │────>│ OrganizationService │────>│ OrganizationRepo   │────>│ PostgreSQL │
└─────────┘  │   └─────────────────────┘     └─────────────────────┘     └────────────────────┘     └────────────┘
             │          │                            │                           │                        │
             │          │                            │                           │                        │
             │          │                            │                           │                        │
             │          │                            │<──────────────────────────│<─────────────────────────┤
             │          │<───────────────────────────│                           │                        │
             │<─────────│                            │                           │                        │
└─────────┘      └─────────────────────┘     └─────────────────────┘     └────────────────────┘     └────────────┘
```
1. Client sends organization creation data (name, display_name, description, website, admin details)
2. Handler validates request format
3. Service creates organization and admin user account
4. Repository creates organization and user records
5. Response with created organization and admin details

#### 2. User Registration (Organization-specific)
```
┌─────────┐      ┌─────────────┐     ┌──────────────┐     ┌────────────┐     ┌────────────┐
│  Client │──┬──>│ Register API │────>│ UserService  │────>│ UserRepo   │────>│ PostgreSQL │
└─────────┘  │   └─────────────┘     └──────────────┘     └────────────┘     └────────────┘
             │          │                    │                   │                  │
             │          │                    │                   │                  │
             │          │                    │<──────────────────│<─────────────────┤
             │          │<───────────────────│                   │                  │
             │<─────────│                    │                   │                  │
└─────────┘      └─────────────┘     └──────────────┘     └────────────┘     └────────────┘
```
1. Client sends user registration data (name, email, password, organization_id, role)
2. Handler validates request format
3. Service validates business rules, checks for duplicate email within the organization
4. Service verifies the organization exists
5. Repository creates new user record with hashed password and organization association
6. Response with created user details (password excluded)

#### 3. User Login (Organization-specific)
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
1. Client sends credentials (email, password, organization_id)
2. Handler validates request format
3. Service authenticates user by finding by email + organization and validating password
4. JWT token is generated with user ID, organization ID, and user role
5. Response with JWT token

#### 4. Fetch User Profile
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
2. JWT middleware validates token and extracts user ID, organization ID, and role
3. Handler receives authorized request with user context
4. Service fetches user by ID
5. Repository retrieves user from database
6. Response with user details (password excluded)

#### 5. List Organization Users (Admin only)
```
┌─────────┐      ┌───────────────┐     ┌─────────────────┐     ┌─────────────────┐     ┌────────────┐     ┌────────────┐
│  Client │──┬──>│JWT+Admin      │────>│ Org Users API   │────>│ OrgService      │────>│ UserRepo   │────>│ PostgreSQL │
│         │  │   │Middleware     │     │                 │     │                 │     │            │     │            │
└─────────┘  │   └───────────────┘     └─────────────────┘     └─────────────────┘     └────────────┘     └────────────┘
             │          │                      │                       │                      │                 │
             │          │                      │                       │                      │                 │
             │          │                      │                       │<─────────────────────│<────────────────┤
             │          │                      │<──────────────────────│                      │                 │
             │          │<─────────────────────│                       │                      │                 │
             │<─────────│                      │                       │                      │                 │
└─────────┘      └───────────────┘     └─────────────────┘     └─────────────────┘     └────────────┘     └────────────┘
```
1. Client sends request with JWT containing admin role
2. JWT middleware validates token and verifies admin role
3. Handler receives authorized request
4. Service fetches organization users with pagination
5. Repository retrieves users for the organization
6. Response with list of users and pagination metadata

## Project Structure

```
user-management-service/
├── api/
│   ├── handlers/        # HTTP handlers
│   │   ├── organization_handler.go # Organization endpoints
│   │   └── user_handler.go         # User endpoints
│   └── middleware/      # Echo middleware
├── cmd/
│   └── server/          # Application entry point
├── config/              # Configuration
├── internal/
│   ├── models/          # Database models
│   │   ├── organization.go  # Organization model
│   │   └── user.go          # User model
│   ├── repositories/    # Data access layer
│   │   ├── organization_repository.go # Organization data access
│   │   └── user_repository.go         # User data access
│   └── services/        # Business logic
│       ├── organization_service.go # Organization operations
│       └── user_service.go         # User operations
├── tests/               # Tests
├── utils/               # Utility packages
├── Dockerfile           # Docker build file
├── docker-compose.yml   # Docker Compose file
├── go.mod               # Go modules file
└── go.sum               # Go modules checksums
```

## API Endpoints

### Organization Endpoints

| Method | URL                        | Description                | Auth Required | Admin Required |
|--------|----------------------------|----------------------------|--------------|---------------|
| POST   | /api/organizations         | Create organization        | No           | No            |
| GET    | /api/organizations         | List organizations         | Yes          | No            |
| GET    | /api/organizations/:id     | Get organization by ID     | Yes          | No            |
| PUT    | /api/organizations/:id     | Update organization        | Yes          | Yes           |
| DELETE | /api/organizations/:id     | Delete organization        | Yes          | Yes           |
| GET    | /api/organizations/:id/users | List organization users  | Yes          | Yes           |

### User Endpoints

| Method | URL                       | Description                | Auth Required | Admin Required |
|--------|---------------------------|----------------------------|--------------|---------------|
| POST   | /api/register             | Register new user          | No           | No            |
| POST   | /api/login                | Login                      | No           | No            |
| GET    | /api/users/profile        | Get user profile           | Yes          | No            |
| GET    | /api/users/:id            | Get user by ID             | Yes          | No            |
| PUT    | /api/users                | Update user                | Yes          | No            |
| PUT    | /api/users/:id/role       | Update user role           | Yes          | Yes           |
| DELETE | /api/users                | Delete current user        | Yes          | No            |
| DELETE | /api/users/:id            | Delete user by ID          | Yes          | Yes           |
| GET    | /api/users                | List users (org-scoped)    | Yes          | No            |
| GET    | /health                   | Health check               | No           | No            |

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

## Complete API Usage Examples

Below are curl commands for the main API flows:

### Organization Flow

```bash
# 1. Create a new organization with an admin user
curl -X POST http://localhost:8080/api/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "acme",
    "display_name": "ACME Corporation",
    "description": "A fictional company",
    "website": "https://acme.example.com",
    "admin_name": "Admin User",
    "admin_email": "admin@acme.example.com",
    "admin_password": "secure_password"
  }'

# 2. Login as the admin user
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@acme.example.com",
    "password": "secure_password",
    "organization_id": 1
  }'
# Response contains JWT token to use in subsequent requests

# 3. Get organization details
curl -X GET http://localhost:8080/api/organizations/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 4. Update organization
curl -X PUT http://localhost:8080/api/organizations/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "display_name": "ACME Corp Updated",
    "description": "Updated company description",
    "website": "https://acme-updated.example.com"
  }'

# 5. List all organizations
curl -X GET "http://localhost:8080/api/organizations?page=1&per_page=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 6. List organization users
curl -X GET "http://localhost:8080/api/organizations/1/users?page=1&per_page=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### User Flow Within Organization

```bash
# 1. Register a new user in the organization
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@acme.example.com",
    "password": "password123",
    "organization_id": 1,
    "role": "user"
  }'

# 2. Login as the new user
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@acme.example.com",
    "password": "password123",
    "organization_id": 1
  }'
# Response contains JWT token to use in subsequent requests

# 3. Get user profile
curl -X GET http://localhost:8080/api/users/profile \
  -H "Authorization: Bearer USER_JWT_TOKEN"

# 4. Get user by ID
curl -X GET http://localhost:8080/api/users/2 \
  -H "Authorization: Bearer USER_JWT_TOKEN"

# 5. Update user profile
curl -X PUT http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer USER_JWT_TOKEN" \
  -d '{
    "name": "John Doe Updated",
    "email": "john.updated@acme.example.com"
  }'

# 6. Admin updating user role (requires admin token)
curl -X PUT http://localhost:8080/api/users/2/role \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -d '{
    "role": "admin"
  }'

# 7. List users in organization
curl -X GET "http://localhost:8080/api/users?page=1&per_page=10" \
  -H "Authorization: Bearer USER_JWT_TOKEN"

# 8. Delete user (self)
curl -X DELETE http://localhost:8080/api/users \
  -H "Authorization: Bearer USER_JWT_TOKEN"

# 9. Admin deleting user by ID (requires admin token)
curl -X DELETE http://localhost:8080/api/users/2 \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"

# 10. Delete organization (requires admin token)
curl -X DELETE http://localhost:8080/api/organizations/1 \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

## License

This project is licensed under the MIT License. 