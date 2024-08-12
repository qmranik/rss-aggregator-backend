package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/qmranik/rss-aggregator-backend/handlers"
	"github.com/qmranik/rss-aggregator-backend/helper"
	"github.com/qmranik/rss-aggregator-backend/internal/auth"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	"github.com/qmranik/rss-aggregator-backend/internal/stripe"

	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Retrieve necessary environment variables
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to the PostgreSQL database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	// Initialize Authenticator with environment variables and token TTLs
	authenticator := &auth.Authenticator{
		DB:              dbQueries,
		JWTSecretKey:    os.Getenv("JWT_SECRET_KEY"),
		JWTRefreshKey:   os.Getenv("JWT_REFRESH_KEY"),
		AccessTokenTTL:  15 * time.Minute,   // Access token TTL set to 15 minutes
		RefreshTokenTTL: 7 * 24 * time.Hour, // Refresh token TTL set to 7 days
	}

	// Initialize ApiConfig for handling user and feed-related requests
	apiCfg := handlers.ApiConfig{
		DB:   dbQueries,
		Auth: authenticator,
	}

	// Initialize UserHandler with Authenticator
	userHandler := &auth.UserHandler{
		Authenticator: authenticator,
	}

	// Initialize Stripe Client and PaymentHandler
	stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeSecretKey == "" {
		log.Fatal("STRIPE_SECRET_KEY environment variable is not set")
	}
	stripeClient := stripe.NewStripeClient(stripeSecretKey)

	paymentHandler := &stripe.PaymentHandler{
		StripeClient: stripeClient,
		DB:           dbQueries,
	}

	// Create a new router with CORS middleware
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Define v1 API routes
	v1Router := chi.NewRouter()

	// Auth Routes
	v1Router.Post("/register", apiCfg.HandlerUsersCreate)
	v1Router.Get("/verify", userHandler.VerifyUsername)
	v1Router.Get("/refresh", userHandler.RefreshToken)
	v1Router.Post("/logout", userHandler.Logout)
	v1Router.Get("/user", authenticator.MiddlewareAuth(apiCfg.HandlerGetUser))

	// Feed Routes
	v1Router.Post("/feeds", authenticator.MiddlewareAuth(apiCfg.HandlerFeedCreate))
	v1Router.Get("/feeds", apiCfg.HandlerGetFeeds)

	// Feed Follow Routes
	v1Router.Get("/feed_follows", authenticator.MiddlewareAuth(apiCfg.HandlerFeedFollowsGet))
	v1Router.Post("/feed_follows", authenticator.MiddlewareAuth(apiCfg.HandlerFeedFollowCreate))
	v1Router.Delete("/feed_follows/{feedFollowID}", authenticator.MiddlewareAuth(apiCfg.HandlerFeedFollowDelete))

	// Post Routes
	v1Router.Get("/posts", authenticator.MiddlewareAuth(apiCfg.HandlerPostsGet))

	// Payment Routes
	v1Router.Post("/create-payment-intent", authenticator.MiddlewareAuth(paymentHandler.CreatePaymentIntent))
	v1Router.Post("/refund", authenticator.MiddlewareAuth(paymentHandler.CreateRefund))
	v1Router.Post("/webhook", paymentHandler.HandleWebhook)

	// Health Check Route
	v1Router.Get("/healthz", handlers.HandlerReadiness)

	// Mount v1 routes
	router.Mount("/v1", v1Router)

	// Start server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start background tasks for scraping
	const collectionConcurrency = 10
	const collectionInterval = time.Minute
	go helper.StartScraping(dbQueries, collectionConcurrency, collectionInterval)

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
