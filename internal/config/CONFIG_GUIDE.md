# Configuration Guide

This guide explains the configuration options available in Infrantery backend and how to set them up properly for different deployment scenarios.

## Configuration Structure

The application configuration is loaded from environment variables with sensible defaults. All configuration options are defined in `config.go` and can be customized via a `.env` file or system environment variables.

## Configuration Options

### Server Settings

#### `PORT`

- **Description**: The port on which the HTTP server listens
- **Default**: `8085`
- **Example**: `PORT=8080`

### Database Settings

#### `MONGODB_URI`

- **Description**: MongoDB connection string
- **Default**: `mongodb://localhost:27017`
- **Example**: `MONGODB_URI=mongodb://user:password@mongodb.example.com:27017`

#### `MONGODB_DATABASE`

- **Description**: Name of the MongoDB database to use
- **Default**: `infrantery`
- **Example**: `MONGODB_DATABASE=infrantery_prod`

### JWT (JSON Web Token) Settings

#### `JWT_SECRET`

- **Description**: Secret key used to sign JWT tokens. **MUST be changed in production!**
- **Default**: `your-super-secret-key`
- **Security**: Use a strong, random string (at least 32 characters)
- **Example**: `JWT_SECRET=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6`

#### `JWT_ACCESS_EXPIRY`

- **Description**: Expiration time for access tokens
- **Default**: `15m` (15 minutes)
- **Format**: Use Go duration format (e.g., `15m`, `1h`, `2h30m`)
- **Example**: `JWT_ACCESS_EXPIRY=30m`

#### `JWT_REFRESH_EXPIRY`

- **Description**: Expiration time for refresh tokens
- **Default**: `168h` (7 days)
- **Format**: Use Go duration format
- **Example**: `JWT_REFRESH_EXPIRY=720h` (30 days)

### Password Hashing (Argon2) Settings

#### `ARGON2_MEMORY`

- **Description**: Memory cost parameter for Argon2 (in KiB)
- **Default**: `65536` (64 MB)
- **Example**: `ARGON2_MEMORY=131072` (128 MB)

#### `ARGON2_ITERATIONS`

- **Description**: Number of iterations for Argon2
- **Default**: `3`
- **Example**: `ARGON2_ITERATIONS=4`

#### `ARGON2_PARALLELISM`

- **Description**: Degree of parallelism for Argon2
- **Default**: `2`
- **Example**: `ARGON2_PARALLELISM=4`

#### `ARGON2_SALT_LENGTH`

- **Description**: Length of the salt in bytes
- **Default**: `16`
- **Example**: `ARGON2_SALT_LENGTH=16`

#### `ARGON2_KEY_LENGTH`

- **Description**: Length of the generated key in bytes
- **Default**: `32`
- **Example**: `ARGON2_KEY_LENGTH=32`

### Logging Settings

#### `LOG_LEVEL`

- **Description**: Logging verbosity level
- **Default**: `info`
- **Allowed Values**: `debug`, `info`, `warn`, `error`
- **Example**: `LOG_LEVEL=debug`

#### `ENVIRONMENT`

- **Description**: Current environment
- **Default**: `development`
- **Allowed Values**: `development`, `production`, `staging`
- **Example**: `ENVIRONMENT=production`

---

## Cookie Settings (Critical for Cross-Origin Setup)

Cookie settings are crucial for proper authentication when your frontend and backend are deployed on different domains or subdomains.

### `COOKIE_DOMAIN`

**Description**: Specifies the domain for which cookies are valid. This is the most critical setting for cross-origin authentication.

**Default**: `localhost`

**How It Works**:

- When set, cookies will be sent to the specified domain and all its subdomains
- When empty, cookies are restricted to the exact host that set them
- Must match the domain structure of your deployment

#### Deployment Scenarios & Examples

##### Scenario 1: Development (Same Domain)

**Setup**: Both frontend and backend on localhost with different ports

- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8085`

**Configuration**:

```bash
COOKIE_DOMAIN=localhost
COOKIE_SECURE=false
COOKIE_SAMESITE=lax
```

**Explanation**:

- `localhost` allows cookies to work across different ports on localhost
- `SECURE=false` because HTTP (not HTTPS) is used in development
- `SAMESITE=lax` allows cookies to be sent with top-level navigations

---

##### Scenario 2: Different Subdomains (Same Top-Level Domain)

**Setup**: Frontend and backend on different subdomains

- Frontend: `https://app.example.com`
- Backend: `https://api.example.com`

**Configuration**:

```bash
COOKIE_DOMAIN=.example.com
COOKIE_SECURE=true
COOKIE_SAMESITE=none
```

**Explanation**:

- `.example.com` (note the leading dot) makes cookies available to all subdomains
- `SECURE=true` required because using HTTPS
- `SAMESITE=none` required for cross-subdomain cookies (must have `SECURE=true`)

**Important**: The leading dot (`.`) is crucial! Without it, the cookie won't work across subdomains.

---

##### Scenario 3: Completely Different Domains

**Setup**: Frontend and backend on entirely different domains

- Frontend: `https://myapp.com`
- Backend: `https://api.backend.io`

**Configuration**:

```bash
COOKIE_DOMAIN=api.backend.io
COOKIE_SECURE=true
COOKIE_SAMESITE=none
```

**Additional Requirements**:

1. **Backend CORS Configuration**: Must explicitly allow the frontend domain
2. **Frontend API Client**: Must send credentials with requests (`credentials: 'include'`)

**Explanation**:

- Cookie domain is set to backend's domain
- `SAMESITE=none` with `SECURE=true` allows cross-origin cookie sharing
- Browser will only send cookies if frontend explicitly includes credentials

**Frontend Configuration (apiClient.ts)**:

```typescript
const response = await fetch(url, {
  credentials: "include", // Critical for cross-origin cookies
  // ... other options
});
```

---

##### Scenario 4: Same Domain, Different Paths

**Setup**: Frontend and backend served from same domain

- Frontend: `https://example.com/`
- Backend: `https://example.com/api`

**Configuration**:

```bash
COOKIE_DOMAIN=example.com
COOKIE_SECURE=true
COOKIE_SAMESITE=strict
```

**Explanation**:

- Simple configuration since both are on the same domain
- `SAMESITE=strict` provides maximum security
- No special cross-origin handling needed

---

##### Scenario 5: Production with CDN

**Setup**: Frontend on CDN, backend on own domain

- Frontend: `https://cdn.cloudfront.net/myapp`
- Backend: `https://api.myapp.com`

**Configuration**:

```bash
COOKIE_DOMAIN=api.myapp.com
COOKIE_SECURE=true
COOKIE_SAMESITE=none
```

**Additional Setup**:

- Configure CDN to allow credentials
- Ensure CORS headers are properly set
- Consider using custom domain for CDN (e.g., `app.myapp.com`)

---

### `COOKIE_SECURE`

**Description**: Determines if cookies should only be sent over HTTPS

**Default**: `false`

**Values**:

- `true`: Cookies only sent over HTTPS (required for production)
- `false`: Cookies sent over HTTP (only for development)

**Examples**:

```bash
# Development
COOKIE_SECURE=false

# Production
COOKIE_SECURE=true
```

**Security Note**: Always use `true` in production. Modern browsers require `SECURE=true` when `SAMESITE=none`.

---

### `COOKIE_SAMESITE`

**Description**: Controls when cookies are sent with cross-site requests

**Default**: `lax`

**Allowed Values**:

- `strict`: Cookies only sent with same-site requests (most secure)
- `lax`: Cookies sent with top-level navigations (balanced)
- `none`: Cookies sent with all requests (requires `SECURE=true`)

**When to Use Each**:

| Value    | Use Case                     | Requirements                  |
| -------- | ---------------------------- | ----------------------------- |
| `strict` | Same domain deployment       | -                             |
| `lax`    | Development, same domain     | -                             |
| `none`   | Different domains/subdomains | Must set `COOKIE_SECURE=true` |

**Examples**:

```bash
# Same domain
COOKIE_SAMESITE=strict

# Cross-subdomain
COOKIE_SAMESITE=none

# Development
COOKIE_SAMESITE=lax
```

---

## Complete Configuration Examples

### Local Development

```bash
# .env
PORT=8085
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=infrantery_dev
JWT_SECRET=dev-secret-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
LOG_LEVEL=debug
ENVIRONMENT=development

# Cookie settings for localhost
COOKIE_DOMAIN=localhost
COOKIE_SECURE=false
COOKIE_SAMESITE=lax
```

### Production (Same Domain)

```bash
# .env
PORT=8080
MONGODB_URI=mongodb://user:password@mongodb.example.com:27017
MONGODB_DATABASE=infrantery_prod
JWT_SECRET=your-very-long-and-random-production-secret-key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=720h
LOG_LEVEL=info
ENVIRONMENT=production

# Cookie settings
COOKIE_DOMAIN=example.com
COOKIE_SECURE=true
COOKIE_SAMESITE=strict
```

### Production (Subdomain Architecture)

```bash
# .env
PORT=8080
MONGODB_URI=mongodb://user:password@mongodb.example.com:27017
MONGODB_DATABASE=infrantery_prod
JWT_SECRET=your-very-long-and-random-production-secret-key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=720h
LOG_LEVEL=info
ENVIRONMENT=production

# Cookie settings for subdomain (note the leading dot!)
COOKIE_DOMAIN=.example.com
COOKIE_SECURE=true
COOKIE_SAMESITE=none
```

### Production (Different Domains)

```bash
# .env
PORT=8080
MONGODB_URI=mongodb://user:password@mongodb.backend.io:27017
MONGODB_DATABASE=infrantery_prod
JWT_SECRET=your-very-long-and-random-production-secret-key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=720h
LOG_LEVEL=info
ENVIRONMENT=production

# Cookie settings for cross-origin
COOKIE_DOMAIN=api.backend.io
COOKIE_SECURE=true
COOKIE_SAMESITE=none
```

---

## Troubleshooting Cookie Issues

### Cookies Not Being Set

1. **Check Cookie Domain**:
   - Ensure `COOKIE_DOMAIN` matches your deployment structure
   - For subdomains, use leading dot (`.example.com`)
   - For different domains, set to backend domain

2. **Verify HTTPS/Secure Settings**:
   - If using `SAMESITE=none`, you MUST set `SECURE=true`
   - Production should always use `SECURE=true`

3. **Check Browser Console**:
   - Look for cookie warnings in browser DevTools
   - Check if cookies are being blocked by browser settings

### Cookies Not Being Sent

1. **Frontend Configuration**:
   - Ensure `credentials: 'include'` in fetch requests
   - Check CORS configuration allows credentials

2. **SameSite Issues**:
   - Cross-origin requires `SAMESITE=none` and `SECURE=true`
   - Same domain can use `strict` or `lax`

3. **Domain Mismatch**:
   - Cookie domain must align with your architecture
   - Check browser DevTools Application tab to see cookie domain

---

## Security Best Practices

1. **Always use strong JWT secrets** (32+ random characters)
2. **Enable COOKIE_SECURE in production** (requires HTTPS)
3. **Use SAMESITE=strict** when possible for better security
4. **Rotate JWT secrets periodically**
5. **Use environment-specific configurations** (never commit production secrets)
6. **Keep JWT expiry times reasonable** (short access tokens, longer refresh tokens)
7. **Monitor and log authentication attempts**
8. **Use HTTPS in production** (required for secure cookies)

---

## Additional Resources

- [MDN: SameSite Cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite)
- [MDN: Secure Cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#restrict_access_to_cookies)
- [OWASP: Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
