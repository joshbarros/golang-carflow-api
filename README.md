# 🚗 CarFlow - B2B SaaS for Car Dealerships

**A modern, cloud-ready fleet management platform built for car dealerships, rental companies, and fleet managers.**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://www.docker.com/)

---

## 📋 Overview

CarFlow is a **multi-tenant B2B SaaS platform** that enables car dealerships to manage their entire inventory in the cloud. Built with Go and PostgreSQL, it features JWT authentication, role-based access control, subscription management, and a comprehensive API.

### 🎯 **Built For:**
- 🏢 Independent car dealerships
- 🚗 Car rental companies
- 🚚 Fleet management companies
- 🏭 Wholesale auto auctions

---

## 🌟 Features

### Core Functionality
- ✅ **Multi-tenant Architecture** - Complete data isolation per dealership
- ✅ **Advanced Vehicle Management** - VIN, stock numbers, pricing, mileage, status tracking
- ✅ **JWT Authentication** - Secure, stateless authentication
- ✅ **Role-Based Access Control** - Owner, Admin, Member roles
- ✅ **PostgreSQL Database** - Production-ready persistence with migrations
- ✅ **RESTful API** - Clean, documented JSON API
- ✅ **Filtering & Pagination** - Advanced search capabilities
- ✅ **Audit Logs** - Complete change history tracking

### SaaS Features
- 🔐 **API Key Management** - Programmatic access for integrations
- 💳 **Subscription Management** - Stripe-ready billing system
- 📊 **Usage Tracking** - Monitor API calls and resource usage
- 📧 **Email Integration** - Brevo-ready for transactional emails
- 📈 **Analytics Dashboard** - Built-in metrics and monitoring
- 🔒 **Security** - Password hashing, rate limiting, CORS

### Developer Experience
- 🐳 **Docker Compose** - One-command development setup
- 📝 **Database Migrations** - Automated schema management
- 🧪 **Comprehensive Tests** - Unit and integration testing
- 📖 **OpenAPI Documentation** - Interactive API docs
- 🛠️ **CLI Tools** - Command-line interface included
- 🎨 **Web UI** - Built-in management interface

---

## 🏗️ Architecture

```
CarFlow SaaS (NX Monorepo)
├── apps/
│   ├── api/                    # Go Backend
│   │   ├── cmd/
│   │   │   ├── server/         # Main API server
│   │   │   ├── migrate/        # Database migrations
│   │   │   ├── cli/            # CLI tool
│   │   │   └── ui/             # Web UI
│   │   ├── internal/
│   │   │   ├── auth/           # JWT & password hashing
│   │   │   ├── tenant/         # Multi-tenancy
│   │   │   ├── user/           # User management
│   │   │   ├── car/            # Vehicle management
│   │   │   └── database/       # DB connection
│   │   └── migrations/         # SQL migrations
│   ├── web/                    # React dashboard (Coming soon)
│   └── admin/                  # Admin dashboard (Coming soon)
├── libs/                       # Shared libraries
├── infra/
│   └── docker/                 # Dockerfiles
└── scripts/                    # Utility scripts
```

---

## 🚀 Quick Start

### Prerequisites

- **Docker** & **Docker Compose** (required)
- **Go 1.22+** (for local development)
- **Node.js 18+** (for frontend development)

### 🎯 **One-Command Setup**

```bash
# Clone the repository
git clone https://github.com/joshbarros/golang-carflow-api.git
cd golang-carflow-api

# Start everything (PostgreSQL + Redis + API)
./scripts/start-dev.sh
```

**That's it!** 🎉

The script will:
1. ✅ Create `.env` from `.env.example`
2. ✅ Start PostgreSQL, Redis, and the API
3. ✅ Run database migrations
4. ✅ Seed demo data

### 📍 **Access Points**

Once running, access:

| Service | URL | Credentials |
|---------|-----|-------------|
| **API** | http://localhost:8080 | - |
| **API Docs** | http://localhost:8080/api-docs | - |
| **Health Check** | http://localhost:8080/healthz | - |
| **Metrics** | http://localhost:8080/metrics | - |
| **PgAdmin** | http://localhost:5050 | admin@carflow.local / admin |
| **Redis Commander** | http://localhost:8081 | - |
| **MailHog** | http://localhost:8025 | - |

### 🧪 **Demo Credentials**

Test the API with pre-seeded data:

```
Tenant:   demo-dealership
Email:    owner@demo-dealership.com
Password: password123
```

---

## 🐳 Docker Commands

### Development

```bash
# Start with dev tools (PgAdmin, MailHog, etc.)
./scripts/start-dev.sh

# View logs
docker-compose -f docker-compose.yml -f docker-compose.dev.yml logs -f

# Stop everything
./scripts/stop.sh
```

### Production

```bash
# Start production stack
./scripts/start.sh

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Database Management

```bash
# Run migrations
./scripts/db.sh migrate

# Rollback last migration
./scripts/db.sh migrate-down

# Reset database (WARNING: deletes all data)
./scripts/db.sh reset

# Seed demo data
./scripts/db.sh seed

# Open PostgreSQL shell
./scripts/db.sh shell

# Create backup
./scripts/db.sh backup

# Restore from backup
./scripts/db.sh restore backups/carflow_20231123_120000.sql
```

---

## 🔧 Local Development (Without Docker)

### 1. Install PostgreSQL

```bash
# macOS
brew install postgresql@15
brew services start postgresql@15

# Ubuntu/Debian
sudo apt install postgresql-15
sudo systemctl start postgresql
```

### 2. Create Database

```bash
createdb carflow
```

### 3. Configure Environment

```bash
cp .env.example .env
# Edit .env with your local database credentials
```

### 4. Run Migrations

```bash
cd apps/api
go run ./cmd/migrate -database "postgres://carflow:carflow123@localhost:5432/carflow?sslmode=disable" -path ./migrations
```

### 5. Start API Server

```bash
cd apps/api
go run ./cmd/server
```

---

## 📊 Database Schema

CarFlow uses a **multi-tenant PostgreSQL** database with the following tables:

| Table | Description |
|-------|-------------|
| `tenants` | Dealership/company accounts |
| `users` | User accounts with RBAC |
| `api_keys` | API keys for integrations |
| `cars` | Vehicle inventory (tenant-isolated) |
| `subscriptions` | Stripe subscription data |
| `usage_tracking` | API usage & billing metrics |
| `audit_logs` | Complete change history |

**Key Features:**
- ✅ UUID primary keys
- ✅ Soft deletes
- ✅ Automatic timestamps
- ✅ JSONB for flexible metadata
- ✅ Comprehensive indexing

---

## 🔐 Authentication

CarFlow uses **JWT (JSON Web Tokens)** for authentication.

### Register New Tenant

```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "dealership_name": "ABC Motors",
  "owner_email": "owner@abcmotors.com",
  "owner_password": "securepassword123",
  "owner_first_name": "John",
  "owner_last_name": "Doe"
}
```

### Login

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "owner@demo-dealership.com",
  "password": "password123"
}

# Response:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {...}
}
```

### Use Token

```bash
GET /api/v1/cars
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

---

## 🚗 API Examples

### List All Vehicles

```bash
curl -X GET "http://localhost:8080/api/v1/cars" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Get Vehicle Details

```bash
curl -X GET "http://localhost:8080/api/v1/cars/car-001" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Create New Vehicle

```bash
curl -X POST "http://localhost:8080/api/v1/cars" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "car-new-001",
    "make": "Toyota",
    "model": "Camry",
    "year": 2024,
    "color": "Silver",
    "vin": "1HGCM82633A789012",
    "stock_number": "STK-1234",
    "price": 32999.00,
    "mileage": 0,
    "status": "available",
    "description": "Brand new 2024 Toyota Camry"
  }'
```

### Filter & Search

```bash
# Filter by make and status
curl "http://localhost:8080/api/v1/cars?make=Toyota&status=available"

# Pagination
curl "http://localhost:8080/api/v1/cars?page=1&page_size=10"

# Sort
curl "http://localhost:8080/api/v1/cars?sort_by=price&sort_order=desc"
```

---

## 💳 Stripe Subscription Management

CarFlow integrates with **Stripe Checkout** for subscription billing and payment processing.

### 🔑 Setup Stripe Integration

1. **Create a Stripe Account**
   - Sign up at [stripe.com](https://stripe.com)
   - Get your API keys from the Stripe Dashboard

2. **Configure Environment Variables**

```bash
# Add to your .env file
STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key
STRIPE_PUBLISHABLE_KEY=pk_test_your_stripe_publishable_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret

# Create products in Stripe and add price IDs
STRIPE_PRICE_STARTER=price_starter_id
STRIPE_PRICE_PROFESSIONAL=price_professional_id
STRIPE_PRICE_ENTERPRISE=price_enterprise_id
```

3. **Create Stripe Products**

In your Stripe Dashboard, create three products:
- **Starter**: $29/month (100 vehicles, 3 users)
- **Professional**: $99/month (500 vehicles, 10 users)
- **Enterprise**: $299/month (unlimited)

Copy the price IDs and add them to your `.env` file.

### 🛒 Subscription Endpoints

#### Create Checkout Session

```bash
POST /api/v1/subscriptions/checkout
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "plan": "professional",
  "success_url": "http://localhost:3000/billing/success",
  "cancel_url": "http://localhost:3000/billing"
}

# Response:
{
  "checkout_url": "https://checkout.stripe.com/c/pay/cs_test_..."
}
```

#### Get Current Subscription

```bash
GET /api/v1/subscriptions/current
Authorization: Bearer YOUR_TOKEN

# Response:
{
  "id": "sub_123...",
  "tenant_id": "550e8400...",
  "plan": "professional",
  "status": "active",
  "current_period_end": "2024-01-15T00:00:00Z"
}
```

#### Cancel Subscription

```bash
POST /api/v1/subscriptions/cancel
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "cancel_immediately": false  # true = cancel now, false = cancel at period end
}
```

### 🔔 Stripe Webhooks

CarFlow handles the following Stripe webhook events:

| Event | Action |
|-------|--------|
| `checkout.session.completed` | Create subscription in database |
| `customer.subscription.created` | Activate subscription |
| `customer.subscription.updated` | Update subscription status |
| `customer.subscription.deleted` | Mark subscription as canceled |
| `invoice.payment_succeeded` | Send payment confirmation email |
| `invoice.payment_failed` | Send payment failure notification |

#### Setup Webhook Forwarding (Development)

```bash
# Install Stripe CLI
brew install stripe/stripe-cli/stripe

# Login to Stripe
stripe login

# Forward webhooks to your local server
stripe listen --forward-to localhost:8080/api/v1/webhooks/stripe

# Copy the webhook signing secret to .env
STRIPE_WEBHOOK_SECRET=whsec_xxx...
```

#### Configure Webhooks (Production)

1. Go to Stripe Dashboard → Webhooks
2. Add endpoint: `https://yourdomain.com/api/v1/webhooks/stripe`
3. Select events:
   - `checkout.session.completed`
   - `customer.subscription.*`
   - `invoice.payment_*`
4. Copy webhook signing secret to production `.env`

### 🧪 Test Subscription Flow

```bash
# Run the subscription test script
./scripts/test-subscriptions.sh

# Or test manually
# 1. Register a new account
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"dealership_name":"Test Motors","owner_email":"test@motors.com","owner_password":"password123","owner_first_name":"John","owner_last_name":"Doe"}'

# 2. Create checkout session
curl -X POST http://localhost:8080/api/v1/subscriptions/checkout \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"plan":"professional","success_url":"http://localhost:3000/success","cancel_url":"http://localhost:3000/billing"}'

# 3. Visit the checkout_url to complete payment (use test card: 4242 4242 4242 4242)

# 4. Check subscription status
curl -X GET http://localhost:8080/api/v1/subscriptions/current \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 💡 Stripe Test Cards

Use these test cards in development:

| Card Number | Description |
|-------------|-------------|
| `4242 4242 4242 4242` | Successful payment |
| `4000 0000 0000 0341` | Requires authentication |
| `4000 0000 0000 0002` | Card declined |

**Expiry:** Any future date
**CVC:** Any 3 digits
**ZIP:** Any 5 digits

---

## 💰 Pricing & Plans

| Plan | Price/Month | Vehicles | Users | API Calls |
|------|-------------|----------|-------|-----------|
| **Starter** | $29 | 100 | 3 | 10,000/mo |
| **Professional** | $99 | 1,000 | 10 | 100,000/mo |
| **Enterprise** | $299 | Unlimited | Unlimited | Unlimited |

---

## 📈 Roadmap

See [ROADMAP.MD](./ROADMAP.MD) for the complete 90-day SaaS transformation plan.

### ✅ **Completed (Week 1-2)**
- NX Monorepo setup
- PostgreSQL database with migrations
- Multi-tenancy architecture
- JWT authentication
- Role-based access control
- Enhanced vehicle management
- Docker Compose setup
- Stripe payment integration
- Brevo email service
- User registration endpoints
- Subscription management endpoints

### 🔄 **In Progress (Week 3)**
- React dashboard with signup/login
- Admin dashboard
- Usage tracking and analytics

### 📅 **Coming Soon**
- Admin dashboard
- Webhook system
- Advanced analytics
- Mobile apps
- Marketplace features

---

## 🧪 Testing

```bash
# Run all tests
cd apps/api
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/car/...

# Benchmark tests
go test -bench=. ./...
```

---

## 📚 Documentation

- **[ROADMAP.MD](./ROADMAP.MD)** - Complete 90-day SaaS roadmap
- **[ARCHITECTURE.MD](./ARCHITECTURE.MD)** - Detailed architecture docs
- **[SAAS.MD](./SAAS.MD)** - SaaS transformation plan
- **[API Docs](http://localhost:8080/api-docs)** - OpenAPI 3.0 specification

---

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

## 🙏 Acknowledgments

Built with ❤️ using:
- [Go](https://golang.org/) - Backend language
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Redis](https://redis.io/) - Caching
- [Docker](https://www.docker.com/) - Containerization
- [NX](https://nx.dev/) - Monorepo tools

---

## 💬 Support

- 📧 Email: support@carflow.com
- 🐛 Issues: [GitHub Issues](https://github.com/joshbarros/golang-carflow-api/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/joshbarros/golang-carflow-api/discussions)

---

**Made with 🚗 by the CarFlow Team**
