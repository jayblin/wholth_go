package main

// #cgo LDFLAGS: -lwholth
import "C"

import (
	"fmt"
	"log"
	"net/http"
	"os"
	// "crypto/tls"
	// "crypto/x509"
	// "io"
	// "io/ioutil"

	"time"
	"wholth_go/logger"
	"wholth_go/route"
	"wholth_go/route/auth"
	"wholth_go/route/consumption_log"
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

func main() {
	fmt.Println("Starting up...")

	secrets := secret.LoadSecrets()

	wholth.Setup()
	wholth.SetPasswordEncryptionSecret(secrets[0])

	secret.SetCsrfSecret(secrets[1])
	secret.SetSessionSecret(secrets[2])
	secret.SetDomain(os.Getenv("DOMAIN"))
	secret.SetUseTemplateCache("" != os.Getenv("USE_TEMPLATE_CACHE"))
	secret.SetAllowRegistration("1" == os.Getenv("ALLOW_REGISTRATION"))
	port := os.Getenv("PORT")

	logger.Info("ENV ready")

	mux := http.NewServeMux()

	mux.Handle(
		"GET /",
		session.CsrfTokenGeneratorMiddleware(
			http.HandlerFunc(food.ListFoods)))

	mux.Handle(
		"GET /static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("./static")),
		))

	// https://matthewsetter.com/restrict-allowed-route-methods-go-122/
	// https://www.alexedwards.net/blog/making-and-using-middleware
	http.HandleFunc("GET /palette", palette)
	mux.Handle(
		"GET /authenticate",
		session.CsrfTokenGeneratorMiddleware(
			http.HandlerFunc(auth.HandleAuthentication)))
	mux.Handle(
		"POST /register",
		session.CsrfTokenValidatorMiddleware(
			session.CsrfTokenGeneratorMiddleware(
				http.HandlerFunc(auth.HandleRegistration))))
	mux.Handle(
		"POST /login",
		session.CsrfTokenValidatorMiddleware(
			session.CsrfTokenGeneratorMiddleware(
				http.HandlerFunc(auth.HandleLogin))))
	mux.Handle(
		"GET /ingredient",
		session.SessionMiddleware(
			http.HandlerFunc(ingredient.ListIngredients)))
	mux.Handle(
		"GET /nutrient",
		session.SessionMiddleware(
			http.HandlerFunc(nutrient.ListNutrients)))
	mux.Handle(
		"GET /consumption_log",
		session.CsrfTokenGeneratorMiddleware(
			auth.AuthenticationMiddleware(
				http.HandlerFunc(consumption_log.ListConsumptionLogs))))
	mux.Handle(
		"POST /consumption_log",
		session.CsrfTokenValidatorMiddleware(
			session.CsrfTokenGeneratorMiddleware(
				auth.AuthenticationMiddleware(
					http.HandlerFunc(consumption_log.PostConsumptionLog)))))
	mux.Handle(
		"GET /food",
		session.SessionMiddleware(
			http.HandlerFunc(food.ListFoods)))
	mux.Handle(
		"GET /food/{id}",
		session.CsrfTokenGeneratorMiddleware(
			http.HandlerFunc(food.GetFoodById)))
	mux.Handle(
		"POST /food",
		session.CsrfTokenValidatorMiddleware(
			session.CsrfTokenGeneratorMiddleware(
				auth.AuthenticationMiddleware(
					http.HandlerFunc(food.PostFood)))))

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
		Addr: ":" + port,
		Handler: mux,
		// TLSConfig: tlsConfig,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		ErrorLog: log.New(os.Stderr, "", 0),
	}

	// log.Fatal(server.ListenAndServeTLS("domain.crt", "domain.key"))
	log.Fatal(server.ListenAndServe())
	// log.Fatal(http.ListenAndServe(":" + port, mux))
	// log.Fatal(http.ListenAndServeTLS(":" + port, "domain.crt", "domain.key", mux))
}
