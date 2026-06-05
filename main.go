package main

// #cgo LDFLAGS: -lwholth
import "C"

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"

	// "crypto/tls"
	// "crypto/x509"
	// "io"
	// "io/ioutil"

	"time"
	"wholth_go/container"
	"wholth_go/logger"
	"wholth_go/route"
	"wholth_go/route/auth"
	"wholth_go/route/body_part"
	"wholth_go/route/consumption_log"
	"wholth_go/route/exercise"
	"wholth_go/route/food"
	"wholth_go/route/ingredient"
	"wholth_go/route/nutrient"
	"wholth_go/secret"
	"wholth_go/session"
	"wholth_go/wholth"
)

func palette(w http.ResponseWriter, r *http.Request) {
	page := route.DefaultHtmlPage(r)
	page.Meta.Title = "Pallete"
	page.Meta.Description = "Color pallete preview"

	route.RenderHtmlTemplates(
		w,
		r,
		page,
		"templates/index.html",
		"templates/palette/page.html",
	)
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}

		next.ServeHTTP(gzr, r)
	})
}

func matchMiddleware(alias string) func(next http.Handler) http.Handler {
	switch alias {
	case "gzip":
		return gzipMiddleware
	case "container":
		return container.ContainerMiddleware
	case "session":
		return session.SessionMiddleware
	case "csrf-generator":
		return session.CsrfTokenGeneratorMiddleware
	case "csrf-validator":
		return session.CsrfTokenValidatorMiddleware
	case "authentication":
		return auth.AuthenticationMiddleware
	}

	return nil
}

func applyMiddleware(handler func(w http.ResponseWriter, r *http.Request), middleware ...string) http.Handler {
	var chain http.Handler = http.HandlerFunc(handler)

	slices.Reverse(middleware)

	for _, m := range middleware {
		chain = matchMiddleware(m)(chain)
	}

	return chain
}

func Index(w http.ResponseWriter, r *http.Request) {
	route.RenderHtmlTemplates(
		w,
		r,
		route.DefaultHtmlPage(r),
		"templates/index.html",
	)
}

// func Test(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Add("Location", "/exercise/1")
// 	w.WriteHeader(http.StatusSeeOther)
// }

func main() {
	logger.Info("Starting up...")

	secrets := secret.LoadSecrets()

	wholth.Setup()
	wholth.SetPasswordEncryptionSecret(secrets[0])

	secret.SetCsrfSecret(secrets[1])
	secret.SetSessionSecret(secrets[2])
	secret.SetDomain(os.Getenv("DOMAIN"))
	secret.SetUseTemplateCache("" != os.Getenv("USE_TEMPLATE_CACHE"))
	secret.SetAllowRegistration("1" == os.Getenv("ALLOW_REGISTRATION"))
	secret.SetVersion(os.Getenv("VERSION"))
	port := os.Getenv("PORT")

	logger.Info("ENV ready")

	mux := http.NewServeMux()
	//
	// mux.Handle(
	// 	"GET /",
	// 	session.SessionMiddleware(
	// 		session.CsrfTokenGeneratorMiddleware(
	// 			http.HandlerFunc(food.ListFoods))))

	mux.Handle(
		"GET /static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("./static")),
		))

	// https://matthewsetter.com/restrict-allowed-route-methods-go-122/
	// https://www.alexedwards.net/blog/making-and-using-middleware
	// http.HandleFunc("GET /palette", palette)

	routes := []struct {
		RouteName      string
		Handler        func(http.ResponseWriter, *http.Request)
		MiddlewareList []string
	}{
		{
			"GET /",
			Index,
			[]string{"gzip", "container", "session", "csrf-generator", "authentication"},
		},
		// {
		// 	"POST /test",
		// 	Test,
		// 	[]string{},
		// },
		{
			"GET /authenticate",
			auth.HandleAuthentication,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator"},
		},
		{
			"POST /register",
			auth.HandleRegistration,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator"},
		},
		{
			"POST /login",
			auth.HandleLogin,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator"},
		},
		{
			"GET /ingredient",
			ingredient.ListIngredients,
			[]string{"gzip", "container", "session", "csrf-generator", "authentication"},
		},
		{
			"GET /nutrient",
			nutrient.ListNutrients,
			[]string{"gzip", "container", "session", "csrf-generator", "authentication"},
		},
		{
			"POST /consumption_log/batch-patch",
			consumption_log.BatchPatchConsumptionLog,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator", "authentication"},
		},
		{
			"POST /consumption_log/batch-delete",
			consumption_log.BatchDeleteConsumptionLog,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator", "authentication"},
		},
		{
			"GET /consumption_log",
			consumption_log.ListConsumptionLogs,
			[]string{"gzip", "container", "session", "csrf-generator", "authentication"},
		},
		{
			"POST /consumption_log",
			consumption_log.PostConsumptionLog,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator", "authentication"},
		},
		{
			"GET /food",
			food.ListFoods,
			[]string{"gzip", "container", "session", "csrf-generator", "authentication"},
		},
		{
			"GET /food/{id}",
			food.GetFoodById,
			[]string{"gzip", "container", "session", "csrf-generator", "authentication"},
		},
		{
			"GET /recipe/add",
			food.GetRecipeForm,
			[]string{"gzip", "container", "session", "csrf-generator", "authentication"},
		},
		{
			"GET /recipe/{id}/copy",
			food.GetRecipeCopyForm,
			[]string{"gzip", "container", "session", "csrf-generator", "authentication"},
		},
		{
			"POST /recipe/{id}/copy",
			food.CopyRecipe,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator", "authentication"},
		},
		{
			"POST /food",
			food.PostFood,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator", "authentication"},
		},
		{
			"GET /exercise",
			exercise.ListExercises,
			[]string{"gzip", "container", "session", "csrf-generator", "authentication"},
		},
		{
			"GET /exercise/{id}",
			exercise.GetExerciseForm,
			[]string{"gzip", "container", "session", "csrf-generator", "authentication"},
		},
		{
			"POST /exercise",
			exercise.PostExercise,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator", "authentication"},
		},
		{
			"GET /exercise_log",
			exercise.ListExerciseLogs,
			[]string{"gzip", "container", "session", "csrf-generator", "authentication"},
		},
		{
			"POST /exercise_log",
			exercise.PostExerciseLog,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator", "authentication"},
		},
		{
			"POST /exercise_log/batch-patch",
			exercise.BatchPatchExerciseLog,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator", "authentication"},
		},
		{
			"POST /exercise_log/batch-delete",
			exercise.BatchDeleteExerciseLog,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator", "authentication"},
		},
		{
			"GET /body_part",
			body_part.ListBodyParts,
			[]string{"gzip", "container", "session", "csrf-generator", "authentication"},
		},
		{
			"POST /body_part",
			body_part.PostBodyPart,
			[]string{"gzip", "container", "session", "csrf-validator", "csrf-generator", "authentication"},
		},
	}

	for _, r := range routes {
		mux.Handle(r.RouteName, applyMiddleware(r.Handler, r.MiddlewareList...))
		logger.Info(fmt.Sprintf("Registered '%s'", r.RouteName))
	}

	logger.Info("Routes ready")

	// clientTLSCert, err := tls.LoadX509KeyPair("domain.crt", "domain.key")
	// if nil != err {
	// 	log.Fatalf("Error loading certificate and key file: %v")
	// 	panic(err)
	// }

	// certPool, err := x509.SystemCertPool()
	// if nil != err {
	// 	panic(err)
	// }

	// if caCertPEM, err := ioutil.ReadFile("domain.crt"); err != nil {
	// 	panic(err)
	// } else if ok := certPool.AppendCertsFromPEM(caCertPEM); !ok {
	// 	panic("invalid cert in CA PEM")
	// }

	// tlsConfig := &tls.Config{
	// 	RootCAs:      certPool,
	// 	Certificates: []tls.Certificate{clientTLSCert},
	// }
	// tr := &http.Transport{
	// 	TLSClientConfig: tlsConfig,
	// }

	logger.Info("Serving...")

	// tlsConfig := &tls.Config{
	// 	MinVersion:               tls.VersionTLS12,
	// 	CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
	// 	PreferServerCipherSuites: true,
	// 	CipherSuites: []uint16{
	// 		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	// 		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	// 		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	// 		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	// 		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	// 	},
	// }

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
		// TLSConfig: tlsConfig,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		ErrorLog:     log.New(os.Stderr, "", 0),
	}

	// log.Fatal(server.ListenAndServeTLS("domain.crt", "domain.key"))
	log.Fatal(server.ListenAndServe())
	// log.Fatal(http.ListenAndServe(":" + port, mux))
	// log.Fatal(http.ListenAndServeTLS(":" + port, "domain.crt", "domain.key", mux))
}
