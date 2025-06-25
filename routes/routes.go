package routes

import (
	"github.com/crafty-ezhik/amocrmproxy/handlers"
	"github.com/go-chi/chi/v5"
)

func InitRoutes(r *chi.Mux, h handlers.AppHandlers) {
	r.Route("/{crm_address}", func(r chi.Router) {
		// Добавление адреса crm системы в контекст
		r.Use(h.AddAddressToCtx)

		// Получение токенов
		r.Post("/oauth2/access_token", h.GetToken())

		// Отправка уведомления о звонке
		r.Post("/api/v2/events", h.CreateCallEvents())

		// Получение и создание пользователей в РТУ Сател
		r.Post("/user", h.CreateUserInRTU())
		r.Get("/user", h.GetUserFromRTU())

		r.Route("/api/v4", func(r chi.Router) {
			r.Get("/contacts", h.GetContacts())
			r.Get("/companies", h.CreateCompanies())
			r.Post("/calls", h.EndCall())
			r.Post("/contacts", h.CreateContacts())
			r.Post("/leads/unsorted/sip", h.AddUnsorted())
			r.Post("/leads/{id}/link", h.LinkUnsorted())
		})

	})
}
