package method

import (
	"fmt"
	"net/http"

	"forum/application"
	"forum/errorhandle"
)

func Method(app *application.Application, args ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if args == nil {
				next.ServeHTTP(w, r)
				return
			}
			for _, m := range args {
				if m == r.Method {
					next.ServeHTTP(w, r)
					return
				}
			}

			app.InfoLog.Printf("Request denied for method: '%v'", r.Method)
			errorhandle.ClientError(app, w, r, http.StatusMethodNotAllowed, fmt.Sprintf("Method %s is not allowed", r.Method))
		})
	}
}
