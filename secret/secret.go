package secret

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"wholth_go/logger"
)

func LoadSecrets() []string {
	// https://zetcode.com/golang/readinput/
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Secrets: ")
	secrets_raw, read_err := reader.ReadString('§')

	if nil != read_err {
		panic(read_err.Error())
	}

	fmt.Println('\n')
	logger.Info("Loaded secrets")

	secrets := strings.Split(secrets_raw, "\n")

	if len(secrets) != 4 {
		logger.Emergency("AGORA")
		panic("Da hell u doin? too much or not enough secrets")
	}

	return secrets
}


var G_csrf_secret string

func SetCsrfSecret(secret string) {
	if len(secret) < 32 {
		panic("ADEKHI")
	}

	G_csrf_secret = secret
}

func GetCsrfSecret() string {
	return G_csrf_secret
}


var G_session_secret string

func SetSessionSecret(secret string) {
	if len(secret) < 32 {
		panic("AILUDA")
	}

	G_session_secret = secret
}

func GetSessionSecret() string {
	return G_session_secret
}


var G_domain string

func SetDomain(domain string) {
	G_domain = domain
}

func GetDomain() string {
	return G_domain
}


var G_useTmplCache bool = true

func SetUseTemplateCache(val bool) {
	G_useTmplCache = val
}

func GetUseTemplateCache() bool {
	return G_useTmplCache
}


var G_allowRegistration bool = false

func SetAllowRegistration(val bool) {
	G_allowRegistration = val
}

func GetAllowRegistration() bool {
	return G_allowRegistration
}
