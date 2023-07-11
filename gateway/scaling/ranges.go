package scaling

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/openfaas/faas-provider/types"
)

const (
	// DefaultMinReplicas is the minimal amount of replicas for a service.
	DefaultMinReplicas = 1

	// DefaultMaxReplicas is the amount of replicas a service will auto-scale up to.
	DefaultMaxReplicas = 20

	// DefaultScalingFactor is the defining proportion for the scaling increments.
	DefaultScalingFactor = 10

	DefaultTypeScale = "rps"

	// MinScaleLabel label indicating min scale for a function
	MinScaleLabel = "com.openfaas.scale.min"

	// MaxScaleLabel label indicating max scale for a function
	MaxScaleLabel = "com.openfaas.scale.max"

	// ScalingFactorLabel label indicates the scaling factor for a function
	ScalingFactorLabel = "com.openfaas.scale.factor"
)

func MakeHorizontalScalingHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
			log.Print("Error cause method")
			return
		}

		if r.Body == nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			log.Printf("error cause empty body")
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			log.Print("error cause unable read request")
			return
		}

		scaleRequest := types.ScaleServiceRequest{}
		// log.Printf("Service %s in namespace %s have requested %d replicas", scaleRequest.ServiceName, scaleRequest.Namespace, scaleRequest.Replicas)
		if err := json.Unmarshal(body, &scaleRequest); err != nil {
			http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
			log.Print("error cause unmarshalling request body")
			return
		}

		if scaleRequest.Replicas < 1 {
			log.Printf("Service name %s want to scale to zero", scaleRequest.ServiceName)
			scaleRequest.Replicas = 0
		}

		if scaleRequest.Replicas > DefaultMaxReplicas {
			scaleRequest.Replicas = DefaultMaxReplicas
		}

		upstreamReq, _ := json.Marshal(scaleRequest)
		// Restore the io.ReadCloser to its original state
		r.Body = ioutil.NopCloser(bytes.NewBuffer(upstreamReq))

		next.ServeHTTP(w, r)
	}
}
