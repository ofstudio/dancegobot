package services

import (
	"sort"

	"github.com/ofstudio/dancegobot/internal/helpers"
	"github.com/ofstudio/dancegobot/internal/models"
)

// eventHandler implements the event handling logic and rules.
type eventHandler struct {
	event *models.Event
	hist  []*models.HistoryItem
	notif []*models.Notification
}

func newEventHandler(event *models.Event) *eventHandler {
	return &eventHandler{
		event: event,
	}
}

// Event returns the event being handled.
func (h *eventHandler) Event() *models.Event {
	return h.event
}

// History returns the list of history items to save
// that were collected during event handling.
func (h *eventHandler) History() []*models.HistoryItem {
	return h.hist
}

// Notifications returns the list of notifications to send
// that were collected during the event handling.
func (h *eventHandler) Notifications() []*models.Notification {
	return h.notif
}

// DancerGetByProfile looking for a dancer at event by provided Telegram profile.
// If not found, creates a new dancer with the provided profile and role.
// If found, Result and Partner fields will be filled accordingly.
func (h *eventHandler) DancerGetByProfile(profile *models.Profile, role models.Role) *models.Dancer {
	newDancer := &models.Dancer{
		Profile:   profile,
		FullName:  profile.FullName(),
		Role:      role,
		Status:    models.StatusNotRegistered,
		CreatedAt: nowFn(),
	}

	if found, ok := h.dancerGet(newDancer); ok {
		return found
	}
	return newDancer
}

// DancerGetByName looking for a dancer at event by provided full name.
// If not found, creates a new dancer with the provided name and role.
// If found, [models.Dancer.Status] and [models.Dancer.Partner] fields will be filled accordingly.
func (h *eventHandler) DancerGetByName(fullname string, role models.Role) *models.Dancer {
	newDancer := &models.Dancer{
		FullName:  fullname,
		Role:      role,
		Status:    models.StatusNotRegistered,
		CreatedAt: nowFn(),
	}
	if found, ok := h.dancerGet(newDancer); ok {
		return found
	}
	return newDancer
}

// dancerGet looking for a dancer at event.
// If not found, returns the provided dancer as is with [models.Dancer.Status] set to [models.StatusNotRegistered].
// If found, [models.Dancer.Status] and [models.Dancer.Partner] fields will be filled accordingly.
func (h *eventHandler) dancerGet(dancer *models.Dancer) (*models.Dancer, bool) {
	if foundDancer, ok := h.findInCouples(dancer); ok {
		foundDancer.Status = models.StatusInCouple
		return foundDancer, ok
	}
	if foundDancer, ok := h.findInSingles(dancer); ok {
		foundDancer.Status = models.StatusSingle
		return foundDancer, ok
	}

	// If not found, return the provided dancer as is with Status set to NotRegistered
	dancer.Status = models.StatusNotRegistered
	return dancer, false
}

// CoupleAdd registers a couple for the event.
//
// If the partner initially was registered as a single, the partner will be notified.
//
// Note: both dancer and partner should be with actual status and partner fields.
func (h *eventHandler) CoupleAdd(dancer, partner *models.Dancer) *models.EventUpdate {
	var result models.UpdateResult

	// 1. Check if event is not forbidden for the dancer or partner
	if dancer.Status == models.StatusForbidden {
		result = models.ResultEventForbiddenDancer
	}
	if partner.Status == models.StatusForbidden {
		result = models.ResultEventForbiddenPartner
	}

	// 2. Check if the partner is already registered in a couple
	if partner.Status == models.StatusInCouple {
		result = models.ResultPartnerTaken
	}

	// 3. Check if the dancer is already registered in a couple
	if dancer.Status == models.StatusInCouple {
		if h.isSame(dancer.Partner, partner) {
			result = models.ResultAlreadyInSameCouple
		} else {
			result = models.ResultAlreadyInCouple
		}
	}

	// 4. Check if event is not closed for new registrations
	if h.event.Closed {
		result = models.ResultEventClosed
	}

	// 5. Check if the partner has the same role as the dancer
	if dancer.Role == partner.Role {
		result = models.ResultPartnerSameRole
	}

	// 6. Check if the dancer is trying to register with itself
	if h.isSame(dancer, partner) {
		result = models.ResultSelfNotAllowed
	}

	// 7. Break if any of the checks failed
	if result != models.ResultUnknown {
		return &models.EventUpdate{
			Event:         h.event,
			Result:        result,
			Dancer:        dancer,
			ChosenPartner: partner,
		}
	}

	// 8. Check if dancer is in singles and remove from singles
	if dancer.Status == models.StatusSingle {
		h.removeFromSingles(dancer)
		h.hist = append(h.hist, &models.HistoryItem{
			Action:    models.HistorySingleRemoved,
			Initiator: dancer.Profile,
			EventID:   &h.event.ID,
			Details:   dancer,
			CreatedAt: nowFn(),
		})
	}

	// 9. Check if partner is in singles and remove from singles
	// and create notification for the partner
	if partner.Status == models.StatusSingle {
		h.removeFromSingles(partner)
		h.hist = append(h.hist, &models.HistoryItem{
			Action:    models.HistorySingleRemoved,
			Initiator: dancer.Profile,
			EventID:   &h.event.ID,
			Details:   partner,
			CreatedAt: nowFn(),
		})
		h.notif = append(h.notif, &models.Notification{
			Recipient: *partner.Profile,
			Initiator: dancer.Profile,
			TmplCode:  models.TmplRegisteredWithSingle,
			Event:     h.event,
		})
	}

	// 10. Set the dancers as a couple
	dancer.Status = models.StatusInCouple
	dancer.Partner = partner
	partner.Status = models.StatusInCouple
	partner.Partner = dancer

	// 11. Create a couple and add to the event
	couple := models.Couple{
		CreatedBy: *dancer.Profile,
		CreatedAt: nowFn(),
	}
	if dancer.Role == models.RoleLeader {
		couple.Dancers = []models.Dancer{*dancer, *partner}
	} else {
		couple.Dancers = []models.Dancer{*partner, *dancer}
	}
	h.event.Couples = append(h.event.Couples, couple)
	h.hist = append(h.hist, &models.HistoryItem{
		Action:    models.HistoryCoupleAdded,
		Initiator: dancer.Profile,
		EventID:   &h.event.ID,
		Details:   &couple,
		CreatedAt: nowFn(),
	})

	// 12. Return the update result
	return &models.EventUpdate{
		Event:         h.event,
		Result:        models.ResultSuccess,
		Dancer:        dancer,
		ChosenPartner: partner,
		Couple:        &couple,
	}
}

// SingleAdd registers a dancer as a single for the event.
// Note: dancer should be with actual status and partner fields.
func (h *eventHandler) SingleAdd(dancer *models.Dancer) *models.EventUpdate {
	var result models.UpdateResult

	// 1. Check if event is not forbidden for the dancer
	if dancer.Status == models.StatusForbidden {
		result = models.ResultEventForbiddenDancer
	}

	// 2. Check if the dancer is already registered in a couple
	if dancer.Status == models.StatusInCouple {
		result = models.ResultAlreadyInCouple
	}

	// 3. Check if dancer is already registered as single
	if dancer.Status == models.StatusSingle {
		result = models.ResultAlreadyAsSingle
	}

	// 4. Check if event is not closed for new registrations
	if h.event.Closed {
		result = models.ResultEventClosed
	}

	/// 5. Break if any of the checks failed
	if result != models.ResultUnknown {
		return &models.EventUpdate{Event: h.event, Result: result, Dancer: dancer}
	}

	// todo signup as a couple with the first dancer in the singles list of opposite role if any
	// todo add tests for this case

	// 6. Set the dancer as a single
	dancer.SingleSignup = true
	dancer.Status = models.StatusSingle

	// 7. Create a single and add to the event
	h.event.Singles = append(h.event.Singles, *dancer)
	h.hist = append(h.hist, &models.HistoryItem{
		Action:    models.HistorySingleAdded,
		Initiator: dancer.Profile,
		EventID:   &h.event.ID,
		Details:   dancer,
		CreatedAt: nowFn(),
	})

	// 8. Return the update result
	return &models.EventUpdate{Event: h.event, Result: models.ResultSuccess, Dancer: dancer}
}

// DancerRemove removes the dancer from the event.
//
// If the dancer is in a couple, and the partner initially signed up as a single,
// the partner will be moved to the singles list back and a notification will be created.
// Otherwise, the partner will be removed from the event as well.
//
// If the dancer is in a couple, and a couple was created by the partner,
// the partner will be notified that the dancer has left the event.
//
// Note: dancer should be with actual status and partner fields.
func (h *eventHandler) DancerRemove(dancer *models.Dancer) *models.EventUpdate {

	// 1. Check if dancer is registered for the event at all
	if dancer.Status != models.StatusSingle && dancer.Status != models.StatusInCouple {
		return &models.EventUpdate{Event: h.event, Result: models.ResultNotRegistered, Dancer: dancer}
	}

	// 2. Check if dancer is in a singles list and remove from singles
	if dancer.Status == models.StatusSingle {
		h.removeFromSingles(dancer)
		h.hist = append(h.hist, &models.HistoryItem{
			Action:    models.HistorySingleRemoved,
			Initiator: dancer.Profile,
			EventID:   &h.event.ID,
			Details:   dancer,
			CreatedAt: nowFn(),
		})
		dancer.Status = models.StatusNotRegistered
		return &models.EventUpdate{Event: h.event, Result: models.ResultSuccess, Dancer: dancer}
	}

	// 3. If dancer is in a couple remove the couple
	couple, ok := h.removeCouple(dancer)
	if !ok {
		return &models.EventUpdate{Event: h.event, Result: models.ResultNotRegistered, Dancer: dancer}
	}
	h.hist = append(h.hist, &models.HistoryItem{
		Action:    models.HistoryCoupleRemoved,
		Initiator: dancer.Profile,
		EventID:   &h.event.ID,
		Details:   couple,
		CreatedAt: nowFn(),
	})

	// 3.1 Set dancer and partner status to not registered
	partner := dancer.Partner

	dancer.Status = models.StatusNotRegistered
	dancer.Partner = nil
	partner.Status = models.StatusNotRegistered
	partner.Partner = nil

	// 3.2 If partner was signed up as a single, move back to singles
	// todo check if any single dancer available in opposite role and register a partner in a couple with
	// todo add tests for this case
	if partner.SingleSignup {
		partner.Status = models.StatusSingle
		h.event.Singles = append(h.event.Singles, *partner)
		sort.Sort(SinglesSorter(h.event.Singles))
		h.hist = append(h.hist, &models.HistoryItem{
			Action:    models.HistorySingleAdded,
			Initiator: dancer.Profile,
			EventID:   &h.event.ID,
			Details:   partner,
			CreatedAt: nowFn(),
		})
	}

	if partner.Profile != nil {
		// 3.3 Send notification to the partner if needed
		if partner.SingleSignup {
			// 3.3.1 If partner was signed up as a single send the partner
			h.notif = append(h.notif, &models.Notification{
				Recipient: *partner.Profile,
				Initiator: dancer.Profile,
				TmplCode:  models.TmplCanceledWithSingle,
				Event:     h.event,
			})
		} else if couple.CreatedBy.ID == partner.Profile.ID {
			// 3.3.2 Otherwise if couple was created by the partner send the partner
			h.notif = append(h.notif, &models.Notification{
				Recipient: *partner.Profile,
				Initiator: dancer.Profile,
				TmplCode:  models.TmplCanceledByPartner,
				Event:     h.event,
			})
		}
	}

	// 4. Return the update result
	return &models.EventUpdate{Event: h.event, Result: models.ResultSuccess, Dancer: dancer}

}

// findInCouples finds dancer in the couples.
// Returns the dancer and true if found, otherwise nil and false.
func (h *eventHandler) findInCouples(dancer *models.Dancer) (*models.Dancer, bool) {
	var foundDancer models.Dancer
	for _, couple := range h.event.Couples {
		if h.isSame(dancer, &couple.Dancers[0]) {
			foundDancer = couple.Dancers[0]
			foundDancer.Partner = &couple.Dancers[1]
			return &foundDancer, true
		}
		if h.isSame(dancer, &couple.Dancers[1]) {
			foundDancer = couple.Dancers[1]
			foundDancer.Partner = &couple.Dancers[0]
			return &foundDancer, true
		}
	}
	return nil, false
}

// findInSingles finds dancer in the singles of the event.
// Returns the dancer and true if found, otherwise nil and false.
func (h *eventHandler) findInSingles(dancer *models.Dancer) (*models.Dancer, bool) {
	for _, single := range h.event.Singles {
		if h.isSame(dancer, &single) {
			return &single, true
		}
	}
	return nil, false
}

// removeFromSingles removes the dancer from the singles list of the event.
// If dancer found returns the dancer and true, otherwise nil and false.
func (h *eventHandler) removeFromSingles(dancer *models.Dancer) (*models.Dancer, bool) {
	for i, single := range h.event.Singles {
		if h.isSame(dancer, &single) {
			h.event.Singles = append(h.event.Singles[:i], h.event.Singles[i+1:]...)
			return &single, true
		}
	}
	return nil, false
}

// removeCouple removes the couple from the couples list of the event.
// If couple with the dancer found returns the couple and true, otherwise nil and false.
func (h *eventHandler) removeCouple(dancer *models.Dancer) (*models.Couple, bool) {
	for i, couple := range h.event.Couples {
		if h.isSame(dancer, &couple.Dancers[0]) || h.isSame(dancer, &couple.Dancers[1]) {
			h.event.Couples = append(h.event.Couples[:i], h.event.Couples[i+1:]...)
			return &couple, true
		}
	}
	return nil, false
}

// isSame checks if dancer is the same as the other dancer on the event
func (h *eventHandler) isSame(dancer, other *models.Dancer) bool {
	switch {
	// Compare profile IDs if both profiles are present
	case dancer.Profile != nil && other.Profile != nil:
		return dancer.ID == other.Profile.ID
	// Compare dancer username (if present in profile ) and other username (if present in full name)
	case dancer.Profile != nil && dancer.Profile.Username != "" && other.Profile == nil:
		username, ok := helpers.Username(other.FullName)
		return ok && (dancer.Profile.Username == username)
	// Compare dancer username (if present in full name) and other username (if present in profile)
	case dancer.Profile == nil && other.Profile != nil && other.Profile.Username != "":
		username, ok := helpers.Username(dancer.FullName)
		return ok && (username == other.Profile.Username)
	// Compare usernames (if present in full names) if both profiles are missing
	case dancer.Profile == nil && other.Profile == nil:
		username1, ok1 := helpers.Username(dancer.FullName)
		username2, ok2 := helpers.Username(other.FullName)
		return (ok1 && ok2) && (username1 == username2)
	default:
		return false
	}
}

// SinglesSorter is a sorter for singles by creation time.
type SinglesSorter []models.Dancer

func (s SinglesSorter) Len() int           { return len(s) }
func (s SinglesSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SinglesSorter) Less(i, j int) bool { return s[i].CreatedAt.Before(s[j].CreatedAt) }
