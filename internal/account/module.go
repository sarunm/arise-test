package account

import (
	"github.com/gin-gonic/gin"
	"github.com/sarunm/arise-test/pkg/cache"
	"gorm.io/gorm"
)

type Module struct {
	Service Service
	Handler *Handler
}

func NewModule(db *gorm.DB, cache cache.Cache) *Module {
	repo    := newRepo(db)
	service := NewService(repo, cache)
	handler := NewHandler(service)

	return &Module{
		Service: service,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	m.Handler.RegisterRoutes(r)
}
