package rpc

// HealthCheckInput is the input struct for the health-check functionality
type HealthCheckInput struct {
	CommonInput
}

// HealthCheckOutput is the output struct for the health-check functionality
type HealthCheckOutput struct {
	Ok bool
}
