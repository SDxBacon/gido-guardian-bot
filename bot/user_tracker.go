package bot

import (
	"fmt"
	"sync"

	"github.com/SDxBacon/gido-guardian-bot/gido"
)

var userTicketTrackersMap = map[string]*gido.TicketTracker{}

var mutex = &sync.Mutex{}

// GetUserTicketTracker retrieves the ticket tracker for a specific user.
// It safely accesses the shared user ticket trackers map with mutex protection.
//
// Parameters:
//   - userID: The unique identifier of the user whose ticket tracker is being retrieved.
//
// Returns:
//   - *gido.TicketTracker: The user's ticket tracker if found, or nil if no tracker exists for the user.
func GetUserTicketTracker(userID string) *gido.TicketTracker {
	mutex.Lock()
	defer mutex.Unlock()

	if tracker, exists := userTicketTrackersMap[userID]; exists {
		return tracker
	}
	return nil
}

// CreateUserTicketTracker creates a new ticket tracker for a specific user.
// It takes the Discord user ID, ticket number to track, and optional configuration options.
// If a tracker already exists for the specified user, it returns an error.
// The function is thread-safe as it uses a mutex to protect access to the shared tracker map.
//
// Parameters:
//   - userID: The Discord user ID as a string
//   - ticketNumber: The ticket number to track
//   - opts: Optional configuration options for the ticket tracker
//
// Returns:
//   - *gido.TicketTracker: The newly created ticket tracker, or nil if an error occurred
//   - error: An error if the user already has a tracker, nil otherwise
func CreateUserTicketTracker(userID string, ticketNumber int, opts ...gido.TicketTrackerOption) (*gido.TicketTracker, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if tickerTracker, exists := userTicketTrackersMap[userID]; exists {

		return nil, fmt.Errorf("tracker for user <@%s> already exists (tracking: %d)", userID, tickerTracker.GetTrackingTicketId())
	}

	tracker := gido.NewTicketTracker(ticketNumber, opts...)
	userTicketTrackersMap[userID] = tracker

	return tracker, nil
}

func RemoveUserTicketTracker(userID string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(userTicketTrackersMap, userID)
}
