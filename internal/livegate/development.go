package livegate

// ProjectDevelopmentQRCommand exposes closed command classification before the
// formal QR/status pairing. A QR pass here proves only the child's zero exit.
func ProjectDevelopmentQRCommand(action CommandAction, observation Observation, duration int64) (Step, error) {
	for _, spec := range append(append([]actionStep{}, terminalQRPrimaryActions[:]...), cleanupActions[:]...) {
		if spec.action != action {
			continue
		}
		if observation.RunnerFault {
			return Step{}, ErrCommandRunnerFault
		}
		if action == ActionQRLogin && observation.ExitCode == CommandExitSuccess {
			if observation.ErrorCode != nil || observation.Session != "" || observation.Outcome != "" || observation.IdentityMatches {
				return Step{}, ErrInvalidObservation
			}
			return Step{Name: StepQRLoginLoggedIn, Result: ResultPass, ExitCode: CommandExitSuccess, DurationSeconds: duration}, nil
		}
		step, err := stepFromObservation(spec, observation, duration)
		if action == ActionStatusOnline && observation.ExitCode == CommandExitSuccess && observation.Session == SessionOnline && !observation.IdentityMatches {
			step.Result = ResultBlocked
		}
		return step, err
	}
	return Step{}, ErrInvalidObservation
}

func ValidateDevelopmentQRSteps(steps []Step, result Result) error {
	if !result.valid() {
		return ErrInvalidEvidence
	}
	return validateStepSequence(SuiteTerminalQR, result, steps)
}

// ValidateDevelopmentPasswordSteps shares only the closed step semantics.
// It does not validate a Candidate or produce acceptable release evidence.
func ValidateDevelopmentPasswordSteps(steps []Step, result Result) error {
	if !result.valid() {
		return ErrInvalidEvidence
	}
	return validateStepSequence(SuitePasswordCore, result, steps)
}

// A canceled/faulted state machine may omit unobserved steps. This validates
// only its ordered observations, never promoting a partial run to evidence.
func ValidateDevelopmentPasswordPrefix(steps []Step) error {
	if len(steps) > 9 {
		return ErrInvalidEvidence
	}
	last := -1
	names := append(append([]StepName{}, passwordPrimarySteps[:]...), StepCleanupLogout, StepCleanupStatusOffline)
	for _, step := range steps {
		if step.Validate() != nil {
			return ErrInvalidEvidence
		}
		index := -1
		for i, name := range names {
			if step.Name == name {
				index = i
				break
			}
		}
		if index <= last || index < 0 || index < 7 && index != last+1 {
			return ErrInvalidEvidence
		}
		if index >= 7 && (!primaryStepPassed(steps, StepStatusOnline) || primaryStepPassed(steps, StepFinalStatusOffline)) {
			return ErrInvalidEvidence
		}
		last = index
	}
	return nil
}
