# RSS Aggregator Backend

## Overview

This project is an RSS Aggregator that enables users to follow and keep up with their favorite blog posts by collecting RSS feeds. The aggregator periodically (every 10 minutes) checks for new or updated feeds and stores them in a PostgreSQL database. Additionally, the backend supports user authentication using JWT tokens, integrates Stripe for payment processing, and provides robust logging and error handling.

## Features

- **RSS Feed Aggregation:** Collects and updates RSS feeds every 10 minutes.
- **User Authentication:** JWT-based authentication with refresh tokens.
- **Payment Integration:** Supports Stripe for payment processing, including refunds and webhooks for payment validation.
- **Database:** Uses PostgreSQL for storing users, feeds, sessions, and payment data.
- **Concurrency:** Efficiently fetches and processes feeds concurrently.
- **Migrations:** Database schema managed with `goose` for easy migration.
- **Logging:** Utilizes `logrus` for comprehensive logging and error tracking.

## Project Structure

```
.
├── Makefile
├── go.mod
├── go.sum
├── handlers
│   ├── config.go
│   ├── feed.go
│   ├── feed_follows.go
│   ├── posts.go
│   ├── ready.go
│   └── user.go
├── helper
│   ├── json.go
│   ├── jwt.go
│   └── scraper.go
├── internal
│   ├── auth
│   │   ├── auth.go
│   │   ├── auth_middleware.go
│   │   ├── handlers.go
│   │   └── models.go
│   ├── database
│   │   ├── auth.sql.go
│   │   ├── db.go
│   │   ├── feed_follows.sql.go
│   │   ├── feeds.sql.go
│   │   ├── models.go
│   │   ├── payment.sql.go
│   │   ├── posts.sql.go
│   │   └── users.sql.go
│   └── stripe
│       ├── client.go
│       └── webhook.go
├── main.go
├── models
│   ├── feeds.go
│   ├── models.go
│   ├── post.go
│   ├── rss.go
│   └── stripe.go
├── readme.md
├── sql
│   ├── queries
│   │   ├── auth.sql
│   │   ├── feed_follows.sql
│   │   ├── feeds.sql
│   │   ├── payment.sql
│   │   ├── posts.sql
│   │   └── users.sql
│   └── schema
│       ├── 001_users.sql
│       ├── 002_users_apikey.sql
│       ├── 003_feeds.sql
│       ├── 004_feed_follows.sql
│       ├── 005_feed_lastfetched.sql
│       ├── 006_posts.sql
│       ├── 007_users.sql
│       ├── 008_jwt.sql
│       └── 009_payment.sql
└── sqlc.yaml
```

## Getting Started

### Prerequisites

- Go 1.19+
- PostgreSQL
- Stripe Account (for payment integration)

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/qmranik/rss-aggregator-backend.git
   cd rss-aggregator-backend
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables:**
   Create a `.env` file in the root directory with the following content:
   ```env
   PORT=8080
   DATABASE_URL=postgres://username:password@localhost:5432/yourdatabase
   JWT_SECRET_KEY=your_jwt_secret
   JWT_REFRESH_KEY=your_jwt_refresh_secret
   STRIPE_SECRET_KEY=your_stripe_secret_key
   STRIPE_WEBHOOK_SECRET=your_stripe_webhook_secret
   ```

4. **Run database migrations:**
   Install `goose` and run the migrations:
   ```bash
   go install github.com/pressly/goose/v3/cmd/goose@latest
   cd sql/schema
   goose postgres "your_database_connection_string" up
   ```

5. **Generate SQL code with `sqlc`:**
   Install `sqlc` and generate the SQL code:
   ```bash
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   sqlc generate
   ```

6. **Run the application:**
   ```bash
   make run
   ```

## Usage

- **User Registration:** Users can register and log in to follow RSS feeds.
- **Feed Management:** Users can add, view, and follow RSS feeds.
- **Payment:** Users can make payments through Stripe and request refunds.
- **Webhooks:** Stripe webhooks are used to validate and process payment events.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Reference
This project is an RSS Aggregator backend service, inspired from [this video](https://www.youtube.com/watch?v=un6ZyFkqFKo) by Lane Wagner but modified to include advanced features such as JWT authentication, PostgreSQL database, and full Stripe payment integration. The service allows users to follow RSS feeds, which are fetched concurrently every 10 minutes, and stores the data in a PostgreSQL database.