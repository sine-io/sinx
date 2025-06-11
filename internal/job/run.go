package job

// Run the job
func (j *Job) Run() {
	// As this function should comply with the Job interface of the cron package we will use
	// the agent property on execution, this is why it need to check if it's set and otherwise fail.
	if j.Agent == nil {
		j.Agent.Logger.Fatal().Msg("job: agent not set")
	}

	// Check if it's runnable
	if j.isRunnable() {
		j.logger.Debug().
			Str("job", j.Name).
			Str("schedule", j.Schedule).
			Msg("job: Running job")

		cronInspect.Set(j.Name, j)

		// Simple execution wrapper
		ex := NewExecution(j.Name)

		if _, err := j.Agent.Run(j.Name, ex); err != nil {
			j.logger.Error().
				Err(err).
				Msg("job: Error running job")
		}
	}
}
