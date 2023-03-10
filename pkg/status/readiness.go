package status

// ReadinessStatus defines if the service is ready to serve.
var ReadinessStatus bool //nolint:gochecknoglobals

// GlobalReadinessStatus return the value of `ReadinessStatus`.
func GlobalReadinessStatus() bool {
	return ReadinessStatus
}
