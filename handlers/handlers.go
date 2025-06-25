package handlers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/crafty-ezhik/amocrmproxy/config"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type AppHandlers interface {
	AddAddressToCtx(next http.Handler) http.Handler
	CreateContacts() http.HandlerFunc
	GetUserFromRTU() http.HandlerFunc
	CreateUserInRTU() http.HandlerFunc
	LinkUnsorted() http.HandlerFunc
	AddUnsorted() http.HandlerFunc
	CreateCallEvents() http.HandlerFunc
	CreateCompanies() http.HandlerFunc
	GetContacts() http.HandlerFunc
	GetToken() http.HandlerFunc
	EndCall() http.HandlerFunc
}

type appHandlers struct {
	log            *zap.Logger
	client         *http.Client
	insecureClient *http.Client
	rtuAddr        string
	serviceCode    string
}

func NewAppHandlers(log *zap.Logger, cfg *config.Config) AppHandlers {
	handlers := &appHandlers{
		log:    log,
		client: http.DefaultClient,
		insecureClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		rtuAddr:     cfg.RTU.Host,
		serviceCode: cfg.CRM.ServiceCode,
	}
	return handlers
}

func (h *appHandlers) AddAddressToCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		crmAddress := chi.URLParam(r, "crm_address")
		ctx := context.WithValue(r.Context(), "crm_address", crmAddress)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *appHandlers) CreateContacts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Println(string(body))
		fmt.Println("Запрос создания контакта")

		req, _ := http.NewRequest("POST", "https://mmvamobizneslinecom.amocrm.ru/api/v4/contacts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", r.Header.Get("Authorization"))

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}

		resp, err := client.Do(req)
		if err != nil {
			slog.Error("Error fetching contacts", err.Error())
		}
		defer resp.Body.Close()

		fmt.Println("response Status:", resp.Status)

		respBody, _ := io.ReadAll(resp.Body)
		fmt.Println(string(respBody))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
	}
}

func (h *appHandlers) CreateUserInRTU() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

func (h *appHandlers) GetUserFromRTU() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Метод запроса: " + r.Method)
		fmt.Println("Параметры запроса: " + r.URL.Query().Get("user_id"))
		fmt.Println("парсинг и маршалинг в мапу")

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}

		userID := r.URL.Query().Get("user_id")
		if userID != "" && r.Method == http.MethodGet {
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://78.155.208.225:8431/user?user_id=%s", userID), nil)
			req.Header.Set("Content-Type", "application/json")
			fmt.Println("Заголовок авторизации: " + r.Header.Get("Authorization"))
			req.Header.Set("Authorization", r.Header.Get("Authorization"))

			resp, err := client.Do(req)
			if err != nil {
				slog.Error("Error fetching  user with ID:", userID)
			}

			respBody, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()

			w.WriteHeader(resp.StatusCode)
			w.Write(respBody)
			return
		}
		if r.Method == http.MethodGet {
			req, _ := http.NewRequest(http.MethodGet, "https://78.155.208.225:8431/user", nil)
			req.Header.Set("Content-Type", "application/json")
			fmt.Println("Заголовок авторизации: " + r.Header.Get("Authorization"))
			req.Header.Set("Authorization", r.Header.Get("Authorization"))

			resp, err := client.Do(req)
			if err != nil {
				slog.Error("Error fetching  users ")
			}

			respBody, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()

			w.WriteHeader(resp.StatusCode)
			w.Write(respBody)
			return
		}

		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		fmt.Println("Запрос от amoCRM: " + string(body))

		var jsonData map[string]interface{}
		_ = json.Unmarshal(body, &jsonData)

		reqBody := map[string]interface{}{
			"id":          jsonData["id"],
			"phoneNumber": jsonData["phoneNumber"],
			"email":       jsonData["email"],
			"name":        jsonData["name"],
			"sipLogin":    jsonData["sipLogin"],
			"sipPassword": jsonData["sipPassword"],
		}

		newBody, _ := json.Marshal(reqBody)
		fmt.Println("Новое тело запроса для RTU: " + string(newBody))

		fmt.Println("Формируем запрос в РТУ")
		req, _ := http.NewRequest(http.MethodPost, "https://78.155.208.225:8431/user", bytes.NewBuffer(newBody))
		req.Header.Set("Content-Type", "application/json")
		fmt.Println("Заголовок авторизации: " + r.Header.Get("Authorization"))
		req.Header.Set("Authorization", r.Header.Get("Authorization"))

		fmt.Println("Делаем запрос к РТУ")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}

		respBody, _ := io.ReadAll(resp.Body)
		fmt.Println("Тело ответа от РТУ:" + string(respBody))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBody)
		/*
			"id": "1",
			"phoneNumber": "9976",
			"email": "ivanov@test.org",
			"name": "Иван Иванов",
			"sipLogin": "9976",
			"sipPassword": "Yk1jbTmW"
		*/
	}
}

func (h *appHandlers) LinkUnsorted() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Запрос на привязку несортированного звонка")
		entityID := chi.URLParam(r, "id")

		token := r.Header.Get("Authorization")

		body, _ := io.ReadAll(r.Body)

		var jsonData []map[string]interface{}
		_ = json.Unmarshal(body, &jsonData)

		reqBody := map[string]interface{}{
			"link": map[string]interface{}{
				"entity_id":   jsonData[0]["to_entity_id"],
				"entity_type": "leads",
			},
		}

		bytesBody, _ := json.Marshal(reqBody)
		fmt.Println(string(bytesBody))

		url := fmt.Sprintf("https://mmvamobizneslinecom.amocrm.ru/api/v4/leads/unsorted/%s/link", entityID)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bytesBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)

		resp, _ := http.DefaultClient.Do(req)
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		fmt.Println(string(respBody))

		w.Header().Set("Content-Type", "application/json")
		w.Write(respBody)
	}
}

func (h *appHandlers) AddUnsorted() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		var jsonData []map[string]interface{}
		err := json.Unmarshal(body, &jsonData)
		jsonData[0]["metadata"].(map[string]interface{})["call_responsible"] = "78123839300"
		jsonData[0]["metadata"].(map[string]interface{})["phone"] = "79211352609"
		jsonData[0]["metadata"].(map[string]interface{})["service_code"] = "bl_vats" // TODO: Вынести в Config
		//delete(jsonData[0]["metadata"].(map[string]interface{}), "from")

		newBody, _ := json.Marshal(jsonData)
		fmt.Println(string(newBody))

		token := r.Header.Get("Authorization")
		req, err := http.NewRequest("POST", "https://mmvamobizneslinecom.amocrm.ru/api/v4/leads/unsorted/sip", bytes.NewBuffer(newBody))
		if err != nil {
			slog.Error("Error creating request", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("Error sending request", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		fmt.Println(string(respBody))
		w.WriteHeader(resp.StatusCode)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBody)

	}
}

func (h *appHandlers) CreateCallEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phoneNumber := r.FormValue("add[0][phone_number]")
		eventType := r.FormValue("add[0][type]")
		users := "12646090" //r.FormValue("add[0][users][0]") // TODO: Поставить ID из запроса

		fmt.Println("Call Event Type: " + eventType)
		fmt.Println("Call Users: " + users)
		fmt.Println("Call Phone: " + phoneNumber)

		type Item map[string]any

		body := Item{
			"add": []Item{
				{
					"type":         eventType,
					"users":        users,
					"phone_number": phoneNumber,
				},
			},
		}

		jsonData, err := json.Marshal(body)
		if err != nil {
			slog.Error("Error marshalling body")
		}

		token := r.Header.Get("Authorization")
		req, err := http.NewRequest("POST", "https://mmvamobizneslinecom.amocrm.ru/api/v2/events", bytes.NewBuffer(jsonData))
		if err != nil {
			slog.Error("Error creating new request")
		}

		req.Header.Set("Authorization", token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("Error creating new request")
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("Error creating new request")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write(respBody)

	}
}

func (h *appHandlers) CreateCompanies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		req, err := http.NewRequest("GET", "https://mmvamobizneslinecom.amocrm.ru/api/v4/companies?limit=100&page=1", nil)
		req.Header.Set("Authorization", token)
		req.Header.Set("Content-Type", "application/json")
		req.URL.Query().Set("limit", "100") // TODO: Подставить значения из request url
		req.URL.Query().Set("page", "1")    // TODO: Подставить значения из request url

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("Error getting companies")
			return
		}
		fmt.Println(resp)

		newBody, _ := io.ReadAll(resp.Body)
		fmt.Println(string(newBody))
		w.WriteHeader(resp.StatusCode)
		w.Write(newBody)
	}
}

func (h *appHandlers) GetContacts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		req, err := http.NewRequest("GET", "https://mmvamobizneslinecom.amocrm.ru/api/v4/contacts?limit=100&page=1", nil)
		req.Header.Set("Authorization", token)
		req.Header.Set("Content-Type", "application/json")
		req.URL.Query().Set("limit", "100") // TODO: Подставить значения из request url
		req.URL.Query().Set("page", "1")    // TODO: Подставить значения из request url

		fmt.Println("Готовый URL: ", req.URL)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("Error creating contacts")
			return
		}
		fmt.Println(resp)

		newBody, _ := io.ReadAll(resp.Body)
		fmt.Println(string(newBody))
		w.WriteHeader(resp.StatusCode)
		w.Write(newBody)
	}
}

func (h *appHandlers) GetToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Получен запрос на получение токенов")
		//http.Redirect(w, r, "mmvamobizneslinecom.amocrm.ru/oauth2/access_token", 302)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Error reading body")
			return
		}
		newBody := strings.Split(string(body), "&")
		temp := make(map[string]string)

		for _, value := range newBody {
			items := strings.Split(value, "=")
			if len(items) != 2 {
				continue
			}
			temp[items[0]] = items[1]
		}

		temp["redirect_uri"] = "https://78ac-46-249-44-245.ngrok-free.app/callback/amo"
		// TODO: Подумать, как сделать так, чтобы это значение парсилось и куда то выдавалось и
		//		не приходилось ручками доставать из строки запроса в браузере

		jsonData, err := json.Marshal(temp)

		resp, err := http.Post("https://mmvamobizneslinecom.amocrm.ru/oauth2/access_token", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			slog.Error("Error creating token")
		}

		respBody, _ := io.ReadAll(resp.Body)
		fmt.Println(string(respBody))

		w.WriteHeader(200)
		w.Write(respBody)
		//fmt.Println("Запрос на получение токенов:" + string(body))
	}
}

func (h *appHandlers) EndCall() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Error reading body")
			return
		}
		fmt.Println("Запрос, что звонок завершился: " + string(body))

		var jsonData []map[string]interface{}
		err = json.Unmarshal(body, &jsonData)
		// TODO: Проверить, какой ID приходит
		jsonData[0]["created_by"] = 12646090
		newBody, _ := json.Marshal(jsonData)
		fmt.Println(jsonData)

		req, err := http.NewRequest("POST", "https://mmvamobizneslinecom.amocrm.ru/api/v4/calls", bytes.NewBuffer(newBody))
		if err != nil {
			slog.Error("Error creating new request")
		}

		token := r.Header.Get("Authorization")
		req.Header.Set("Authorization", token)
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			slog.Error("Error creating new request")
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("Error creating new request")
		}
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Println(string(respBody))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(202)
		w.Write(respBody)
		return
	}
}
