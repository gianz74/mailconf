package mockservice

import (
	"github.com/gianz74/mailconf/internal/config"
	"github.com/gianz74/mailconf/internal/service"
)

func SetupMockServices() {
	service.SetMbsync(NewMockMbsync)
	service.SetImapnotify(NewMockImapnotify)
}

func RestoreServices() {
	service.SetMbsync(service.MbsyncCtor)
	service.SetImapnotify(service.ImapnotifyCtor)
}

func NewMockMbsync(cfg *config.Config) service.Service {
	return &MockService{
		Service: service.MbsyncCtor(cfg),
	}
}

func NewMockImapnotify(cfg *config.Config, profile *config.Profile) service.Service {
	return &MockService{
		Service: service.ImapnotifyCtor(cfg, profile),
	}
}

type MockService struct {
	service.Service
}

func (MockService) Start()        {}
func (MockService) Stop()         {}
func (MockService) Enable()       {}
func (MockService) Disable()      {}
func (MockService) Remove() error { return nil }
