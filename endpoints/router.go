package endpoints

import (
	"auth_service_template/cache_storage"
	"auth_service_template/firewall"
	"auth_service_template/models"

	"github.com/gorilla/mux"
)

type Router struct {
	router       *mux.Router
	environments *map[string]string
	db           *models.DB
	cache        *cache_storage.TimeStorage
	wall         firewall.Firewall
}

func NewRouter(env *map[string]string, dbConn *models.DB,
	storage *cache_storage.TimeStorage, f firewall.Firewall) *Router {
	return &Router{
		router:       mux.NewRouter().StrictSlash(false),
		environments: env,
		db:           dbConn,
		cache:        storage,
		wall:         f,
	}
}

func (r *Router) LoadRoutes() {
	// r.router.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println("Success")
	// })
	// Auth routes group.
	authRouter := r.router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/signin", r.handleSignin() /*r.middlewareSignin(r.handleSignin())*/)
	authRouter.HandleFunc("/signup", r.handleSignup()) /*.
	Methods("POST").
	Schemes((*r.environments)["APP_SHEME"]) //.
	// Host((*r.environments)["APP_HOST"])*/
	authRouter.HandleFunc("/refresh", r.handleRefresh())
	authRouter.HandleFunc("/test", r.wall.BunHttpMiddleware(r.wall.LimitHttpMiddleware(r.handleAny()))).Methods("GET")
}

func (r *Router) GetRoutes() *mux.Router {
	//r.LoadRoutes()
	return r.router
}
