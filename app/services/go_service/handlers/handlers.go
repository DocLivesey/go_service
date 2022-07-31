//Package handlers contains full set of handlers and routes suported by web api
package handlers

import (
	"expvar"
	"net/http"
	"net/http/pprof"
)

//DebugStandartLibraryMux register all debuging routes of standart library into a new mux bypassing
//the use of the DefaultServerMux. For better security, stability and control
func DebugStandartLibraryMux() *http.ServeMux {

	mux := http.NewServeMux()

	//Registration of standart library debuging endpoints
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}
