package job

func RecursiveSetJob(jobs []*Job) []string {
	result := make([]string, 0)
	for _, job := range jobs {
		err := job.Agent.GRPCClient.SetJob(job)
		if err != nil {
			result = append(result, "fail create "+job.Name)
			continue
		} else {
			result = append(result, "success create "+job.Name)
			if len(job.ChildJobs) > 0 {
				recursiveResult := RecursiveSetJob(job.ChildJobs)
				result = append(result, recursiveResult...)
			}
		}
	}
	return result
}
