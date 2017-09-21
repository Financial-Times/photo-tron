package main

import (
	"net/http"
	"os"

	api "github.com/Financial-Times/api-endpoint"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/photo-tron/annotations"
	"github.com/Financial-Times/photo-tron/fotoware"
	"github.com/Financial-Times/photo-tron/health"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/husobee/vestigo"
	"github.com/jawher/mow.cli"
	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

const appDescription = "PAC Draft Annotations API"

func main() {
	app := cli.App("photo-tron", appDescription)

	appSystemCode := app.String(cli.StringOpt{
		Name:   "app-system-code",
		Value:  "photo-tron",
		Desc:   "System Code of the application",
		EnvVar: "APP_SYSTEM_CODE",
	})

	appName := app.String(cli.StringOpt{
		Name:   "app-name",
		Value:  "photo-tron",
		Desc:   "Application name",
		EnvVar: "APP_NAME",
	})

	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})

	annotationsEndpoint := app.String(cli.StringOpt{
		Name:   "annotations-endpoint",
		Value:  "http://test.api.ft.com/content/%v/annotations",
		Desc:   "Endpoint to get annotations from UPP",
		EnvVar: "ANNOTATIONS_ENDPOINT",
	})

	uppAPIKey := app.String(cli.StringOpt{
		Name:   "upp-api-key",
		Value:  "",
		Desc:   "API key to access UPP",
		EnvVar: "UPP_APIKEY",
	})

	apiYml := app.String(cli.StringOpt{
		Name:   "api-yml",
		Value:  "./api.yml",
		Desc:   "Location of the API Swagger YML file.",
		EnvVar: "API_YML",
	})

	fotowareAPIKey := app.String(cli.StringOpt{
		Name:   "fotoware-api-key",
		Value:  "",
		Desc:   "",
		EnvVar: "FW_APIKEY",
	})

	log.SetLevel(log.InfoLevel)
	log.Infof("[Startup] %v is starting", *appSystemCode)

	app.Action = func() {
		log.Infof("System code: %s, App Name: %s, Port: %s", *appSystemCode, *appName, *port)

		fwAPI := fotoware.NewFotowareAPI(*fotowareAPIKey)
		annotationsAPI := annotations.NewAnnotationsAPI(*annotationsEndpoint, *uppAPIKey)
		annotationsHandler := annotations.NewHandler(annotationsAPI, fwAPI)
		healthService := health.NewHealthService(*appSystemCode, *appName, appDescription, annotationsAPI)

		serveEndpoints(*port, apiYml, annotationsHandler, healthService)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("App could not start, error=[%s]\n", err)
		return
	}
}

func serveEndpoints(port string, apiYml *string, handler *annotations.Handler, healthService *health.HealthService) {

	r := vestigo.NewRouter()
	r.Get("/photos-by-uuid/:uuid", handler.ServeHTTP)
	var monitoringRouter http.Handler = r
	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	http.HandleFunc("/__health", healthService.HealthCheckHandleFunc())
	http.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(healthService.GTG))
	http.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)

	http.Handle("/", monitoringRouter)

	if apiYml != nil {
		apiEndpoint, err := api.NewAPIEndpointForFile(*apiYml)
		if err != nil {
			log.WithError(err).WithField("file", *apiYml).Warn("Failed to serve the API Endpoint for this service. Please validate the Swagger YML and the file location")
		} else {
			r.Get(api.DefaultPath, apiEndpoint.ServeHTTP)
		}
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Unable to start: %v", err)
	}
}
