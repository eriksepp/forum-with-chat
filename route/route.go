package route

import (
	"net/http"

	"forum/application"
	"forum/controllers"
	"forum/logger"
	"forum/route/middleware/acl"
	"forum/route/middleware/log"
	"forum/route/middleware/method"
)

type Mux struct {
	Mux        *http.ServeMux
	Middleware []func(http.Handler) http.Handler
	Pattern    string
}

func (m *Mux) Then(handler http.Handler) {
	if handler == nil {
		logger.Error("HTTP: nil handler")
		return
	}
	result := handler
	//we atually want it reverse, so exec goes form left to right
	/*
		//old
		for _, f := range m.Middleware {
			result = f(result)

		}
	*/
	for idx := range m.Middleware {
		result = m.Middleware[len(m.Middleware)-idx-1](result)
	}
	m.Mux.Handle(m.Pattern, result)
}

func (m *Mux) ThenFunc(fn http.HandlerFunc) {
	if fn == nil {
		m.Then(nil)
		return
	}
	m.Then(fn)
}

func (m *Mux) Handle(pattern string, middleware ...func(http.Handler) http.Handler) *Mux {
	return &Mux{Mux: m.Mux, Pattern: pattern, Middleware: middleware}
}

func Load(app *application.Application) http.Handler {
	return middleware(app, routes(app))
}

func routes(app *application.Application) *http.ServeMux {
	var (
		GET  = method.Method(app, "GET")  // Only allow GET requests
		//POST = method.Method(app, "POST") // Only allow POST requests
	)

	r := Mux{Mux: http.NewServeMux()}

	r.Handle("/", GET).ThenFunc(controllers.Index(app))
	r.Handle("/ws").ThenFunc(controllers.IndexWs(app)) // TODO-Solved do we need GET here?-Answer: no, gorilla/websocket checks it in Upgrader

	// r.Handle("/chat/", GET, acl.DisallowAnon(app)).ThenFunc(controllers.OpenChat(app))
	// r.Handle("/chatws/", acl.DisallowAnon(app)).ThenFunc(controllers.HandleChatWs(app))


	// r.Handle("/profile/", GET, acl.DisallowAnon(app)).ThenFunc(controllers.ProfileGET(app))

	// r.Handle("/reactions", POST).ThenFunc(controllers.ReactionsPOST(app)) // no alc, contoller handles it because need unifed way to report errors (session expired)

	// GitHub
	// r.Handle("/login/github").ThenFunc(controllers.OAuthGitHub(app))
	// r.Handle("/login/github/callback").ThenFunc(controllers.OAuthGitHubCallback(app))

	// // Google
	// r.Handle("/login/google").ThenFunc(controllers.OAuthGoogle(app))
	// r.Handle("/login/google/callback").ThenFunc(controllers.OAuthGoogleCallback(app))

	staticDirectory := http.Dir("webui/static")
	staticServer := http.FileServer(staticDirectory)
	r.Handle("/static/").Then(http.StripPrefix("/static/", staticServer))

	return r.Mux
}

// this will applay middleware to all controllers
func middleware(app *application.Application, h http.Handler) http.Handler {
	// Print info about the request
	h = log.Log(app, h)
	h = acl.AddUser(app)(h)
	return h
}
