package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Lyearn/mgod"
	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/dto"
	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/handler"
	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/middleware"
	"github.com/dhanuprys/infrantery-backend-go/internal/adapter/repository"
	"github.com/dhanuprys/infrantery-backend-go/internal/config"
	"github.com/dhanuprys/infrantery-backend-go/internal/core/service"
	"github.com/dhanuprys/infrantery-backend-go/pkg/logger"
	"github.com/dhanuprys/infrantery-backend-go/pkg/validation"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	cfg         *config.Config
	mongoClient *mongo.Client
	router      *gin.Engine
}

func NewServer(cfg *config.Config) (*Server, error) {
	// Setup MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDBURI))
	if err != nil {
		return nil, err
	}

	// Ping MongoDB
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	logger.Info().
		Str("database", cfg.MongoDBDatabase).
		Msg("Connected to MongoDB successfully")

	// Set default mgod connection
	db := client.Database(cfg.MongoDBDatabase)
	mgod.SetDefaultConnection(db)

	router := gin.New()
	server := &Server{
		cfg:         cfg,
		mongoClient: client,
		router:      router,
	}

	if err := server.setupDependencies(); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) setupDependencies() error {
	// Initialize repositories
	userRepo, err := repository.NewUserRepository("users")
	if err != nil {
		return err
	}

	refreshTokenRepo, err := repository.NewRefreshTokenRepository("refresh_tokens")
	if err != nil {
		return err
	}

	projectRepo, err := repository.NewProjectRepository("projects")
	if err != nil {
		return err
	}

	projectMemberRepo, err := repository.NewProjectMemberRepository("project_members")
	if err != nil {
		return err
	}

	noteRepo, err := repository.NewNoteRepository("notes")
	if err != nil {
		return err
	}

	diagramRepo, err := repository.NewDiagramRepository("diagrams")
	if err != nil {
		return err
	}

	nodeRepo, err := repository.NewNodeRepository("nodes")
	if err != nil {
		return err
	}

	nodeVaultRepo, err := repository.NewNodeVaultRepository("node_vaults")
	if err != nil {
		return err
	}

	// Initialize services
	jwtService := service.NewJWTService(
		s.cfg.JWTSecret,
		s.cfg.JWTAccessExpiry,
		s.cfg.JWTRefreshExpiry,
	)

	argon2Params := &service.Argon2Params{
		Memory:      s.cfg.Argon2Memory,
		Iterations:  s.cfg.Argon2Iterations,
		Parallelism: s.cfg.Argon2Parallelism,
		SaltLength:  s.cfg.Argon2SaltLength,
		KeyLength:   s.cfg.Argon2KeyLength,
	}

	authService := service.NewAuthService(
		userRepo,
		refreshTokenRepo,
		jwtService,
		argon2Params,
	)

	userService := service.NewUserService(
		userRepo,
		refreshTokenRepo,
		argon2Params,
	)

	projectService := service.NewProjectService(
		projectRepo,
		projectMemberRepo,
		userRepo,
		noteRepo,
		diagramRepo,
	)

	noteService := service.NewNoteService(
		noteRepo,
		projectMemberRepo,
		projectRepo,
	)

	diagramService := service.NewDiagramService(
		diagramRepo,
		projectMemberRepo,
		projectRepo,
		nodeRepo,
	)

	nodeService := service.NewNodeService(
		nodeRepo,
		diagramRepo,
		projectMemberRepo,
	)

	nodeVaultService := service.NewNodeVaultService(
		nodeVaultRepo,
		nodeRepo,
		diagramRepo,
		projectMemberRepo,
	)

	breadcrumbService := service.NewBreadcrumbService(
		projectRepo,
		noteRepo,
		diagramRepo,
		nodeRepo,
		nodeVaultRepo,
	)

	// Initialize validator
	validator := validation.NewValidationEngine()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, validator, s.cfg)
	profileHandler := handler.NewProfileHandler(userService, validator)
	projectHandler := handler.NewProjectHandler(projectService, userRepo, validator)
	noteHandler := handler.NewNoteHandler(noteService, validator)
	diagramHandler := handler.NewDiagramHandler(diagramService, validator)
	nodeHandler := handler.NewNodeHandler(nodeService, validator)
	nodeVaultHandler := handler.NewNodeVaultHandler(nodeVaultService, validator)
	breadcrumbHandler := handler.NewBreadcrumbHandler(breadcrumbService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	s.setupRoutes(authMiddleware, authHandler, profileHandler, projectHandler, noteHandler, diagramHandler, nodeHandler, nodeVaultHandler, breadcrumbHandler)

	return nil
}

func (s *Server) setupRoutes(
	authMiddleware *middleware.AuthMiddleware,
	authHandler *handler.AuthHandler,
	profileHandler *handler.ProfileHandler,
	projectHandler *handler.ProjectHandler,
	noteHandler *handler.NoteHandler,
	diagramHandler *handler.DiagramHandler,
	nodeHandler *handler.NodeHandler,
	nodeVaultHandler *handler.NodeVaultHandler,
	breadcrumbHandler *handler.BreadcrumbHandler,
) {
	// Add middlewares
	s.router.Use(gin.Recovery())                // Recovery middleware
	s.router.Use(middleware.LoggerMiddleware()) // Our custom logger middleware

	// CORS configuration
	s.router.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	}))

	s.router.NoRoute(func(c *gin.Context) {
		c.JSON(
			http.StatusNotFound,
			dto.NewAPIResponse[any](nil, dto.NewErrorResponse(dto.ErrCodePageNotFound)),
		)
	})

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Public routes
		public := v1.Group("")
		{
			public.POST("/auth/register", authHandler.Register)
			public.POST("/auth/login", authHandler.Login)
			public.POST("/auth/refresh", authHandler.RefreshToken)
			public.POST("/auth/logout", authHandler.Logout)
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			// Profile routes
			protected.GET("/profile", profileHandler.GetProfile)
			protected.PUT("/profile", profileHandler.UpdateProfile)
			protected.PUT("/profile/password", profileHandler.ChangePassword)

			// Project routes
			projects := protected.Group("/projects")
			{
				projects.POST("", projectHandler.CreateProject)
				projects.GET("", projectHandler.GetUserProjects)
				projects.GET("/:project_id", projectHandler.GetProjectDetails)
				projects.PUT("/:project_id", projectHandler.UpdateProject)
				projects.DELETE("/:project_id", projectHandler.DeleteProject)

				// Breadcrumbs
				projects.GET("/:project_id/breadcrumbs", breadcrumbHandler.GetBreadcrumbs)

				// Project member management
				projects.POST("/:project_id/members", projectHandler.AddMember)
				projects.GET("/:project_id/members", projectHandler.GetMembers)
				projects.PUT("/:project_id/members/:user_id", projectHandler.UpdateMember)
				projects.DELETE("/:project_id/members/:user_id", projectHandler.RemoveMember)

				// Note management
				projects.POST("/:project_id/notes", noteHandler.CreateNote)
				projects.GET("/:project_id/notes", noteHandler.ListNotes)
				projects.GET("/:project_id/notes/:note_id", noteHandler.GetNote)
				projects.PUT("/:project_id/notes/:note_id", noteHandler.UpdateNote)
				projects.DELETE("/:project_id/notes/:note_id", noteHandler.DeleteNote)

				// Diagram management
				projects.POST("/:project_id/diagrams", diagramHandler.CreateDiagram)
				projects.GET("/:project_id/diagrams", diagramHandler.ListDiagrams)
				projects.GET("/:project_id/diagrams/:diagram_id", diagramHandler.GetDiagram)
				projects.PUT("/:project_id/diagrams/:diagram_id", diagramHandler.UpdateDiagram)
				projects.DELETE("/:project_id/diagrams/:diagram_id", diagramHandler.DeleteDiagram)

				// Node management
				projects.GET("/:project_id/diagrams/:diagram_id/nodes/:node_id", nodeHandler.GetOrCreateNode)
				projects.PUT("/:project_id/diagrams/:diagram_id/nodes/:node_id", nodeHandler.UpdateNode)
				projects.DELETE("/:project_id/diagrams/:diagram_id/nodes/:node_id", nodeHandler.DeleteNode)

				// Node Vault management
				projects.GET("/:project_id/diagrams/:diagram_id/nodes/:node_id/vault", nodeVaultHandler.ListVaultItems)
				projects.GET("/:project_id/diagrams/:diagram_id/nodes/:node_id/vault/:vault_id", nodeVaultHandler.GetVaultItem)
				projects.POST("/:project_id/diagrams/:diagram_id/nodes/:node_id/vault", nodeVaultHandler.CreateVaultItem)
				projects.PUT("/:project_id/diagrams/:diagram_id/nodes/:node_id/vault/:vault_id", nodeVaultHandler.UpdateVaultItem)
				projects.DELETE("/:project_id/diagrams/:diagram_id/nodes/:node_id/vault/:vault_id", nodeVaultHandler.DeleteVaultItem)
			}
		}
	}
}

func (s *Server) Run() error {
	logger.Info().Str("port", s.cfg.Port).Msg("Server starting")
	return s.router.Run(":" + s.cfg.Port)
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info().Msg("Server shutting down...")
	if err := s.mongoClient.Disconnect(ctx); err != nil {
		return err
	}
	logger.Info().Msg("MongoDB connection closed")
	return nil
}
