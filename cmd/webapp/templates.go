package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"maps"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jaitaiwan/haproxy-manager/internal/errors"
	"github.com/jaitaiwan/haproxy-manager/internal/util"
)

//go:embed static
var staticFiles embed.FS

// Taken from https://github.com/tarampampam/error-pages/blob/master/internal/template/template.go
var builtInFunctions = template.FuncMap{ //nolint:gochecknoglobals
	// the current time in unix format (seconds since 1970 UTC):
	//	`{{ nowUnix }}`	// `1631610000`
	"nowUnix": func() int64 { return time.Now().Unix() },

	// current hostname:
	//	`{{ hostname }}`	// `localhost`
	"hostname": func() string { h, _ := os.Hostname(); return h }, //nolint:nlreturn

	// json-serialized value (safe to use with any type):
	//	`{{ json "test" }}`	// `"test"`
	//	`{{ json 42 }}`	// `42`
	"json": func(v any) string { b, _ := json.Marshal(v); return string(b) }, //nolint:nlreturn,errchkjson

	// cast any type to int, or return 0 if it's not possible:
	//	`{{ int "42" }}`	// `42`
	//	`{{ int 42 }}`	// `42`
	//	`{{ int 3.14 }}`	// `3`
	//	`{{ int "test" }}`	// `0`
	//	`{{ int "42test" }}`	// `0`
	"int": func(v any) int { // cast any type to int, or return 0 if it's not possible
		switch v := v.(type) {
		case string:
			if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
				return i
			}
		case int:
			return v
		case int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			if i, err := strconv.Atoi(fmt.Sprintf("%d", v)); err == nil { // not effective, but safe
				return i
			}
		case float32, float64:
			if i, err := strconv.ParseFloat(fmt.Sprintf("%f", v), 32); err == nil { // not effective, but safe
				return int(i)
			}
		case fmt.Stringer:
			if i, err := strconv.Atoi(v.String()); err == nil {
				return i
			}
		}

		return 0
	},

	// current application version:
	//	`{{ version }}`	// `1.0.0`
	"version": func() string { return os.Getenv("VERSION") },

	// counts the number of non-overlapping instances of substr in s:
	//	`{{ strCount "test" "t" }}`	// `2`
	"strCount": strings.Count,

	// reports whether substr is within s:
	//	`{{ strContains "test" "es" }}`	// `true`
	//	`{{ strContains "test" "ez" }}`	// `false`
	"strContains": strings.Contains,

	// returns a slice of the string s, with all leading and trailing white space removed:
	//	`{{ strTrimSpace "  test  " }}`	// `test`
	"strTrimSpace": strings.TrimSpace,

	// returns s without the provided leading prefix string:
	//	`{{ strTrimPrefix "test" "te" }}`	// `st`
	"strTrimPrefix": strings.TrimPrefix,

	// returns s without the provided trailing suffix string:
	//	`{{ strTrimSuffix "test" "st" }}`	// `te`
	"strTrimSuffix": strings.TrimSuffix,

	// returns a copy of the string s with all non-overlapping instances of old replaced by new:
	//	`{{ strReplace "test" "t" "z" }}`	// `zesz`
	"strReplace": strings.ReplaceAll,

	// returns the index of the first instance of substr in s, or -1 if substr is not present in s:
	//	`{{ strIndex "barfoobaz" "foo" }}`	// `3`
	"strIndex": strings.Index,

	// splits the string s around each instance of one or more consecutive white space characters:
	//	`{{ strFields "foo bar baz" }}`	// `[foo bar baz]`
	"strFields": strings.Fields,

	// retrieves the value of the environment variable named by the key:
	//	`{{ env "SHELL" }}`	// `/bin/bash`
	"env": os.Getenv,

	// escapes special characters like "<" to become "&lt;":
	//	`{{ escape "<test>" }}`	// `&lt;test&gt;`
	"escape": html.EscapeString,

	"noescape": html.UnescapeString,
}

type DefaultViewData struct {
	Title  string
	Flags  *util.FeatureFlags
	Active string
}

func WriteTemplate(w http.ResponseWriter, tmpl *template.Template, data interface{}) {
	var body bytes.Buffer

	if err := tmpl.ExecuteTemplate(&body, "index.html", data); err != nil {
		WriteErrorTemplate(w, map[string]interface{}{
			"code":        http.StatusInternalServerError,
			"message":     "Could not render page",
			"description": err.Error(),
			"error":       err,
		})
		return
	}

	w.Header().Set("Content-Type", "text/html")
	io.Copy(w, &body)
}

func WriteErrorTemplate(w http.ResponseWriter, data map[string]interface{}) {
	var body bytes.Buffer

	// Data should have by minimum the following fields:
	// - code: int
	// - message: string
	// - description: string
	// - show_details : bool
	// -- Following are optional
	// - host: string
	// - original_uri: string
	// - forwarded_for: string
	// - namespace: string
	// - ingress_name: string
	// - service_name: string
	// - service_port: string
	// - request_id: string
	// - nowUnix: string

	var fns = maps.Clone(builtInFunctions)

	tmpl, err := template.New("error").Funcs(fns).ParseFS(f, "error.html")
	if err != nil {
		panic(err)
	}

	if data["show_details"] == nil {
		data["show_details"] = false
	}

	if os.Getenv("ENV") != PROD_ENV {
		data["show_errors"] = true
		data["show_details"] = true
		if data["error"] != nil {
			if stackErr, ok := data["error"].(errors.CustomError); ok {
				data["error"] = formatStackTrace(stackErr.StackTrace)
			}
		}

		if data["stack"] == nil {
			data["stack"] = formatStackTrace(errors.CaptureStackTrace(3))
		}
	} else {
		data["error"] = nil
	}

	if err := tmpl.ExecuteTemplate(&body, "error.html", data); err != nil {
		log.Printf("error rendering page: %v", err)
		http.Error(w, "Could not render page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(data["code"].(int))
	io.Copy(w, &body)
}

func formatStackTrace(raw string) template.HTML {
	escaped := template.HTMLEscapeString(raw)                               // Prevents XSS
	escaped = strings.ReplaceAll(escaped, " ", "&nbsp;")                    // Space to &nbsp;
	escaped = strings.ReplaceAll(escaped, "\t", "&nbsp;&nbsp;&nbsp;&nbsp;") // Tab to 4 spaces
	escaped = strings.ReplaceAll(escaped, "\n", "<br>")                     // Newlines to <br>
	return template.HTML(escaped)
}
