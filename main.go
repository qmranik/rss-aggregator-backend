package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/qmranik/rss-aggregator-backend/handlers"
	"github.com/qmranik/rss-aggregator-backend/internal/ai"
	"github.com/qmranik/rss-aggregator-backend/internal/auth"
	"github.com/qmranik/rss-aggregator-backend/internal/database"
	"github.com/qmranik/rss-aggregator-backend/internal/queue"
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

	// Source Routes
	v1Router.Post("/sources", authenticator.MiddlewareAuth(apiCfg.HandlerSourceCreate))
	v1Router.Get("/sources", apiCfg.HandlerGetSources)

	// Source Follow Routes
	v1Router.Get("/source_follows", authenticator.MiddlewareAuth(apiCfg.HandlerSourceFollowsGet))
	v1Router.Post("/source_follows", authenticator.MiddlewareAuth(apiCfg.HandlerSourceFollowCreate))
	v1Router.Delete("/source_follows/{sourceFollowID}", authenticator.MiddlewareAuth(apiCfg.HandlerSourceFollowDelete))

	// Post Routes
	v1Router.Get("/posts", authenticator.MiddlewareAuth(apiCfg.HandlerPostsGet))

	v1Router.Post("/chat", authenticator.MiddlewareAuth(apiCfg.HandlerChat))

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

	// Initialize Queue Worker and Client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	queueClient := queue.NewClient(redisURL)
	defer queueClient.Close()

	aiClient, err := ai.NewClient(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to initialize AI client: %v", err)
	}

	queueWorker := queue.NewWorker(redisURL, dbQueries, aiClient, 10)
	go queueWorker.Start()
	defer queueWorker.Stop()

	// Start background tasks for scheduling
	const schedulingInterval = time.Minute
	go queue.StartScheduling(dbQueries, queueClient, schedulingInterval)

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
