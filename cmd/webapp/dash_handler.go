package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/haproxytech/client-native/v6/models"
	"github.com/jaitaiwan/haproxy-manager/internal/util"
	dpc "github.com/jaitaiwan/haproxy-manager/internal/util/dataplane-client"
)

func dashboardEndpoint(ff *util.FeatureFlags) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the index.html and login.partial.html templates from the embedded filesystem.
		tmpl, err := template.ParseFS(f, "index.html", "nav.partial.html", "dashboard/default.partial.html")
		if err != nil {
			http.Error(w, "Could not parse templates", http.StatusInternalServerError)
			return
		}

		// Execute the template and write the output to the response.
		w.Header().Set("Content-Type", "text/html")

		frontends, e, err := dpc.GetFrontends()
		if err != nil {
			http.Error(w, "Could not get frontends", http.StatusInternalServerError)
			return
		}

		if e != nil {
			log.Printf("Error: %v", e)
		}

		data := struct {
			DefaultViewData
			Frontends models.Frontends
		}{
			DefaultViewData: DefaultViewData{
				Title: "Dashboard",
				Flags: ff,
			},
			Frontends: frontends,
		}

		if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, "Could not render page", http.StatusInternalServerError)
		}
	})
}
