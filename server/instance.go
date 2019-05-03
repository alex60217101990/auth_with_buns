package server

import (
	"auth_service_template/cache_storage"
	"auth_service_template/endpoints"
	"auth_service_template/firewall"
	logs "auth_service_template/logger"
	"auth_service_template/models"
	"context"
	"fmt"
	"net/http"
	"time"
)

type Instance struct {
	db         *models.DB
	storage    *cache_storage.TimeStorage
	wall       firewall.Firewall
	httpServer *http.Server
	logger     *logs.Log
	env        *map[string]string
}

func NewInstance(l *logs.Log, e *map[string]string) *Instance {
	w := firewall.NewFirewall(l, e)
	s := &Instance{
		db:      models.NewDB(e, l),
		storage: cache_storage.NewTimeStorage(e, l),
		wall:    w,
		logger:  l,
		env:     e,
	}
	return s
}

func (s *Instance) Start() {
	ctx_firewall, cancel_firewall := context.WithCancel(context.Background())
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case error:
				s.db.Close()
				s.storage.Close()
				s.logger.Println(logs.NewError(
					"instance",
					fmt.Sprintf("Auth instance was stoped with panic: %v", x.Error()),
					logs.ERROR,
				))
				return
			}
		}
		cancel_firewall()
		s.db.Close()
		s.storage.Close()
		s.logger.Println(logs.NewError(
			"instance",
			fmt.Sprintf("Auth instance was stoped success"),
			logs.INFO,
		))
	}()
	// Startup the http Server in a way that
	// we can gracefully shut it down again
	router := endpoints.NewRouter(s.env, s.db, s.storage, s.wall)
	s.httpServer = &http.Server{Addr: ":8070", //fmt.Sprintf("%s:%s", (*s.env)["APP_HOST"], (*s.env)["APP_PORT"]),
		Handler: router.GetRoutes(),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  time.Second * 60,
	}
	router.LoadRoutes()
	go s.wall.CleanupVisitors(ctx_firewall)
	err := s.httpServer.ListenAndServe() // Blocks!
	if err != http.ErrServerClosed {
		s.logger.Println(logs.NewError(
			"instance",
			fmt.Sprintf("Http Server stopped unexpected: %v", err),
			logs.FATAL,
		))
		s.Shutdown()
	} else {
		s.logger.Println(logs.NewError(
			"instance",
			"Http Server stopped",
			logs.INFO,
		))
	}
}

func (s *Instance) Shutdown() {
	if s.httpServer != nil {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			s.logger.Println(logs.NewError(
				"instance",
				fmt.Sprintf("Failed to shutdown http server gracefully: %v", err),
				logs.ERROR,
			))
		} else {
			s.httpServer = nil
		}
	}
}
