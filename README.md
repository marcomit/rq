# rq

> A simple, powerful CLI for API testing. Raw HTTP syntax with superpowers.

**rq** lets you test APIs using the actual HTTP protocol syntax you already know, enhanced with variables, workspaces (docks), and multi-protocol support. No abstractions, no new syntax to learn—just HTTP with superpowers.

## Why rq?

- **Zero learning curve** - Use raw HTTP syntax, not custom formats
- **Project integration** - Keep requests alongside your code
- **Version control friendly** - Everything is plain text files
- **Multi-protocol ready** - HTTP today, WebSocket/gRPC tomorrow
- **Workspace inheritance** - Configure once, use everywhere
- **Simple as possible** - But still incredibly powerful

## Quick Start

```bash
# Create a dock (workspace)
rq dock init my-api

# Create your first request
rq new login

# Edit the generated login.http file
GET {{BASE_URL}}/auth/login HTTP/1.1
Content-Type: application/json

{"username": "{{USER}}", "password": "{{PASS}}"}

# Configure environment
echo "BASE_URL=api.example.com" > env
echo "USER=admin" >> env  
echo "PASS=secret123" >> env

# Run the request
rq run login
```

## Installation

### Build from Source
```bash
git clone https://github.com/marcomit/rq.git
cd rq
go build -o rq
```

## Core Concepts

### Docks (Workspaces)
A dock is a workspace containing your API requests and configuration. Think of it like a Git repository for API testing.

```bash
my-project/
├── api-dock/           # Your dock
│   ├── .dock          # Dock marker file
│   ├── env            # Environment config
│   ├── login.http     # HTTP request
│   └── users.http     # Another request
```

### Requests
Requests are written in standard HTTP syntax with variable substitution:

```http
POST {{BASE_URL}}/api/users HTTP/1.1
Authorization: Bearer {{JWT_TOKEN}}
Content-Type: application/json

{
  "name": "{{USER_NAME}}",
  "avatar": "{{file(avatar.jpg)}}"
}
```

### Environment Configuration
Simple key-value configuration files:

```bash
# env file
BASE_URL=api.example.com
JWT_TOKEN=eyJ0eXAiOiJKV1QiLCJhbGci...
USER_NAME=john_doe
```

### Subdocks (Inherited Configuration)
Organize related requests with inherited configuration:

```bash
my-dock/
├── env                 # Global config
├── login.http
└── auth/              # Subdock
    ├── env            # Auth-specific config (inherits from parent)
    ├── signin.http
    └── oauth/         # Nested subdock
        ├── env        # OAuth config (inherits from auth + global)
        └── token.http
```

## Commands

### Dock Management
```bash
rq dock init <name>     # Create new dock
rq dock list            # List available docks  
rq dock use <name>      # Switch to dock
rq dock status          # Show current dock info
```

### Request Management
```bash
rq new <name>           # Create HTTP request
rq new <path/name>      # Create request in subdock
rq new chat --type ws   # Create WebSocket request (future)

rq run <name>           # Run request
rq run <name> --env dev # Run with specific environment
rq run <name> -o out.json # Save output to file
```

### Environment Management
```bash
rq env list             # Show available environments
rq env set dev          # Switch to dev environment  
rq env edit <path>      # Edit environment config
rq env show <path>      # Show effective config
rq env tree             # Show config inheritance
```

## Variable System

### Simple Variables
```http
GET {{BASE_URL}}/users HTTP/1.1
Authorization: {{AUTH_HEADER}}
```

### Function Variables
```http
POST {{BASE_URL}}/upload HTTP/1.1
Content-Type: multipart/form-data

{
  "file": "{{file(document.pdf)}}",
  "checksum": "{{sha256(document.pdf)}}",
  "id": "{{uuid()}}"
}
```

Available functions:
- `{{file(path)}}` - Read file as base64
- `{{sha256(text)}}` - SHA256 hash
- `{{uuid()}}` - Generate UUID
- `{{timestamp()}}` - Current timestamp
- More coming soon...

## File Structure

### Basic Dock
```
my-api/
├── .dock              # Dock marker
├── env                # Default environment
├── env.dev            # Development environment
├── env.prod           # Production environment
├── login.http         # Login request
└── users.http         # Users request
```

### Complex Dock with Subdocks
```
enterprise-api/
├── .dock
├── env                # Global: BASE_URL=api.company.com
├── health.http
├── auth/              # Authentication subdock
│   ├── env            # Auth: AUTH_URL=auth.company.com
│   ├── login.http
│   ├── refresh.http
│   └── oauth/         # OAuth subdock
│       ├── env        # OAuth: CLIENT_ID=abc123
│       └── token.http
├── users/             # Users subdock  
│   ├── env            # Users: API_VERSION=v2
│   ├── list.http
│   └── create.http
└── orders/
    ├── list.http
    └── create.http
```

## Output Options

```bash
# Save response
rq run users -o response.json

# Save only response body
rq run users --output-body data.json

# Save full request + response
rq run users --output-full exchange.json

# Append to file
rq run users -o log.json --append
```

## Examples

### REST API Testing
```http
# users.http
GET {{BASE_URL}}/api/v1/users?limit={{LIMIT}} HTTP/1.1
Authorization: Bearer {{JWT_TOKEN}}
Accept: application/json
```

### File Upload
```http
# upload.http
POST {{BASE_URL}}/api/v1/files HTTP/1.1
Authorization: Bearer {{JWT_TOKEN}}
Content-Type: multipart/form-data

{
  "name": "{{FILENAME}}",
  "data": "{{file(./uploads/document.pdf)}}",
  "hash": "{{sha256(./uploads/document.pdf)}}"
}
```

### Environment-Specific Requests
```bash
# env
BASE_URL=http://localhost:3000
JWT_TOKEN=dev_token_123

# env.prod  
BASE_URL=https://api.production.com
JWT_TOKEN=prod_token_456
```

```bash
rq run users              # Uses default env
rq run users --env prod   # Uses production env
```

## Philosophy

**rq** follows these principles:

1. **Use existing standards** - HTTP syntax, not custom formats
2. **Simple over complex** - Powerful features with simple interfaces  
3. **Files over databases** - Everything is version-controllable text
4. **Project integration** - API tests live with your code
5. **Protocol agnostic** - HTTP today, anything tomorrow

## Roadmap

- [x] HTTP requests with variables
- [x] Dock (workspace) management
- [x] Environment inheritance
- [x] File functions and output saving
- [ ] WebSocket support (`.ws` files)
- [ ] gRPC support (`.grpc` files)
- [ ] Request templates
- [ ] Testing assertions
- [ ] CI/CD integration helpers
- [ ] Plugin system

## Contributing

rq is built with Go and focuses on simplicity and developer experience. Contributions welcome!

```bash
git clone https://github.com/yourname/rq.git
cd rq
go run main.go dock init test
go run main.go new hello
go run main.go run hello
```

## License

MIT License - see [LICENSE](LICENSE) file.

---

**rq** - Because testing APIs should be as simple as writing HTTP.
