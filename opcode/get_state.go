package opcode

import (
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/relby/achikaps/match_state"
)

func GetStateHandler(dispatcher runtime.MatchDispatcher, msg runtime.MatchData, state *match_state.State) error {
	b, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("can't unmarshal state: %w", err)
	}
	if err := dispatcher.BroadcastMessage(int64(GetState), b, []runtime.Presence{state.Presences[msg.GetUserId()]}, nil, true); err != nil {
		return fmt.Errorf("can't broadcast message state: %w", err)
	}

	return nil
}