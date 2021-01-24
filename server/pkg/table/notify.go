package table

import (
	"github.com/Zamiell/hanabi-live/server/pkg/types"
)

func (m *Manager) notifyTable() {
	// Local variables
	t := m.table

	if !t.Visible {
		return
	}

	tableDescription := newDescription(t)
	m.Dispatcher.Sessions.NotifyAllTable(tableDescription)
}

func (m *Manager) notifyAll(notifyFunc func(int)) {
	// Local variables
	t := m.table

	if !t.Replay {
		for _, p := range t.Players {
			if p.Present {
				notifyFunc(p.UserID)
			}
		}
	}

	for _, sp := range t.spectators {
		notifyFunc(sp.userID)
	}
}

// ------------------------------------------------
// Notifications before or after a game has started
// ------------------------------------------------

func (m *Manager) notifyChat(username string, msg string) {
	// Local variables
	t := m.table

	m.notifyAll(func(userID int) {
		m.Dispatcher.Sessions.NotifyChat(userID, username, msg, t.getRoomName())
	})
}

func (m *Manager) notifyStopTyping(username string) {
	// Local variables
	t := m.table

	m.notifyAll(func(userID int) {
		m.Dispatcher.Sessions.NotifyChatTyping(userID, t.ID, username, false)
	})
}

// ---------------------------------------
// Notifications before a game has started
// ---------------------------------------

// notifyPlayerChanged sends the people in the pre-game an update about the new amount of players.
// This is only called in situations where the game has not started yet.
func (m *Manager) notifyPlayerChanged() {
	// Local variables
	t := m.table

	if t.Running {
		m.logger.Error("The \"NotifyPlayerChange()\" method was called on a game that has already started.")
		return
	}

	for _, p := range t.Players {
		if !p.Present {
			continue
		}

		// First, make the array that contains information about all of the players in the game
		gamePlayers := make([]*types.GamePlayerData, 0)
		for j, p2 := range t.Players {
			gamePlayer := &types.GamePlayerData{
				Index:   j,
				Name:    p2.Username,
				You:     p.UserID == p2.UserID,
				Present: p2.Present,
				Stats:   p2.Stats,
			}
			gamePlayers = append(gamePlayers, gamePlayer)
		}

		// Second, send information about the game and the players in one big message
		m.Dispatcher.Sessions.NotifyGame(p.UserID, &types.GameData{
			TableID:           t.ID,
			Name:              t.Name,
			Owner:             t.OwnerID,
			Players:           gamePlayers,
			Options:           t.Options,
			PasswordProtected: t.PasswordHash != "",
		})
	}
}

// --------------------------------------
// Notifications after a game has started
// --------------------------------------

func (m *Manager) notifyGameAction() {
	// Local variables
	t := m.table
	g := t.Game

	if !t.Running {
		// We might be doing the initial actions;
		// don't send any messages to players if this is the case
		return
	}

	// Get the last action of the game
	lastAction := g.Actions[len(g.Actions)-1]

	for _, gp := range g.Players {
		p := t.Players[gp.Index]
		if p.Present {
			scrubbedAction := m.checkScrub(t, lastAction, p.UserID)
			m.Dispatcher.Sessions.NotifyGameAction(p.UserID, t.ID, scrubbedAction)
		}
	}

	for _, sp := range t.spectators {
		// We don't need to scrub any actions for spectators
		m.Dispatcher.Sessions.NotifyGameAction(sp.userID, t.ID, lastAction)
	}
}

func (m *Manager) notifyProgress() {
	// Local variables
	t := m.table

	if !t.Running {
		// We might be doing the initial actions;
		// don't send any messages to players if this is the case
		return
	}

	if !t.Visible {
		// Don't send progress for solo replays
		return
	}

	// Table progress only gets sent to interested users
	m.Dispatcher.Sessions.NotifyTableProgress(m.getNotifyUserIDs(), t.ID, t.Progress)
}

func (m *Manager) notifySpectatorsChanged() {
	// Local variables
	t := m.table

	if !t.Visible {
		return
	}

	spectators := make([]*types.SpectatorDescription, 0)
	for _, sp := range t.spectators {
		spectators = append(spectators, &types.SpectatorDescription{
			Username:             sp.username,
			ShadowingPlayerIndex: sp.shadowingPlayerIndex,
		})
	}

	m.notifyAll(func(userID int) {
		m.Dispatcher.Sessions.NotifySpectators(userID, t.ID, spectators)
	})
}

func (m *Manager) notifySpectatorsNote(order int) {
	// Local variables
	t := m.table
	g := t.Game

	for _, sp := range t.spectators {
		// Make an array that contains the combined notes for all the players & spectators
		// (for a specific card)
		// However, if this spectator is shadowing a specific player,
		// then only include the note for the shadowed player
		notes := make([]*types.Note, 0)
		for _, p := range g.Players {
			if sp.shadowingPlayerIndex == -1 || sp.shadowingPlayerIndex == p.Index {
				notes = append(notes, &types.Note{
					Name: p.Name,
					Text: p.Notes[order],
				})
			}
		}

		if sp.shadowingPlayerIndex == -1 {
			for _, sp2 := range t.spectators {
				notes = append(notes, &types.Note{
					Name: sp2.username,
					Text: sp2.notes[order],
				})
			}
		}

		m.Dispatcher.Sessions.NotifyNote(sp.userID, t.ID, order, notes)
	}
}

// -----------
// Subroutines
// -----------

func (m *Manager) getNotifyUserIDs() []int {
	// Local variables
	t := m.table

	userIDs := make([]int, 0)

	if !t.Replay {
		for _, p := range t.Players {
			userIDs = append(userIDs, p.UserID)
		}
	}

	for _, sp := range t.spectators {
		userIDs = append(userIDs, sp.userID)
	}

	return userIDs
}
