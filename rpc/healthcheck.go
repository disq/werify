package rpc

// Type HealthCheckInput is the input struct for the health-check functionality
type HealthCheckInput struct {
	CommonInput
}

// Type HealthCheckOutput is the output struct for the health-check functionality
type HealthCheckOutput struct {
	Ok bool
}
