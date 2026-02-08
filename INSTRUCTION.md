# Infrantery Implementation Plan

## Overview

This is an application for documenting infrastructure. It manages users, projects, diagrams, markdown notes, diagram node details, and node vaults.

## User Requirements

### Functional

- **User Management**: Authentication and Project membership with **Flexible Permissions**.
- **Projects**: Contain notes, diagrams, and vaults.
- **Notes**: Markdown content (file name + encrypted content).
- **Diagrams**:
  - Hierarchical structure (Parent -> Child).
  - Stored flat in MongoDB with `parent_diagram_id` reference.
  - Nodes and edges are **E2E encrypted on frontend** and stored as encrypted strings.
- **Node Vaults**: Sensitive data storage (Type + Encrypted Value).

### Non-Functional

- **Security**:
  - E2E Encryption.
  - Server stores only: Salt, Public Key, Encrypted Private Key per project.
  - Frontend handles encryption/decryption.
  - Diagram nodes/edges are encrypted before sending to backend.
- **Database**: MongoDB (Scalable, Flexible schema).
- **Architecture**: Clean, scalable, maintainable Go project structure.
- **Logging**: Structured logging with best practices:
  - Use structured logger (e.g., `zerolog` or `zap`).
  - Log levels: DEBUG, INFO, WARN, ERROR.
  - Include context: request ID, user ID, trace information.
  - Log important business events and errors.
  - Do NOT log sensitive data (passwords, tokens, encryption keys).

## Technical Architecture

### 1. Project Structure (Standard Go Layout)

```text
/
├── cmd/
│   └── server/
│       └── main.go       # Entry point
├── internal/
│   ├── core/             # Core Domain Logic (Hexagonal/Clean Arch)
│   │   ├── domain/       # Enterprise Business Rules (Models)
│   │   ├── port/         # Interfaces (Repositories, Services)
│   │   └── service/      # Application Business Rules
│   ├── adapter/          # adapters (implementations)
│   │   ├── handler/      # HTTP Handlers (Gin)
│   │   ├── repository/   # Database Access (MongoDB/mgod)
│   │   ├── auth/         # Authentication logic
│   │   └── dto/          # Data Transfer Objects
│   └── config/           # Configuration loading
├── pkg/                  # Public shared code (utils, etc.)
│   └── validation/       # Custom validation engine
└── go.mod
```

### 2. Database Layer (MongoDB)

- **Driver**: Official `go.mongodb.org/mongo-driver`.
- **ODM**: `github.com/Lyearn/mgod`.
- **Connection**: Managed via `mgod.ConfigureDefaultClient` or `mgod.SetDefaultClient`.

### 3. Models with mgod

#### Important: mgod Usage Pattern

**mgod does NOT have a `DefaultModel` base struct.** Models are plain Go structs with `bson` tags. When you create an `EntityMongoModel`, mgod automatically adds:

- `_id` (ObjectID)
- `__v` (version number)
- Timestamps (if configured)

Example from mgod documentation:

```go
type User struct {
    Name     string `bson:"name" json:"name"`
    EmailID  string `bson:"emailId" json:"email_id"`
    Age      *int32 `bson:",omitempty" json:"age,omitempty"`
}

// Usage:
model := User{}
opts := mgod.NewEntityMongoModelOptions(dbName, collection, nil)
userModel, _ := mgod.NewEntityMongoModel(model, *opts)
```

#### Flexible Permission System

To support flexible permissions, we use a granular permission system in the `ProjectMember` model.

**Permission Constants:**

```go
const (
    PermissionViewDiagram   = "view_diagram"
    PermissionEditDiagram   = "edit_diagram"
    PermissionViewNote      = "view_note"
    PermissionEditNote      = "edit_note"
    PermissionViewVault     = "view_vault"
    PermissionEditVault     = "edit_vault"
    PermissionManageProject = "manage_project"
)
```

#### Model Schemas

**User:**

```go
type User struct {
    Name     string `bson:"name" json:"name"`
    Username string `bson:"username" json:"username"`
    Password string `bson:"password" json:"password"`
    Email    string `bson:"email" json:"email"`
}
```

**Project:**

```go
type Project struct {
    Name                string `bson:"name" json:"name"`
    Description         string `bson:"description" json:"description"`
    EncryptionSalt      string `bson:"encryption_salt" json:"encryption_salt"`
    EncryptedPrivateKey string `bson:"encrypted_private_key" json:"encrypted_private_key"`
    EncryptionPublicKey string `bson:"encryption_public_key" json:"encryption_public_key"`
}
```

**ProjectMember (with Flexible Permissions):**

```go
type ProjectMember struct {
    ProjectID   primitive.ObjectID `bson:"project_id" json:"project_id"`
    UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
    Permissions []string           `bson:"permissions" json:"permissions"`
    Role        string             `bson:"role" json:"role"` // Optional preset name
}
```

**Note:**

```go
type Note struct {
    ProjectID        primitive.ObjectID `bson:"project_id" json:"project_id"`
    FileName         string             `bson:"file_name" json:"file_name"`
    FileType         string             `bson:"file_type" json:"file_type"`
    EncryptedContent *string            `bson:"encrypted_content,omitempty" json:"encrypted_content,omitempty"`
}
```

**Diagram (with E2E Encrypted Nodes/Edges):**

```go
type Diagram struct {
    ProjectID       primitive.ObjectID  `bson:"project_id" json:"project_id"`
    ParentDiagramID *primitive.ObjectID `bson:"parent_diagram_id,omitempty" json:"parent_diagram_id,omitempty"`
    Name            string              `bson:"name" json:"name"`
    Description     string              `bson:"description" json:"description"`
    EncryptedNodes  string              `bson:"encrypted_nodes" json:"encrypted_nodes"`
    EncryptedEdges  string              `bson:"encrypted_edges" json:"encrypted_edges"`
}
```

**Note:** Nodes and edges are encrypted on the frontend and stored as encrypted strings for E2E encryption.

**NodeVault:**

```go
type NodeVault struct {
    NodeId         primitive.ObjectID `bson:"node_id" json:"node_id"`
    Type           string             `bson:"type" json:"type"`
    EncryptedValue string             `bson:"encrypted_value" json:"encrypted_value"`
}
```

### 4. API Schema & Validation

- **Framework**: [Gin](https://github.com/gin-gonic/gin)
- **Response Format**: Standardized JSON (Success/Error wrappers).
- **Validation Engine**:
  - Use `go-playground/validator` (v10).
  - Custom validation engine in `pkg/validation` maps `validator.FieldError` to `[]map[string]string` format.
  - Easy integration with `APIResponse.Error.Fields`.

**APIResponse Structure:**

```go
type APIResponse[T any] struct {
    Data       T                   `json:"data"`
    Meta       *MetadataResponse   `json:"meta"`
    Error      *ErrorResponse      `json:"error,omitempty"`
    Pagination *PaginationResponse `json:"pagination,omitempty"`
}

type ErrorResponse struct {
    Code    string               `json:"code"`    // Application error code (e.g., "USER_ALREADY_EXISTS")
    Message string               `json:"message"` // Human-readable error message
    Fields  *[]map[string]string `json:"fields,omitempty"` // Validation errors
}
```

### 5. Error Code Dictionary

We use **application-specific error codes** (strings) separate from HTTP status codes for better error handling on the frontend.

**Error Code Constants** (`internal/adapter/dto/error_codes.go`):

```go
const (
    // Authentication errors
    ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
    ErrCodeUserAlreadyExists  = "USER_ALREADY_EXISTS"
    ErrCodeInvalidToken       = "INVALID_TOKEN"
    ErrCodeUnauthorized       = "UNAUTHORIZED"

    // Validation errors
    ErrCodeValidationFailed = "VALIDATION_FAILED"
    ErrCodeInvalidRequest   = "INVALID_REQUEST"

    // Resource errors
    ErrCodeNotFound      = "RESOURCE_NOT_FOUND"
    ErrCodeAlreadyExists = "RESOURCE_ALREADY_EXISTS"
    ErrCodeForbidden     = "FORBIDDEN"

    // Server errors
    ErrCodeInternalError = "INTERNAL_SERVER_ERROR"
    ErrCodeDatabaseError = "DATABASE_ERROR"
)
```

**Helper Functions:**

```go
// Create error response with code from dictionary
dto.NewErrorResponse(dto.ErrCodeInvalidCredentials)

// Create error response with custom message
dto.NewErrorResponse(dto.ErrCodeUnauthorized, "Authorization header required")

// Create validation error response
dto.NewValidationErrorResponse(validationErrors)

// Success response
dto.NewAPIResponse(data, nil)

// Error response
dto.NewAPIResponse[any](nil, dto.NewErrorResponse(dto.ErrCodeInternalError))
```

**Example API Responses:**

Success:

```json
{
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "abc123...",
    "expires_in": 900
  },
  "meta": {
    "request_id": "",
    "timestamp": "2026-02-08T19:20:00+08:00"
  }
}
```

Error:

```json
{
  "data": null,
  "meta": {
    "request_id": "",
    "timestamp": "2026-02-08T19:20:00+08:00"
  },
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid email/username or password"
  }
}
```

Validation Error:

```json
{
  "data": null,
  "meta": {...},
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Validation failed",
    "fields": [
      {"email": "Invalid email format"},
      {"password": "Minimum length is 8"}
    ]
  }
}
```

## Implementation Roadmap

### Phase 1: Foundation & Structure ✅

1.  ✅ **Structure**: Created `cmd`, `internal`, `pkg` directories.
2.  ✅ **Dependencies**: Updated `go.mod` with `mgod` and `mongo-driver`.
3.  ✅ **Validation**: Implemented `pkg/validation/validator.go`.

### Phase 2: Domain Models (Core) ✅

1.  ✅ **Migrate Models**: Implemented `internal/core/domain/*.go` with proper mgod pattern.
2.  ✅ **Permission System**: Implemented flexible permission system in `ProjectMember`.

### Phase 3: Business Logic (Next)

1.  **Interfaces**: Define repository interfaces in `internal/core/port`.
2.  **Database Connection**: Implement MongoDB connection initialization.
3.  **Repositories**: Implement repository layer using mgod.
4.  **Services**: Implement business logic in `internal/core/service`.
5.  **Handlers**: Update Gin handlers to use Services and Validation Engine.
6.  **Router**: Wire everything in `cmd/server/main.go`.
