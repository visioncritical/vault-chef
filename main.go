package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

// Setup

var cfg = ParseConfig()
var clients = SetupClients()

// HTTP Methods & Routes

// Authenticate protects the endpoints from requests with invalid signatures
func Authenticate(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		client, err := clients.Chef.GetClient(r.Header.Get("X-VAULT-CLIENT"))
		if err != nil {
			log.Printf("[INFO] Failed to find Chef client for %s, error: %s", r.Header.Get("X-VAULT-CLIENT"), err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		err = VerifySignature(r.Header.Get("X-VAULT-SIGNATURE"), client.PublicKey)
		if err != nil {
			log.Printf("[INFO] Failed to verify signature for Chef client %s, error: %s", r.Header.Get("X-VAULT-CLIENT"), err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			log.Printf("[INFO] Successfully authenticated %s", r.Header.Get("X-VAULT-CLIENT"))
			// Delegate request to the given handle
			h(w, r, ps)
		}
	}
}

// TokenRole Issues a Role Token
func TokenRole(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	tcr := &api.TokenCreateRequest{}
	err := json.NewDecoder(r.Body).Decode(&tcr)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	token, err := clients.Vault.IssueTokenRole(tcr, ps.ByName("role"))
	if err != nil {
		http.Error(w, "Failed to create token from role", http.StatusBadRequest)
		log.Printf("Error: %s", err)
		return
	}
	fmt.Fprint(w, token)
}

// Main

func main() {
	srvCert, err := clients.Vault.ReadSecretPath(cfg.SecretCertPath)
	if err != nil {
		log.Fatalf("Failed to get certificates from Vault: %s", err)
	}
	cert, err := ParseX509KeyPair(srvCert.Data["cert"].(string), srvCert.Data["key"].(string))
	if err != nil {
		log.Fatalf("Failed to parse certifcates: %s", err)
	}

	if cfg.RenewToken {
		go clients.Vault.RenewToken()
	}

	router := httprouter.New()
	router.POST("/token/issue/:role", Authenticate(TokenRole))

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())
	n.UseHandler(router)

	s := &http.Server{
		Addr:           cfg.ListenAddr,
		TLSConfig:      TLSConfig(cert),
		Handler:        n,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Certs are specified in the tls.Config struct
	log.Println("[INFO] Starting vault-chef server, listening on", cfg.ListenAddr)
	log.Fatal(s.ListenAndServeTLS("", ""))
}
