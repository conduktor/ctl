# Environment Variables Configuration

Before using the CLI, you need to configure environment variables depending on which Conduktor services you're interacting with. The CLI supports both Console API and Gateway API operations.

### Console API Configuration

These variables are required for managing Console resources (topics, users, groups, applications, etc.):

#### Connection
- **CDK_BASE_URL**: Base URL of your Conduktor Console instance (required)

#### Authentication (choose one method)
**Method 1: API Key (Recommended)**
- **CDK_API_KEY**: API key generated from Console

**Method 2: Username/Password**
- **CDK_USER**: Username for authentication
- **CDK_PASSWORD**: Password for authentication

#### Additional Console Client Options
- **CDK_AUTH_MODE**: Authentication mode (`conduktor` or `external`, default: `conduktor`)

#### Example Console Configuration
```bash
# Option 1: API Key authentication
export CDK_BASE_URL="https://console.conduktor.example.com"
export CDK_API_KEY="your-console-api-key"

# Option 2: Username/Password authentication
export CDK_BASE_URL="https://console.conduktor.example.com"
export CDK_USER="your-username"
export CDK_PASSWORD="your-password"
```

### Gateway API Configuration

These variables are required for managing Gateway resources (interceptors, virtual clusters, etc.):

#### Connection & Authentication
- **CDK_GATEWAY_BASE_URL**: Base URL of your Conduktor Gateway instance (required)
- **CDK_GATEWAY_USER**: Gateway username (required)
- **CDK_GATEWAY_PASSWORD**: Gateway password (required)

#### Example Gateway Configuration
```bash
export CDK_GATEWAY_BASE_URL="https://gateway.conduktor.example.com"
export CDK_GATEWAY_USER="gateway-admin"
export CDK_GATEWAY_PASSWORD="gateway-password"
```

### Dual Environment Setup

For environments using both Console and Gateway, configure all relevant variables:

```bash
# Console API
export CDK_BASE_URL="https://console.conduktor.example.com"
export CDK_API_KEY="your-console-api-key"

# Gateway API
export CDK_GATEWAY_BASE_URL="https://gateway.conduktor.example.com"
export CDK_GATEWAY_USER="gateway-admin"
export CDK_GATEWAY_PASSWORD="gateway-password"

# Optional: TLS settings (apply to both)
export CDK_INSECURE="false"
export CDK_CACERT="/path/to/ca.crt"
```

### Additional Console/Gateway Client Options
- **CDK_INSECURE**: Set to `true` to ignore server TLS certificate verification
- **CDK_CACERT**: Path to certificate authority file for server TLS verification
- **CDK_KEY**: Path to client private key file (if backend is behhind a TLS authentication based proxy like Teleport)
- **CDK_CERT**: Path to client certificate file  (if backend is behhind a TLS authentication based proxy like Teleport)

