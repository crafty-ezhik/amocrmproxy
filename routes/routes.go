package routes

import (
	"github.com/crafty-ezhik/amocrmproxy/handlers"
	"github.com/go-chi/chi/v5"
)

// InitRoutes - производит инициализацию маршрутов.
// Также добавляет middleware, который добавляет в контекст запроса такой параметр пути, как crm_address
//
// Принимает chi.Mux и интерфейс handlers.AppHandlers
func InitRoutes(r *chi.Mux, h handlers.AppHandlers) {
	// Получение кода авторизации
	r.Route("/crmproxy", func(r chi.Router) {
		r.Get("/callback/amo", h.GetAuthCode())

		// Получение и создание пользователей в РТУ Сател
		r.Post("/user", h.CreateUserInRTU())
		r.Get("/user", h.GetUserFromRTU())

		// Заказ обратного звонка
		r.Post("/call", h.OrderingCallback())

		r.Route("/{crm_address}", func(r chi.Router) {
			// Добавление адреса crm системы в контекст

			r.Use(h.AddAddressToCtx)

			// Получение токенов
			r.Post("/oauth2/access_token", h.GetToken())

			// Отправка уведомления о звонке
			r.Post("/api/v2/events", h.CreateCallEvents())

			r.Route("/api/v4", func(r chi.Router) {
				r.Get("/contacts", h.GetContacts())
				r.Get("/companies", h.GetCompanies())
				r.Post("/calls", h.EndCall())
				r.Post("/contacts", h.CreateContacts())
				r.Post("/leads/unsorted/sip", h.AddUnsorted())
				r.Post("/leads/{entity_id}/link", h.LinkUnsorted())
			})

		})
	})

}
