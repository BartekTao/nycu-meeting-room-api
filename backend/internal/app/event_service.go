package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/BartekTao/nycu-meeting-room-api/internal/domain"
	"github.com/BartekTao/nycu-meeting-room-api/pkg/lock"
	"github.com/BartekTao/nycu-meeting-room-api/pkg/notification"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type UpsertEventRequest struct {
	ID              *string     `json:"_id,omitempty"`
	Title           string      `json:"title"`
	Description     *string     `json:"description"`
	StartAt         int64       `json:"startAt"`
	EndAt           int64       `json:"endAt"`
	RoomID          *string     `json:"roomId"`
	ParticipantsIDs []string    `json:"participantsIDs"`
	Notes           *string     `json:"notes"`
	RemindAt        int64       `json:"remindAt"`
	UpdaterID       string      `json:"updaterID"`
	AttachedFile    domain.File `json:"attachedFile"`
}

type EventService interface {
	Upsert(ctx context.Context, req UpsertEventRequest) (*domain.Event, error)
	Delete(ctx context.Context, id string) (*domain.Event, error)
	GetByID(ctx context.Context, id string) (*domain.Event, error)
	GetUserEvents(ctx context.Context, ids []string, startAt, endAt int64) (map[string][]domain.Event, error)
	UpdateSummary(ctx context.Context, id string, summary string, updaterID string) (bool, error)
}

type eventService struct {
	eventRepository domain.EventRepository
	userRepo        domain.UserRepo
	mailHandler     notification.MailHandler
	locker          lock.DistributedLocker
}

func NewEventService(
	eventRepository domain.EventRepository,
	locker lock.DistributedLocker,
	userRepo domain.UserRepo,
	mailHandler notification.MailHandler,
) EventService {
	return &eventService{
		eventRepository: eventRepository,
		locker:          locker,
		userRepo:        userRepo,
		mailHandler:     mailHandler,
	}
}

func (s *eventService) Upsert(ctx context.Context, req UpsertEventRequest) (*domain.Event, error) {
	event := domain.Event{
		ID:              req.ID,
		Title:           req.Title,
		Description:     req.Description,
		StartAt:         req.StartAt,
		EndAt:           req.EndAt,
		ParticipantsIDs: req.ParticipantsIDs,
		Notes:           req.Notes,
		RemindAt:        req.RemindAt,
		UpdaterID:       req.UpdaterID,
		AttachedFile:    req.AttachedFile,
	}

	var res *domain.Event
	var err error

	if req.RoomID == nil {
		res, err = s.eventRepository.Upsert(ctx, event)
		if err != nil {
			return nil, err
		}
	} else {
		event.RoomReservation = &domain.RoomReservation{
			RoomID:            req.RoomID,
			ReservationStatus: domain.ReservationStatus_Confirmed,
		}

		var oldEvent *domain.Event

		if event.ID != nil {
			oldEvent, err = s.eventRepository.GetByID(ctx, *event.ID)
			if err != nil {
				return nil, err
			}
			if oldEvent == nil {
				return nil, gqlerror.Errorf("Event not exist")
			}
		}

		if oldEvent != nil && oldEvent.StartAt == event.StartAt &&
			oldEvent.EndAt == event.EndAt &&
			*oldEvent.RoomReservation.RoomID == *event.RoomReservation.RoomID {

			res, err = s.eventRepository.Upsert(ctx, event)
			if err != nil {
				return nil, err
			}
		} else {
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			key := *req.RoomID
			locked, err := s.locker.TryLockWithWait(key, 500*time.Millisecond, 3)
			if err != nil {
				return nil, err
			}
			defer s.locker.Unlock(locked)

			if oldEvent != nil {

				oldEvent.RoomReservation.ReservationStatus = domain.ReservationStatus_Canceled
				_, err = s.eventRepository.Upsert(ctx, *oldEvent)
				if err != nil {
					return nil, err
				}
			}

			available, err := s.eventRepository.CheckAvailableRoom(ctx, *event.RoomReservation.RoomID, event.StartAt, event.EndAt)
			if err != nil {
				return nil, err
			}

			if !available {
				event.RoomReservation.ReservationStatus = domain.ReservationStatus_Canceled
			}

			res, err = s.eventRepository.Upsert(ctx, event)
			if err != nil {
				return nil, err
			}
		}
	}

	// sent email
	go func(event *domain.Event, req UpsertEventRequest) {
		err := s.sendEmailAfterUpsert(context.Background(), event, req)
		if err != nil {
			log.Printf("Error sending email: %v\n", err.Error())
		}
	}(res, req)

	return res, nil
}

func (s *eventService) sendEmailAfterUpsert(ctx context.Context, event *domain.Event, req UpsertEventRequest) error {
	var err error
	if req.ID == nil {
		err = s.sendEmail(ctx, event, "Invited", "")
	} else {
		err = s.sendEmail(ctx, event, "Updated", "")
	}
	return err
}

func (s *eventService) sendEmail(ctx context.Context, res *domain.Event, mailPrefix string, content string) error {
	users, err := s.userRepo.GetByIDs(ctx, res.ParticipantsIDs)
	if err != nil {
		log.Printf("Error getting userEmail: %v\n", err)
		return err
	}
	userEmails := make([]string, len(users))
	for i, user := range users {
		userEmails[i] = user.Email
	}
	startTime := time.UnixMilli(res.StartAt)
	loc, err := time.LoadLocation("Local")
	if err != nil {
		log.Printf("Error loading location: %v\n", err)
		return err
	}
	startTime = startTime.In(loc)
	formattedTime := startTime.Format("2006-01-02 15:04:05 -0700")
	s.mailHandler.Send(userEmails, fmt.Sprintf("%s: %s - %s", mailPrefix, res.Title, formattedTime), content)
	return nil
}

func (s *eventService) Delete(ctx context.Context, id string) (*domain.Event, error) {
	res, err := s.eventRepository.Delete(ctx, id)
	if err != nil {
		return nil, err
	} else {
		// sent email
		go func(event *domain.Event) {
			err := s.sendEmail(context.Background(), event, "Canceled", "")
			if err != nil {
				log.Printf("Error sending email: %v\n", err.Error())
			}
		}(res)
		return res, nil
	}
}

func (s *eventService) GetByID(ctx context.Context, id string) (*domain.Event, error) {
	res, err := s.eventRepository.GetByID(ctx, id)
	return res, err
}

func (s *eventService) GetUserEvents(ctx context.Context, ids []string, startAt, endAt int64) (map[string][]domain.Event, error) {
	userEventMap, err := s.eventRepository.GetByUsers(ctx, ids, startAt, endAt)
	if err != nil {
		return nil, err
	}
	return userEventMap, nil
}

func (s *eventService) CheckUserAvailable(ctx context.Context, ids []string, startAt, endAt int64) (map[string]bool, error) {
	userEventMap, err := s.eventRepository.GetByUsers(ctx, ids, startAt, endAt)
	if err != nil {
		return nil, err
	}

	availableMap := make(map[string]bool)
	for userID, userEvent := range userEventMap {
		availableMap[userID] = userEvent != nil
	}
	return availableMap, nil
}

func (s *eventService) UpdateSummary(ctx context.Context, id string, summary string, updaterID string) (bool, error) {
	res, err := s.eventRepository.UpdateSummary(ctx, id, summary, updaterID)
	if err != nil {
		return false, err
	}

	go func(eventId string, summary string) {
		ctx := context.Background()
		event, err := s.eventRepository.GetByID(ctx, eventId)
		if err != nil {
			log.Printf("Error sending email: %v\n", err.Error())
			return
		}
		err = s.sendEmail(ctx, event, "Conclusion", summary)
		if err != nil {
			log.Printf("Error sending email: %v\n", err.Error())
		}
	}(id, summary)

	return res, nil
}
