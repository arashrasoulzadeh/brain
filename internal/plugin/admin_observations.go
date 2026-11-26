package plugin

type adminObservation struct {
	projectID string
	event     string
	sequence  uint64
}

// The admin invalidation feed retains only a bounded window of event identities.
// It never stores or returns observation payloads.
func (r *Registry) recordAdminObservation(projectID, event string) {
	r.adminObservationMu.Lock()
	defer r.adminObservationMu.Unlock()
	r.adminObservationSeq++
	sequence := r.adminObservationSeq
	r.adminObservations[(sequence-1)%uint64(len(r.adminObservations))] = adminObservation{projectID: projectID, event: event, sequence: sequence}
}

func (r *Registry) AdminObservationRevision(projectID string, events []string) uint64 {
	r.adminObservationMu.RLock()
	defer r.adminObservationMu.RUnlock()
	var latest uint64
	for _, observation := range r.adminObservations {
		if observation.projectID != projectID || observation.sequence <= latest {
			continue
		}
		for _, event := range events {
			if observation.event == event {
				latest = observation.sequence
				break
			}
		}
	}
	return latest
}
