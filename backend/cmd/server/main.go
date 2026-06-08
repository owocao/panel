package main

import (
	"log"
	"net/http"

	"biu-panel/backend/internal/config"
	"biu-panel/backend/internal/httpx"
	"biu-panel/backend/internal/store"
)

func main() {
	cfg := config.Load()
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer st.DB.Close()

	if cfg.AdminUser != "" && cfg.AdminPass != "" {
		// Environment-based initialization is applied only when no user exists.
		if initialized, err := st.HasUser(); err == nil && !initialized {
			if err := httpx.CreateInitialUser(st, cfg.AdminUser, cfg.AdminPass); err != nil {
				log.Fatal(err)
			}
		}
	}

	addr := ":" + cfg.Port
	log.Printf("biu-panel backend listening on %s", addr)
	if err := http.ListenAndServe(addr, httpx.New(cfg, st).Routes()); err != nil {
		log.Fatal(err)
	}
}
