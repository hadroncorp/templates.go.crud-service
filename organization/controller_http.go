package organization

import (
	"log/slog"
	"net/http"

	"github.com/hadroncorp/geck/persistence/identifier"
	"github.com/hadroncorp/geck/transport"
	geckhttp "github.com/hadroncorp/geck/transport/http"
	"github.com/hadroncorp/geck/validation"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

type ControllerHTTP struct {
	manager   Manager
	fetcher   Fetcher
	lister    Lister
	idFactory identifier.Factory
	validator validation.Validator
	logger    *slog.Logger
}

// compile-time assertion
var _ geckhttp.Controller = (*ControllerHTTP)(nil)

// NewControllerHTTP creates a new instance of [ControllerHTTP].
func NewControllerHTTP(manager Manager, fetcher Fetcher, lister Lister, idFactory identifier.Factory,
	validator validation.Validator, logger *slog.Logger) ControllerHTTP {
	return ControllerHTTP{
		manager:   manager,
		fetcher:   fetcher,
		lister:    lister,
		idFactory: idFactory,
		validator: validator,
		logger:    logger,
	}
}

func (c ControllerHTTP) SetEndpoints(_ *echo.Echo) {
}

func (c ControllerHTTP) SetVersionedEndpoints(g *echo.Group) {
	g.POST("/organizations", c.register)
	g.GET("/organizations/:organization_id", c.get)
	g.PATCH("/organizations/:organization_id", c.update)
	g.DELETE("/organizations/:organization_id", c.delete)
	g.GET("/organizations", c.list)
}

func (c ControllerHTTP) register(e echo.Context) error {
	id, err := c.idFactory.NewID()
	if err != nil {
		return err
	}

	body := registerRequestHTTP{}
	if err = e.Bind(&body); err != nil {
		return err
	}

	if err = c.validator.Validate(e.Request().Context(), body); err != nil {
		return err
	}

	org, err := c.manager.Register(e.Request().Context(), RegisterArguments{
		ID:   id,
		Name: body.Name,
	})
	if err != nil {
		return err
	}
	return e.JSON(http.StatusCreated, transport.DataContainer[responseHTTP]{
		Data: newResponseHTTP(org),
	})
}

func (c ControllerHTTP) get(e echo.Context) error {
	id := e.Param("organization_id")
	org, err := c.fetcher.GetByID(e.Request().Context(), id)
	if err != nil {
		return err
	}
	return e.JSON(http.StatusOK, transport.DataContainer[responseHTTP]{
		Data: newResponseHTTP(org),
	})
}

func (c ControllerHTTP) update(e echo.Context) error {
	id := e.Param("organization_id")

	body := updateResponseHTTP{}
	if err := e.Bind(&body); err != nil {
		return err
	}

	if err := c.validator.Validate(e.Request().Context(), body); err != nil {
		return err
	}

	opts := []UpdateOption{
		WithUpdatedName(body.Name),
	}

	org, err := c.manager.ModifyByID(e.Request().Context(), id, opts...)
	if err != nil {
		return err
	}

	return e.JSON(http.StatusOK, transport.DataContainer[responseHTTP]{
		Data: newResponseHTTP(org),
	})
}

func (c ControllerHTTP) delete(e echo.Context) error {
	id := e.Param("organization_id")
	err := c.manager.DeleteByID(e.Request().Context(), id)
	if err != nil {
		return err
	}
	return e.NoContent(http.StatusNoContent)
}

func (c ControllerHTTP) list(e echo.Context) error {
	page, err := c.lister.List(e.Request().Context(),
		WithListPageOptions(geckhttp.NewPaginationOptions(e)...),
	)
	if err != nil {
		return err
	} else if len(page.Items) == 0 {
		return e.NoContent(http.StatusNotFound)
	}

	return e.JSON(http.StatusOK, transport.DataContainer[transport.PageResponse[responseHTTP]]{
		Data: transport.PageResponse[responseHTTP]{
			TotalItems:        page.TotalItems,
			PreviousPageToken: page.PreviousPageToken,
			NextPageToken:     page.NextPageToken,
			Items: lo.Map(page.Items, func(o Organization, _ int) responseHTTP {
				return newResponseHTTP(o)
			}),
		},
	})
}

// -- Models --

type registerRequestHTTP struct {
	Name string `json:"name" validate:"required,lte=48"`
}

type updateResponseHTTP struct {
	Name *string `json:"name" validate:"omitempty,lte=48"`
}

type responseHTTP struct {
	ID   string `json:"organization_id"`
	Name string `json:"name"`
}

func newResponseHTTP(org Organization) responseHTTP {
	return responseHTTP{
		ID:   org.ID(),
		Name: org.Name(),
	}
}
