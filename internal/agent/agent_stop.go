package agent

// StopAgent stops an agent, if the agent is a server and is running for election
// stop running for election, if this server was the leader
// this will force the cluster to elect a new leader and start a new scheduler.
// If this is a server and has the scheduler started stop it, ignoring if this server
// was participating in leader election or not (local storage).
// Then actually leave the cluster.
func StopAgent(a *Agent) error {
	a.Logger.Info().Msg("agent: Called member stop, now stopping")

	if a.Config.Server {
		if a.sched.Started() {
			<-a.sched.Stop().Done()
		}

		// TODO: Check why Shutdown().Error() is not working
		_ = a.raft.Shutdown()

		if err := a.Store.Shutdown(); err != nil {
			return err
		}
	}

	if err := a.Serf.Leave(); err != nil {
		return err
	}

	if err := a.Serf.Shutdown(); err != nil {
		return err
	}

	return nil
}
