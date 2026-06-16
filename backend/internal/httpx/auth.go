package httpx

import (
	"net/http"
	"time"

	"biu-panel/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) setupStatus(w http.ResponseWriter, _ *http.Request) {
	ok, err := s.store.HasUser()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"initialized": ok})
}
func (s *Server) setup(w http.ResponseWriter, r *http.Request) {
	initialized, err := s.store.HasUser()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if initialized {
		writeError(w, 409, "系统已初始化")
		return
	}
	var req struct{ Username, Password string }
	if !decodeJSON(w, r, &req) {
		return
	}
	if len(req.Username) < 3 || len(req.Password) < 8 {
		writeError(w, 400, "账号至少 3 位，密码至少 8 位")
		return
	}
	if err := CreateInitialUser(s.store, req.Username, req.Password); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, map[string]bool{"ok": true})
}
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username, Password string
		Remember           bool
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	locked, err := s.isLocked(req.Username)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if locked {
		s.store.LogLogin(req.Username, false, clientIP(r), "locked")
		writeError(w, 429, "连续失败次数过多，请 15 分钟后再试")
		return
	}
	u, err := s.store.FindUser(req.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		s.store.LogLogin(req.Username, false, clientIP(r), "invalid credentials")
		writeError(w, 401, "账号或密码错误")
		return
	}
	token, err := randomToken()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	var expires *time.Time
	if req.Remember {
		t := time.Now().Add(30 * 24 * time.Hour)
		expires = &t
	}
	if err := s.store.SaveSession(token, u.ID, expires, req.Remember); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	cookie := &http.Cookie{Name: "biu_session", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode}
	if req.Remember {
		cookie.Expires = *expires
		cookie.MaxAge = int(time.Until(*expires).Seconds())
	}
	http.SetCookie(w, cookie)
	s.store.LogLogin(req.Username, true, clientIP(r), "login")
	writeJSON(w, 200, map[string]any{"username": u.Username})
}
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("biu_session"); err == nil {
		s.store.DeleteSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "biu_session", Value: "", Path: "/", MaxAge: -1})
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	u, err := s.currentUser(r)
	if err != nil {
		writeError(w, 401, "未登录")
		return
	}
	writeJSON(w, 200, map[string]any{"username": u.Username})
}
func (s *Server) isLocked(username string) (bool, error) {
	n, err := s.store.FailedLoginsSince(username, time.Now().Add(-15*time.Minute))
	return n >= 5, err
}
func (s *Server) currentUser(r *http.Request) (store.User, error) {
	c, err := r.Cookie("biu_session")
	if err != nil {
		return store.User{}, err
	}
	return s.store.UserBySession(c.Value)
}
