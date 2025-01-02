package services

import (
	"fmt"
	"sort"
	"time"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
)

// EventHandler implements the event logic and rules.
type EventHandler struct {
	event *models.Event
	hist  []*models.HistoryItem
	notif []*models.Notification
}

func NewEventHandler(event *models.Event) *EventHandler {
	return &EventHandler{
		event: event,
	}
}

// Event returns the event being handled.
func (h *EventHandler) Event() *models.Event {
	return h.event
}

// History returns the list of history items to save
// that were collected during event handling.
func (h *EventHandler) History() []*models.HistoryItem {
	return h.hist
}

// Notifications returns the list of notifications to send
// that were collected during the event handling.
func (h *EventHandler) Notifications() []*models.Notification {
	return h.notif
}

// CanManage checks if the profile can manage the event.
// Returns true if the profile is the owner of the event.
func (h *EventHandler) CanManage(profile *models.Profile) bool {
	return profile != nil && profile.ID == h.event.Owner.ID
}

// UpdateSettings updates the event settings.
func (h *EventHandler) UpdateSettings(initiator *models.Profile, settings models.EventSettings) error {
	if !h.CanManage(initiator) {
		return fmt.Errorf("profile is not allowed to update the event settings")
	}
	h.event.Settings = settings
	h.hist = append(h.hist, &models.HistoryItem{
		Action:    models.HistoryEventSettingsUpdated,
		Initiator: initiator,
		EventID:   &h.event.ID,
		Details:   settings,
		CreatedAt: nowFn(),
	})
	return nil
}

// RegistrationGet returns registration for given dancer at the event.
// If the dancer is not registered, returns a new registration.
func (h *EventHandler) RegistrationGet(dancer *models.Dancer) *models.Registration {
	if existingReg := h.findInCouples(dancer); existingReg != nil {
		return existingReg
	}
	if existingReg := h.findInSingles(dancer); existingReg != nil {
		return existingReg
	}
	dancer.CreatedAt = nowFn()
	return &models.Registration{
		Dancer: dancer,
		Status: models.StatusNotRegistered,
		Event:  h.event,
	}
}

// LimitChangeAffected returns list of affected couples after the limit change.
// Returns true as second argument if the limit was increased, otherwise false.
// Returns the start index of the affected couples as the third argument.
func (h *EventHandler) LimitChangeAffected(oldLimit int) ([]models.Couple, bool, int) {
	start, end, increased := limitChangeRange(oldLimit, h.event.Settings.Limit, len(h.event.Couples))
	return h.event.Couples[start:end], increased, start
}

// LimitChangeNotify notifies the affected dancers about the event limit change.
func (h *EventHandler) LimitChangeNotify(oldLimit int) {
	// Get the affected couples
	from, to, increased := limitChangeRange(oldLimit, h.event.Settings.Limit, len(h.event.Couples))
	couples := h.event.Couples[from:to]

	// Select the template code
	tmplCode := models.TmplEventLimitDecreased
	if increased {
		tmplCode = models.TmplEventLimitIncreased
	}

	// Iterate over the couples and create notifications
	for _, couple := range couples {
		// Iterate over the dancers in the couple
		for i, dancer := range couple.Dancers {
			// Notify only the couple creator or who signed up as a single
			if dancer.Profile != nil &&
				(dancer.Profile.ID == couple.CreatedBy.ID || dancer.AsSingle) {
				partner := couple.Dancers[i^1]
				h.notif = append(h.notif, &models.Notification{
					TmplCode:  tmplCode,
					Recipient: dancer.Profile,
					Payload: models.NotificationPayload{
						Event:   h.event,
						Partner: &partner,
					},
				})
			}
		}
	}
}

// CoupleAdd registers a couple for the event.
// If the partner initially was registered as a single, the partner will be notified.
func (h *EventHandler) CoupleAdd(d, p *models.Dancer) *models.Registration {
	reg := h.RegistrationGet(d)
	reg.Result = models.ResultNoResult
	reg.Related = h.RegistrationGet(p)

	// Check if event not removed
	if h.event.Removed {
		reg.Result = models.ResultEventRemoved
		return reg
	}

	// Check if event is not closed for new registrations
	if h.event.Settings.Closed {
		reg.Result = models.ResultEventClosed
		return reg
	}

	// Check if event is not forbidden for the dancer or partner
	if reg.Status == models.StatusForbidden {
		reg.Result = models.ResultDancerForbidden
		return reg
	}
	if reg.Related.Status == models.StatusForbidden {
		reg.Result = models.ResultPartnerForbidden
		return reg
	}

	// Check if the dancer is already registered in a couple
	if reg.Status == models.StatusInCouple {
		if h.isSame(reg.Partner, reg.Related.Dancer) {
			reg.Result = models.ResultAlreadyInSameCouple
		} else {
			reg.Result = models.ResultAlreadyInCouple
		}
		return reg
	}

	// Check if the partner is already registered in a couple
	if reg.Related.Status == models.StatusInCouple {
		reg.Result = models.ResultPartnerTaken
		return reg
	}

	// Check if the dancer is trying to register with itself
	if h.isSame(reg.Dancer, reg.Related.Dancer) {
		reg.Result = models.ResultSelfNotAllowed
		return reg
	}

	// Check if the partner has the same role as the partner
	if reg.Role == reg.Related.Role {
		reg.Result = models.ResultPartnerSameRole
		return reg
	}

	// Register as couple
	return h.coupleAdd(reg, false, nowFn())
}

// coupleAdd processes the couple registration.
// The autoPair flag indicates if the couple was auto paired.
// The createdAt time is used for keeping couples order when partner changes.
func (h *EventHandler) coupleAdd(reg *models.Registration, autoPair bool, createdAt time.Time) *models.Registration {
	// Check if dancer is in singles and remove from singles
	if reg.Status == models.StatusAsSingle {
		h.removeFromSingles(reg.Dancer)
		h.hist = append(h.hist, &models.HistoryItem{
			Action:    models.HistorySingleRemoved,
			Initiator: reg.Dancer.Profile,
			EventID:   &h.event.ID,
			Details:   reg.Dancer,
			CreatedAt: nowFn(),
		})
	}

	// Check if partner is in singles and remove from singles
	// and create notification for the partner
	initiator := reg.Profile
	if autoPair {
		initiator = config.BotProfile()
	}
	if reg.Related.Status == models.StatusAsSingle {
		h.removeFromSingles(reg.Related.Dancer)

		// Add history item
		h.hist = append(h.hist, &models.HistoryItem{
			Action:    models.HistorySingleRemoved,
			Initiator: initiator,
			EventID:   &h.event.ID,
			Details:   reg.Related.Dancer,
			CreatedAt: nowFn(),
		})

		// Add notification for the partner
		tmplCode := models.TmplRegisteredWithSingle
		if autoPair {
			tmplCode = models.TmplAutoPairPartnerFound
		}
		h.notif = append(h.notif, &models.Notification{
			TmplCode:  tmplCode,
			Recipient: reg.Related.Profile,
			Payload: models.NotificationPayload{
				Event:   h.event,
				Partner: reg.Dancer,
			},
		})
	}

	// Create a couple
	couple := models.Couple{
		CreatedBy: *reg.Profile,
		AutoPair:  autoPair,
		CreatedAt: createdAt, // Use provided time to keep couples order
	}
	if reg.Role == models.RoleLeader {
		couple.Dancers = []models.Dancer{*reg.Dancer, *reg.Related.Dancer}
	} else {
		couple.Dancers = []models.Dancer{*reg.Related.Dancer, *reg.Dancer}
	}

	// Add couple to the event history
	h.hist = append(h.hist, &models.HistoryItem{
		Action:    models.HistoryCoupleAdded,
		Initiator: initiator,
		EventID:   &h.event.ID,
		Details:   &couple,
		CreatedAt: nowFn(),
	})

	// Add couple to the event
	h.event.Couples = append(h.event.Couples, couple)

	// Keep couples ordered by creation time
	sort.Sort(CouplesSorter(h.event.Couples))

	// Return the registration
	reg.Result = models.ResultRegisteredInCouple
	reg.Status = models.StatusInCouple
	reg.Partner = reg.Related.Dancer
	reg.Related.Result = models.ResultRegisteredInCouple
	reg.Related.Status = models.StatusInCouple
	reg.Related.Partner = reg.Dancer
	return reg
}

// SingleAdd registers a dancer as a single for the event.
// If auto pairing is enabled, tries to auto pair the dancer.
func (h *EventHandler) SingleAdd(d *models.Dancer) *models.Registration {

	reg := h.RegistrationGet(d)
	reg.Result = models.ResultNoResult

	// Check if event not removed
	if h.event.Removed {
		reg.Result = models.ResultEventRemoved
		return reg
	}

	// Check if event is not forbidden for the dancer
	if reg.Status == models.StatusForbidden {
		reg.Result = models.ResultDancerForbidden
		return reg
	}

	// Check if the dancer not already registered in a couple
	if reg.Status == models.StatusInCouple {
		reg.Result = models.ResultAlreadyInCouple
		return reg
	}

	// Check if dancer not already registered as single
	if reg.Status == models.StatusAsSingle {
		reg.Result = models.ResultAlreadyAsSingle
		return reg
	}

	// Check if event is not closed for new registrations
	if h.event.Settings.Closed {
		reg.Result = models.ResultEventClosed
		return reg
	}

	// Try to auto pair the reg if possible
	// This should be done before checking if singles are allowed
	if autoPairReg := h.tryAutoPair(reg, nowFn()); autoPairReg != nil {
		return autoPairReg
	}

	// Create a single and add to the event
	reg.Dancer.AsSingle = true
	h.event.Singles = append(h.event.Singles, *reg.Dancer)
	h.hist = append(h.hist, &models.HistoryItem{
		Action:    models.HistorySingleAdded,
		Initiator: reg.Profile,
		EventID:   &h.event.ID,
		Details:   reg.Dancer,
		CreatedAt: nowFn(),
	})

	// Return the update result
	reg.Status = models.StatusAsSingle
	reg.Result = models.ResultRegisteredAsSingle
	reg.Partner = nil
	reg.Related = nil
	return reg
}

// DancerRemove removes the dancer from the event.
//
// If the dancer is in a couple, and the partner initially signed up as a single,
// the partner will be moved to the singles list back (or auto paired if enabled)
// and a notification will be created.
// Otherwise, the partner will be removed from the event as well.
//
// If the dancer is in a couple, and a couple was created by the partner,
// the partner will be notified that the dancer has left the event.
func (h *EventHandler) DancerRemove(d *models.Dancer) *models.Registration {
	reg := h.RegistrationGet(d)

	// Check if event not removed
	if h.event.Removed {
		reg.Result = models.ResultEventRemoved
		return reg
	}

	// Check if event is closed for new registrations
	if h.event.Settings.Closed {
		reg.Result = models.ResultEventClosed
		return reg
	}

	// Check if dancer is registered for the event at all
	if !reg.Status.IsRegistered() {
		reg.Result = models.ResultWasNotRegistered
		return reg
	}

	// Check if dancer is in a singles list and remove from singles
	if reg.Status == models.StatusAsSingle {
		h.removeFromSingles(reg.Dancer)
		h.hist = append(h.hist, &models.HistoryItem{
			Action:    models.HistorySingleRemoved,
			Initiator: reg.Profile,
			EventID:   &h.event.ID,
			Details:   reg.Dancer,
			CreatedAt: nowFn(),
		})
		reg.Result = models.ResultRegistrationRemoved
		reg.Status = models.StatusNotRegistered
		return reg
	}

	// If dancer is in a couple remove the couple
	exCouple := h.removeCouple(reg.Dancer)
	h.hist = append(h.hist, &models.HistoryItem{
		Action:    models.HistoryCoupleRemoved,
		Initiator: reg.Dancer.Profile,
		EventID:   &h.event.ID,
		Details:   exCouple,
		CreatedAt: nowFn(),
	})

	// Set dancer and partner status to not registered
	reg.Status = models.StatusNotRegistered
	reg.Result = models.ResultRegistrationRemoved
	reg.Related = &models.Registration{
		Dancer: reg.Partner,
		Status: models.StatusNotRegistered,
		Result: models.ResultRegistrationRemoved,
		Event:  h.event,
	}
	reg.Partner = nil

	// If partner was signed up as a single, move back to singles (or auto pair if available)
	if reg.Related.AsSingle {
		reg.Related = h.singleRestore(reg.Related, reg.Dancer, exCouple.CreatedAt)
		return reg
	}

	// Otherwise, if couple was created by the partner send notification to the partner
	if reg.Related.Profile != nil && exCouple.CreatedBy.ID == reg.Related.Profile.ID {
		h.notif = append(h.notif, &models.Notification{
			TmplCode:  models.TmplCanceledByPartner,
			Recipient: reg.Related.Profile,
			Payload: models.NotificationPayload{
				Event:   h.event,
				Partner: reg.Dancer,
			},
		})
	}

	// Return the registration
	return reg
}

// tryAutoPair tries to auto pair the dancer with a partner from the singles list.
// Returns the updated registration if the partner was found and paired, otherwise nil.
// If auto pairing is disabled for the event, returns nil.
// In case of auto pairing after couple removal the createdAt time should be set
// to the time of the previous couple creation.
func (h *EventHandler) tryAutoPair(reg *models.Registration, createdAt time.Time) *models.Registration {
	// skip if auto pairing is disabled for the event
	if !h.event.Settings.AutoPairing {
		return nil
	}
	reg.Related = h.firstSingle(reg.Dancer.Role.Opposite())
	if reg.Related == nil {
		return nil
	}
	reg.AsSingle = true
	return h.coupleAdd(reg, true, createdAt)
}

// singleRestore restores the dancer to the singles list after a couple removal.
// If auto pairing is enabled, tries to auto pair the dancer.
// The exPartner parameter is used to send proper the notification to the dancer.
// The exCoupleCreatedAt should be provided to keep the order of couples
// in case of dancer will be auto paired with a new partner.
func (h *EventHandler) singleRestore(reg *models.Registration, exPartner *models.Dancer, exCoupleCreatedAt time.Time) *models.Registration {
	// Try to auto pair the dancer
	if autoPairReg := h.tryAutoPair(reg, exCoupleCreatedAt); autoPairReg != nil {
		h.notif = append(h.notif, &models.Notification{
			TmplCode:  models.TmplAutoPairPartnerChanged,
			Recipient: autoPairReg.Profile,
			Payload: models.NotificationPayload{
				Event:      h.event,
				Partner:    exPartner,
				NewPartner: autoPairReg.Partner,
			},
		})
		return autoPairReg
	}

	// Otherwise, move back to singles
	reg.Status = models.StatusAsSingle
	reg.Result = models.ResultRegisteredAsSingle
	h.event.Singles = append(h.event.Singles, *reg.Dancer)

	// Sort singles by creation time
	sort.Sort(SinglesSorter(h.event.Singles))

	// Send notification that the partner has canceled the registration
	h.notif = append(h.notif, &models.Notification{
		TmplCode:  models.TmplCanceledWithSingle,
		Recipient: reg.Profile,
		Payload: models.NotificationPayload{
			Event:   h.event,
			Partner: exPartner,
		},
	})

	// Add history item
	h.hist = append(h.hist, &models.HistoryItem{
		Action:    models.HistorySingleAdded,
		Initiator: exPartner.Profile,
		EventID:   &h.event.ID,
		Details:   reg.Dancer,
		CreatedAt: nowFn(),
	})

	return reg
}

// findInCouples finds dancers registration in the couples.
// Returns nil if not found.
func (h *EventHandler) findInCouples(dancer *models.Dancer) *models.Registration {
	reg := &models.Registration{
		Status: models.StatusInCouple,
		Event:  h.event,
	}
	for _, couple := range h.event.Couples {
		if h.isSame(dancer, &couple.Dancers[0]) {
			reg.Dancer = &couple.Dancers[0]
			reg.Partner = &couple.Dancers[1]
			return reg
		}
		if h.isSame(dancer, &couple.Dancers[1]) {
			reg.Dancer = &couple.Dancers[1]
			reg.Partner = &couple.Dancers[0]
			return reg
		}
	}
	return nil
}

// findInSingles finds dancer in the singles of the event.
// Returns nil if not found.
func (h *EventHandler) findInSingles(dancer *models.Dancer) *models.Registration {
	reg := &models.Registration{
		Status: models.StatusAsSingle,
		Event:  h.event,
	}
	for _, single := range h.event.Singles {
		if h.isSame(dancer, &single) {
			reg.Dancer = &single
			return reg
		}
	}
	return nil
}

// firstSingle returns registration of the first dancer with the given role in singles list.
// Returns nil if no single dancer with this role found.
func (h *EventHandler) firstSingle(role models.Role) *models.Registration {
	reg := &models.Registration{
		Status: models.StatusAsSingle,
		Event:  h.event,
	}
	for _, single := range h.event.Singles {
		if single.Role == role {
			reg.Dancer = &single
			return reg
		}
	}
	return nil
}

// removeFromSingles removes the dancer from the singles list of the event.
// If dancer found returns the dancer and true, otherwise nil and false.
func (h *EventHandler) removeFromSingles(dancer *models.Dancer) (*models.Dancer, bool) {
	for i, single := range h.event.Singles {
		if h.isSame(dancer, &single) {
			h.event.Singles = append(h.event.Singles[:i], h.event.Singles[i+1:]...)
			return &single, true
		}
	}
	return nil, false
}

// removeCouple removes the couple from the couples list of the event.
// If dancer found returns removed couple, otherwise nil.
func (h *EventHandler) removeCouple(dancer *models.Dancer) *models.Couple {
	for i, couple := range h.event.Couples {
		if h.isSame(dancer, &couple.Dancers[0]) || h.isSame(dancer, &couple.Dancers[1]) {
			h.event.Couples = append(h.event.Couples[:i], h.event.Couples[i+1:]...)
			return &couple
		}
	}
	return nil
}

// isSame checks if dancer is the same as the other dancer on the event
func (h *EventHandler) isSame(dancer, other *models.Dancer) bool {
	switch {
	// Compare profile IDs if both profiles are present
	case dancer.Profile != nil && other.Profile != nil:
		return dancer.ID == other.Profile.ID
	// Compare dancer username (if present in profile ) and other username (if present in full name)
	case dancer.Profile != nil && dancer.Profile.Username != "" && other.Profile == nil:
		u, ok := getUsername(other.FullName)
		return ok && (dancer.Profile.Username == u)
	// Compare dancer username (if present in full name) and other username (if present in profile)
	case dancer.Profile == nil && other.Profile != nil && other.Profile.Username != "":
		u, ok := getUsername(dancer.FullName)
		return ok && (u == other.Profile.Username)
	// Compare usernames (if present in full names) if both profiles are missing
	case dancer.Profile == nil && other.Profile == nil:
		u1, ok1 := getUsername(dancer.FullName)
		u2, ok2 := getUsername(other.FullName)
		return (ok1 && ok2) && (u1 == u2)
	default:
		return false
	}
}

// limitChangeRange calculates the range of couples affected by the limit change.
// Returns the start and end indexes of the couples list
// and a boolean indicating if the limit was increased or decreased.
//
// Zero values for from and to will be treated as no limit.
//
// The start index is inclusive and the end index is exclusive
// so the range can be used as a slice index: couples[start:end].
func limitChangeRange(from, to, couplesNum int) (int, int, bool) {
	if from == 0 || from > couplesNum {
		from = couplesNum
	}
	if to == 0 || to > couplesNum {
		to = couplesNum
	}

	switch {
	case from == to:
		return 0, 0, false
	case from < to:
		return from, to, true
	default:
		return to, from, false
	}
}

// SinglesSorter is a sorter for singles by creation time.
type SinglesSorter []models.Dancer

func (s SinglesSorter) Len() int           { return len(s) }
func (s SinglesSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SinglesSorter) Less(i, j int) bool { return s[i].CreatedAt.Before(s[j].CreatedAt) }

// CouplesSorter is a sorter for couples by creation time.
type CouplesSorter []models.Couple

func (c CouplesSorter) Len() int           { return len(c) }
func (c CouplesSorter) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c CouplesSorter) Less(i, j int) bool { return c[i].CreatedAt.Before(c[j].CreatedAt) }
