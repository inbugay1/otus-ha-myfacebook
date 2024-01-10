package postfanoutservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"myfacebook/internal/config"
	"myfacebook/internal/postfeedcache"
	"myfacebook/internal/repository"
	"myfacebook/internal/rmq"
)

type postFeedRMQMessage struct {
	Operation string `json:"operation,omitempty"`
	PostID    string `json:"post_id"`
	PostText  string `json:"post_text"`
	AuthorID  string `json:"author_id"`
}

type Service struct {
	rmq            *rmq.RMQ
	userRepository repository.UserRepository
	postFeedCache  *postfeedcache.Cache
	envConfig      *config.EnvConfig

	done chan struct{}
	wg   *sync.WaitGroup
}

var errInvalidPostOperation = errors.New("invalid post operation")

func New(rmq *rmq.RMQ, userRepository repository.UserRepository, postFeedCache *postfeedcache.Cache, envConfig *config.EnvConfig) *Service {
	return &Service{
		rmq:            rmq,
		userRepository: userRepository,
		postFeedCache:  postFeedCache,
		envConfig:      envConfig,
		done:           make(chan struct{}),
		wg:             &sync.WaitGroup{},
	}
}

func (s *Service) Start(ctx context.Context) error {
	msgs, err := s.rmq.Consume(ctx, "/post/feed")
	if err != nil {
		return fmt.Errorf("postfanoutservice failed to consume rmq messages: %w", err)
	}

	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					slog.Error("RMQ channel is closed")

					select {
					case <-time.After(1 * time.Second):
						newMsgs, err := s.rmq.Consume(ctx, "/post/feed")
						if err != nil {
							slog.Error(fmt.Sprintf("postfanoutservice failed to consume rmq messages: %s", err))

							continue
						}

						msgs = newMsgs
					case <-s.done:
						return
					}

					continue
				}

				err := s.processMessage(ctx, msg.Body)
				if err != nil {
					if err := msg.Reject(true); err != nil {
						slog.Error(fmt.Sprintf("Error on rejecting rmq message: %s", err))
					}

					slog.Error(fmt.Sprintf("Error on processing rmq message: %s", err))

					continue
				}

				if err := msg.Ack(false); err != nil {
					slog.Error(fmt.Sprintf("Error on ack rmq message: %s", err))
				}

			case <-s.done:
				return
			}
		}
	}()

	slog.Info("Successfully started post fanout service")

	return nil
}

func (s *Service) processMessage(ctx context.Context, msg []byte) error {
	var postMsg postFeedRMQMessage

	err := json.Unmarshal(msg, &postMsg)
	if err != nil {
		return fmt.Errorf("postfanoutservice failed to unmarshal rmq message: %w", err)
	}

	usersCount, err := s.userRepository.GetUsersCountByFriendID(ctx, postMsg.AuthorID)
	if err != nil {
		return fmt.Errorf("postfanoutservice failed to count users that have the post author as friend: %w", err)
	}

	if usersCount >= s.envConfig.PopularFriendUsersCount {
		return nil
	}

	usersIDs, err := s.userRepository.GetUsersIDsByFriendID(ctx, postMsg.AuthorID)
	if err != nil {
		return fmt.Errorf("postfanoutservice failed to get users ids from repo: %w", err)
	}

	switch postMsg.Operation {
	case "add":
		postFeedRMQMsg, err := json.Marshal(postFeedRMQMessage{
			PostID:   postMsg.PostID,
			PostText: postMsg.PostText,
			AuthorID: postMsg.AuthorID,
		})
		if err != nil {
			return fmt.Errorf("postfanoutservice failed to marshal rmq message: %w", err)
		}

		for _, userID := range usersIDs {
			err := s.postFeedCache.AddPostID(ctx, userID, postMsg.PostID)
			if err != nil {
				return fmt.Errorf("postfanoutservice failed to push post to post feed cache: %w", err)
			}

			err = s.rmq.Publish(ctx, "/post/feed/posted", userID, postFeedRMQMsg)
			if err != nil {
				return fmt.Errorf("postfanoutservice failed to publish rmq message: %w", err)
			}
		}
	case "remove":
		for _, userID := range usersIDs {
			err := s.postFeedCache.RemovePostID(ctx, userID, postMsg.PostID)
			if err != nil {
				return fmt.Errorf("postfanoutservice failed to remove post from post feed cache: %w", err)
			}
		}
	default:
		return fmt.Errorf("postfanoutservice failed to process message %s: %w", msg, errInvalidPostOperation)
	}

	return nil
}

func (s *Service) Stop() {
	slog.Info("Stopping post fanout service...")

	close(s.done)
	s.wg.Wait()

	slog.Info("Post fanout service stopped")
}
