package executor

type nativeTurn struct {
	Session string `json:"sessionId"`
	Turn    string `json:"turnId"`
}

func recordNativeTurn(path string) func(string, string) error {
	return func(session, turn string) error {
		return save(path, nativeTurn{Session: session, Turn: turn})
	}
}
