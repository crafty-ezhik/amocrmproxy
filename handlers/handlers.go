package handlers

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/crafty-ezhik/amocrmproxy/config"
	"github.com/crafty-ezhik/amocrmproxy/email"
	"github.com/crafty-ezhik/amocrmproxy/utils"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
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
	GetCompanies() http.HandlerFunc
	GetContacts() http.HandlerFunc
	GetToken() http.HandlerFunc
	EndCall() http.HandlerFunc
	GetAuthCode() http.HandlerFunc
	OrderingCallback() http.HandlerFunc
}

type appHandlers struct {
	log            *zap.Logger
	client         *http.Client
	insecureClient *http.Client
	rtuAddr        string
	serviceCode    string
	serverHost     string
	emailRecipient string
	ec             *email.EmailClient
}

func NewAppHandlers(log *zap.Logger, cfg *config.Config, ec *email.EmailClient) AppHandlers {
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
		rtuAddr:        cfg.RTU.Host,
		serviceCode:    cfg.CRM.ServiceCode,
		serverHost:     cfg.Server.Host,
		emailRecipient: cfg.Email.Recipient,
		ec:             ec,
	}
	return handlers
}

func (h *appHandlers) OrderingCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Ordering callback")

		// Parse and unmarshalling body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("Error reading body", zap.Error(err))
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}
		h.log.Debug("Body", zap.ByteString("body", body))

		h.log.Debug("Start unmarshalling body in to map")
		var jsonData utils.AnyMap
		err = json.Unmarshal(body, &jsonData)
		if err != nil {
			h.log.Error("Error unmarshalling body", zap.Error(err))
			http.Error(w, "Error unmarshalling body", http.StatusBadRequest)
			return
		}
		h.log.Debug("Unmarshalling body in to map successfully")

		// Get Body data
		h.log.Debug("Get data from request body")
		rawUserID, ok := jsonData["user"].(float64)
		if !ok {
			h.log.Error("Error converting the interface to the float64 type ", zap.Error(err))
			http.Error(w, "Error converting the interface to the float64 type", http.StatusBadRequest)
			return
		}
		userID := strconv.FormatFloat(rawUserID, 'f', 0, 64)
		destNumber, ok := jsonData["destination"].(string)
		if !ok {
			h.log.Error("Error converting the interface to the string type ", zap.Error(err))
			http.Error(w, "Error converting the interface to the string type ", http.StatusBadRequest)
			return
		}
		h.log.Debug("Get data from request body successfully")

		// Get Domain
		h.log.Debug("Get domain from Authorization header")
		domain, err := getDomain(r)
		if err != nil {
			h.log.Error("Error getting domain", zap.Error(err))
			http.Error(w, "Error getting domain", http.StatusBadRequest)
			return
		}
		h.log.Debug("Get domain from Authorization header successfully")

		// Get user info from RTU
		h.log.Debug("Sending request to RTU for get user info")
		url := fmt.Sprintf("https://%s/user?user_id=%s", h.rtuAddr, userID)
		resp := utils.MakeRequest(r, h.insecureClient, http.MethodGet, url, nil)
		h.log.Debug("Response body", zap.ByteString("body", resp.Body))

		type userData struct {
			List []struct {
				Email       string `json:"email"`
				PhoneNumber string `json:"sipLogin"`
				Password    string `json:"sipPassword"`
			} `json:"list"`
		}

		var data userData
		err = json.Unmarshal(resp.Body, &data)
		if err != nil {
			h.log.Error("Error unmarshalling body", zap.Error(err))
			http.Error(w, "Error unmarshalling body", http.StatusBadRequest)
			return
		}

		srcNumber := data.List[0].PhoneNumber
		userLogin := data.List[0].Email
		userPassword := data.List[0].Password
		h.log.Debug("Sending request to RTU for get user info successfully")

		// Sending request to creating ordering callback
		h.log.Debug("Sending request to creating ordering callback")
		status, err := sendOrderingCallbackRequest(destNumber, srcNumber, userLogin, userPassword, domain)
		if err != nil {
			h.log.Error("Error getting user info", zap.Error(err))
			http.Error(w, "Error getting user info", http.StatusBadRequest)
			return
		}
		if status != http.StatusOK {
			utils.SendResponse(w, &utils.Response{
				Body:   nil,
				Status: 500,
				Error:  err,
			})
		}

		h.log.Debug("Sending request to creating ordering callback successfully")
		h.log.Info("Ordering callback successfully")
		utils.SendResponse(w, &utils.Response{
			Body:   []byte("{\"code\": 200, \"reason\":\"OK\"}"),
			Status: status,
			Error:  nil,
		})
	}
}

func sendOrderingCallbackRequest(destNumber, srcNumber, userLogin, userPassword, domain string) (int, error) {
	// Инициализировать структуры

	// Сделать запрос

	// Отдать ответ

	return http.StatusOK, nil
}

func (h *appHandlers) GetAuthCode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: На данный момент заглушка. При реализации ЛК, надо будет видоизменить
		code := r.URL.Query().Get("code")
		refer := r.URL.Query().Get("referer")

		//отправка на email
		msg := fmt.Sprintf("Code: %s\nClient account: %s", code, refer)
		go func() {
			err := h.ec.SendEmailWithTLS(h.emailRecipient, msg)
			if err != nil {
				h.log.Error("Error", zap.Error(err))
				return
			}
		}()

		url := fmt.Sprintf("https://%s/amo-market#category-installed", refer)
		http.Redirect(w, r, url, http.StatusPermanentRedirect)
	}
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
		h.log.Info("Creating Contacts")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("Error reading body", zap.Error(err))
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}
		h.log.Debug("Body", zap.ByteString("body", body))

		url := fmt.Sprintf("https://%s", trimPath(r.URL.Path))
		h.log.Debug("Request URL", zap.String("url", url))

		resp := utils.MakeRequest(r, h.client, http.MethodPost, url, body)
		h.log.Debug("Response body", zap.ByteString("body", resp.Body))

		h.log.Info("Creating contacts successfully")
		utils.SendResponse(w, resp)
	}
}

func (h *appHandlers) CreateUserInRTU() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Creating User in RTU")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("Error reading body", zap.Error(err))
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		h.log.Debug("Body", zap.ByteString("body", body))

		h.log.Debug("Start unmarshalling body in to map")
		var jsonData utils.AnyMap
		err = json.Unmarshal(body, &jsonData)
		if err != nil {
			h.log.Error("Error unmarshalling body", zap.Error(err))
			http.Error(w, "Error unmarshalling body", http.StatusBadRequest)
			return
		}
		h.log.Debug("Unmarshalling body in to map successfully")

		reqBody := utils.AnyMap{
			"id":          jsonData["id"],
			"phoneNumber": jsonData["phoneNumber"],
			"email":       jsonData["email"],
			"name":        jsonData["name"],
			"sipLogin":    jsonData["sipLogin"],
			"sipPassword": jsonData["sipPassword"],
		}
		h.log.Debug("Marshalling newBody")
		newBody, err := json.Marshal(reqBody)
		if err != nil {
			h.log.Error("Error marshalling newBody", zap.Error(err))
			http.Error(w, "Error marshalling newBody", http.StatusBadRequest)
			return
		}
		h.log.Debug("New body for RTU", zap.ByteString("body", newBody))

		h.log.Debug("Send request to RTU")
		url := fmt.Sprintf("https://%s/user", h.rtuAddr)
		resp := utils.MakeRequest(r, h.client, http.MethodPost, url, newBody)
		h.log.Debug("Response body", zap.ByteString("body", resp.Body))

		h.log.Info("Creating user in RTU successfully")
		utils.SendResponse(w, resp)
	}
}

func (h *appHandlers) GetUserFromRTU() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Debug("Getting User from RTU")

		userID := r.URL.Query().Get("user_id")
		if userID != "" {
			h.log.Debug("Send request to RTU")
			url := fmt.Sprintf("https://%s/user?user_id=%s", h.rtuAddr, userID)
			resp := utils.MakeRequest(r, h.insecureClient, http.MethodGet, url, nil)
			h.log.Debug("Response body", zap.ByteString("body", resp.Body))
			utils.SendResponse(w, resp)
			return
		}

		h.log.Debug("Send request to RTU")
		url := fmt.Sprintf("https://%s/user", h.rtuAddr)
		resp := utils.MakeRequest(r, h.insecureClient, http.MethodGet, url, nil)
		h.log.Debug("Response body", zap.ByteString("body", resp.Body))
		utils.SendResponse(w, resp)
	}
}

func (h *appHandlers) LinkUnsorted() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Request to bind an unsorted call")
		entityID := chi.URLParam(r, "entity_id")
		h.log.Debug("EntityID: ", zap.String("entityID", entityID))

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("Error reading body", zap.Error(err))
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		h.log.Debug("Body", zap.ByteString("body", body))

		h.log.Debug("Start unmarshalling body in to map")
		var jsonData []utils.AnyMap
		err = json.Unmarshal(body, &jsonData)
		if err != nil {
			h.log.Error("Error unmarshalling body", zap.Error(err))
			http.Error(w, "Error unmarshalling body", http.StatusBadRequest)
			return
		}
		h.log.Debug("Unmarshalling body in to map successfully")

		reqBody := utils.AnyMap{
			"link": utils.AnyMap{
				"entity_id":   jsonData[0]["to_entity_id"],
				"entity_type": "leads",
			},
		}

		h.log.Debug("Marshalling newBody")
		bytesBody, err := json.Marshal(reqBody)
		if err != nil {
			h.log.Error("Error marshalling newBody", zap.Error(err))
			http.Error(w, "Error marshalling newBody", http.StatusBadRequest)
			return
		}
		h.log.Debug("New body: ", zap.ByteString("body", bytesBody))

		h.log.Debug("Send request to CRM")
		amoHost := r.Context().Value("crm_address").(string)
		url := fmt.Sprintf("https://%s/api/v4/leads/unsorted/%s/link", amoHost, entityID)
		resp := utils.MakeRequest(r, h.client, http.MethodPost, url, bytesBody)
		h.log.Debug("Response body", zap.ByteString("body", resp.Body))

		h.log.Info("Request to bind an unsorted call successfully")
		utils.SendResponse(w, resp)
	}
}

func (h *appHandlers) AddUnsorted() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Request to create a record in the unsorted list")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("Error reading body", zap.Error(err))
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		h.log.Debug("Body", zap.ByteString("body", body))

		h.log.Debug("Start unmarshalling body in to map")
		var jsonData []utils.AnyMap
		err = json.Unmarshal(body, &jsonData)
		if err != nil {
			h.log.Error("Error unmarshalling body", zap.Error(err))
			http.Error(w, "Error unmarshalling body", http.StatusBadRequest)
			return
		}
		h.log.Debug("Unmarshalling body in to map successfully")

		jsonData[0]["metadata"].(map[string]any)["service_code"] = h.serviceCode

		h.log.Debug("Marshalling newBody")
		newBody, err := json.Marshal(jsonData)
		if err != nil {
			h.log.Error("Error marshalling newBody", zap.Error(err))
			http.Error(w, "Error marshalling newBody", http.StatusBadRequest)
			return
		}
		h.log.Debug("New body: ", zap.ByteString("body", newBody))

		url := fmt.Sprintf("https://%s", trimPath(r.URL.Path))
		resp := utils.MakeRequest(r, h.client, http.MethodPost, url, newBody)

		h.log.Info("Request to create a record in the unsorted list successfully")
		utils.SendResponse(w, resp)

	}
}

func (h *appHandlers) CreateCallEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Request to create a call event")
		phoneNumber := r.FormValue("add[0][phone_number]")
		eventType := r.FormValue("add[0][type]")
		users := r.FormValue("add[0][users][0]")

		h.log.Debug("Call Event Type: " + eventType)
		h.log.Debug("Call Users: " + users)
		h.log.Debug("Call Phone: " + phoneNumber)

		body := utils.AnyMap{
			"add": []utils.AnyMap{
				{
					"type":         eventType,
					"users":        users,
					"phone_number": phoneNumber,
				},
			},
		}

		h.log.Debug("Marshalling newBody")
		jsonData, err := json.Marshal(body)
		if err != nil {
			h.log.Error("Error marshalling newBody", zap.Error(err))
			http.Error(w, "Error marshalling newBody", http.StatusBadRequest)
			return
		}
		h.log.Debug("New body: ", zap.ByteString("body", jsonData))

		h.log.Debug("Send request to CRM")
		url := fmt.Sprintf("https://%s", trimPath(r.URL.Path))
		resp := utils.MakeRequest(r, h.client, http.MethodPost, url, jsonData)
		h.log.Debug("Response body", zap.ByteString("body", resp.Body))

		h.log.Info("Request to create a call event successfully")
		utils.SendResponse(w, resp)
	}
}

func (h *appHandlers) GetCompanies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Request to get companies from crm")

		h.log.Debug("Send request to CRM")
		url := fmt.Sprintf("https://%s?limit=%s&page=%s", trimPath(r.URL.Path), r.URL.Query().Get("limit"), r.URL.Query().Get("page"))
		resp := utils.MakeRequest(r, h.client, http.MethodGet, url, nil)
		h.log.Debug("Response body", zap.ByteString("body", resp.Body))

		h.log.Info("Request to get companies from crm successfully")
		utils.SendResponse(w, resp)
	}
}

func (h *appHandlers) GetContacts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Request to get contacts from crm")

		h.log.Debug("Send request to CRM")
		url := fmt.Sprintf("https://%s?limit=%s&page=%s", trimPath(r.URL.Path), r.URL.Query().Get("limit"), r.URL.Query().Get("page"))
		resp := utils.MakeRequest(r, h.client, http.MethodGet, url, nil)
		h.log.Debug("Response body", zap.ByteString("body", resp.Body))

		h.log.Info("Request to get contacts from crm successfully")
		utils.SendResponse(w, resp)
	}
}

func (h *appHandlers) GetToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Request to get token from crm")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("Error reading body", zap.Error(err))
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		h.log.Debug("Body", zap.ByteString("body", body))

		newBody := strings.Split(string(body), "&")
		temp := make(map[string]string)

		for _, value := range newBody {
			items := strings.Split(value, "=")
			if len(items) != 2 {
				continue
			}
			temp[items[0]] = items[1]
		}

		temp["redirect_uri"] = fmt.Sprintf("https://%s/crmproxy/callback/amo", h.serverHost)

		h.log.Debug("Marshalling newBody")
		jsonData, err := json.Marshal(temp)
		if err != nil {
			h.log.Error("Error marshalling newBody", zap.Error(err))
			http.Error(w, "Error marshalling newBody", http.StatusBadRequest)
			return
		}
		h.log.Debug("New body: ", zap.ByteString("body", jsonData))

		h.log.Debug("Make request to get token")
		url := fmt.Sprintf("https://%s", trimPath(r.URL.Path))
		resp := utils.MakeRequest(r, h.client, http.MethodPost, url, jsonData)
		h.log.Debug("Response body", zap.ByteString("body", resp.Body))

		h.log.Info("Request to get token from crm send successfully")
		utils.SendResponse(w, resp)
	}
}

func (h *appHandlers) EndCall() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("Request to end call from crm")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("Error reading body", zap.Error(err))
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		h.log.Debug("Body", zap.ByteString("body", body))

		url := fmt.Sprintf("https://%s", trimPath(r.URL.Path))
		resp := utils.MakeRequest(r, h.client, http.MethodPost, url, body)

		h.log.Info("Request to end call from crm successfully")
		utils.SendResponse(w, resp)

	}
}

func trimPath(path string) string {
	pathSlice := strings.Split(path, "/")
	pathSlice = pathSlice[2:]
	return strings.Join(pathSlice, "/")
}

func getDomain(r *http.Request) (string, error) {
	credential := strings.TrimPrefix(r.Header.Get(utils.HeaderAuthorization), "Basic ")
	decoded, err := base64.URLEncoding.DecodeString(credential)
	if err != nil {
		return "", err
	}
	credSlice := strings.Split(string(decoded), ":")
	return credSlice[0], nil
}
