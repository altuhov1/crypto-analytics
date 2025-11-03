package app

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crypto-analytics/internal/config"
	"crypto-analytics/internal/handlers"
	"crypto-analytics/internal/services"
	"crypto-analytics/internal/storage"
)

type App struct {
	cfg      *config.Config
	server   *http.Server
	services *Services
	storages *Storages
}

type Services struct {
	notifier services.Notifier
	crypto   *services.CryptoService
	news     *services.NewsService
	users    *services.UserService
	pairs    *services.CryptoPairsService
	analysis *services.AnalysisService
	sysStat  *services.SystemMonitor
	posts    *services.PostsService
}

type Storages struct {
	contacts     storage.FormStorage
	users        storage.UserStorage
	news         storage.NewsStorage
	pairs        storage.CacheStorage
	anslysis     storage.AnalysisStorage
	analysisTemp storage.AnalysisTempStorage
	posts        storage.PostStorage
}

func NewApp(cfg *config.Config) *App {

	app := &App{
		cfg: cfg,
	}

	app.initStorages()
	app.initServices()
	app.initHTTP()

	return app
}

func (a *App) initStorages() {
	dbPGConfig := &config.PGXConfig{
		Host:     a.cfg.PG_DBHost,
		User:     a.cfg.PG_DBUser,
		Password: a.cfg.PG_DBPassword,
		DBName:   a.cfg.PG_DBName,
		SSLMode:  a.cfg.PG_DBSSLMode,
		Port:     a.cfg.PG_PORT,
	}
	dbMongoConfig := &config.MGConfig{
		DBUser:     a.cfg.MG_DBUser,
		DBPassword: a.cfg.MG_DBPassword,
		DBHost:     a.cfg.MG_DBHost,
		DBName:     a.cfg.MG_DBName,
		DBAuth:     a.cfg.MG_Auth,
		Port:       a.cfg.MG_Port,
	}
	redisCfg := &config.RedisConfig{
		Host:     a.cfg.RedisHost,
		Password: a.cfg.RedisPassword,
		DB:       a.cfg.RedisDB,
		PoolSize: a.cfg.RedisPoolSize,
		Port:     a.cfg.RedisPort,
	}
	poolPG, err := storage.NewPoolPg(dbPGConfig)
	if err != nil {
		slog.Error("Failed to initialize PG (pool)", "error", err)
		os.Exit(1)
	}

	clientMG, err := storage.NewMongoClient(dbMongoConfig)
	if err != nil {
		slog.Error("Failed to initialize PG (pool)", "error", err)
		os.Exit(1)
	}
	redisClient, err := storage.NewRedisClient(redisCfg)
	if err != nil {
		slog.Error("Failed to initialize PG (pool)", "error", err)
		os.Exit(1)
	}
	contactsStorage := storage.NewContactPostgresStorage(poolPG)

	usersStorage := storage.NewUserPostgresStorage(poolPG)

	postStorage := storage.NewPostsMongoStorage(clientMG)
	reddisAnalysis := storage.NewAnalysisTempStorage(redisClient)

	newsStorage := storage.NewNewsFileStorage("storage/news_cache.json")

	pairsStorage := storage.NewPairsFileStorage("storage/pairs_cache.json")

	analysisStorage := storage.NewAnalysisFileStorage("storage/analysis_cache.json")

	a.storages = &Storages{
		contacts:     contactsStorage,
		users:        usersStorage,
		news:         newsStorage,
		pairs:        pairsStorage,
		anslysis:     analysisStorage,
		analysisTemp: reddisAnalysis,
		posts:        postStorage,
	}
}

func (a *App) initServices() {
	IsItProd := false
	if a.cfg.LaunchLoc == "prod" {
		IsItProd = true
	} else {
		IsItProd = false
	}
	a.services = &Services{
		notifier: services.NewNotifier(),
		crypto:   services.NewCryptoService(IsItProd, "storage/crypto_cache.json"),
		news:     services.NewNewsService(a.storages.news, IsItProd),
		users:    services.NewUserService(a.storages.users),
		pairs:    services.NewCryptoPairsService(a.storages.pairs, IsItProd),
		analysis: services.NewAnalysisService(IsItProd, a.storages.anslysis, a.storages.analysisTemp),
		sysStat:  services.NewSystemMonitor(),
		posts:    services.NewPostService(a.storages.posts),
	}
}

func (a *App) initHTTP() {
	go a.services.sysStat.StartStatsReporter()
	handler, err := handlers.NewHandler(
		a.storages.contacts,
		a.services.notifier,
		a.services.crypto,
		a.services.users,
		a.cfg.KeyUsersGorilla,
		a.services.news,
		a.services.pairs,
		a.services.analysis,
		a.services.posts,
	)
	if err != nil {
		slog.Error("Failed to create handler", "error", err)
		os.Exit(1)
	}

	router := a.setupRoutes(handler)

	a.server = &http.Server{
		Addr:         ":" + a.cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
}

func (a *App) setupRoutes(handler *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// API routes
	apiRoutes := map[string]http.HandlerFunc{
		"/api/allFavoriteCoin":    handler.GetFavorites,
		"/api/changeFavoriteCoin": handler.ChangeFavorite,
		"/api/printUserstInfo":    handler.InfoOfUsers,
		"/api/printContactInfo":   handler.InfoOfContacts,
		"/api/cache-info":         handler.CacheInfoHandler,
		"/api/all-pairs":          handler.GetAllPairsHandler,
		"/api/select-pair":        handler.SelectPairHandler,
		"/api/pair":               handler.GetPairInfo,
		"/api/available":          handler.GetAvailablePairs,
		"/api/posts/create":       handler.CreatePostHandler,
		"/api/comments/create":    handler.CreateCommentHandler,
		"/api/posts":              handler.GetPostsHandler,
		"/api/comments":           handler.GetCommentsHandler,
		"/api/posts/update":       handler.UpdatePostHandler,
		"/api/posts/delete":       handler.UpdatePostHandler,
		"/api/comments/update":    handler.UpdateCommentHandler,
		"/api/comments/delete":    handler.DeleteCommentHandler,
	}

	for path, handlerFunc := range apiRoutes {
		mux.HandleFunc(path, handlerFunc)
	}

	// Web routes
	webRoutes := map[string]http.HandlerFunc{
		"/news":          handler.NewsPage,
		"/pairs":         handler.CryptoPairsPageHandler,
		"/logout":        handler.LogoutHandler,
		"/login":         handler.LoginHandler,
		"/check-Sess-Id": handler.CheckAuthHandler,
		"/register":      handler.AuthUserFormHandler,
		"/contact":       handler.ContactFormHandler,
		"/crypto-top":    handler.CryptoTopHandler,
	}

	for path, handlerFunc := range webRoutes {
		mux.HandleFunc(path, handlerFunc)
	}
	if a.cfg.LaunchLoc == "prod" {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	}
	// Root route
	mux.HandleFunc("/", a.rootHandler)

	return mux
}

func (a *App) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.ServeFile(w, r, "static/index.html")
}

func (a *App) Run() {
	go a.startServer()
	a.waitForShutdown()
}

func (a *App) startServer() {
	slog.Info("Server starting", "port", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func (a *App) waitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down server gracefully...")
	a.shutdown()
}

func (a *App) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	a.storages.contacts.Close()
	a.storages.users.Close()
	a.storages.posts.Close()
	slog.Info("Server stopped")
	if a.cfg.LaunchLoc == "prod" {
		time.Sleep(1 * time.Second)
	}
}
