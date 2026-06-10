// Питання та відповіді:
//
// 1. 4xx errors are not logged - що за 4хх, 5хх помилки? що це таке?
//    ВІДПОВІДЬ: HTTP status codes:
//    - 4xx (400-499) = помилки клієнта: 400 Bad Request (невалідний JSON), 401 Unauthorized,
//      403 Forbidden, 404 Not Found, 429 Too Many Requests.
//      Це очікувана поведінка — клієнт зробив щось неправильно, НЕ баг сервера. Не логуємо.
//    - 5xx (500-599) = помилки сервера: 500 Internal Server Error — НАШ код зламався.
//      Логуємо ЗАВЖДИ (default case у switch → h.jsonLogger.Error).
//    Поточний код правильний: 4xx handlers тільки пишуть відповідь, default (5xx) логує err.
//
// 2. Відсутні machine-readable error codes у відповідях. - це тіпа просто такий envelope
//    зробити Envelope{"error": err.Error(), code: "email_not_verified"}?
//    ВІДПОВІДЬ: Майже. Але НЕ err.Error() — він може розкрити internal деталі. Використовуй
//    статичний рядок. Два окремих ключі у JSON: {"error": "human msg", "code": "machine_code"}.
//    Вже реалізовано нижче у handleErrorLoginResponse. ⚠ Коди були переплутані — пофіксовано.
//
// 3. r.RemoteAddr за reverse proxy = IP проксі. То тут пропонуєш щоб NewRouteRateLimiter
//    записував ip в context? чи якийсь інший middleware?
//    ВІДПОВІДЬ: SetIPAddress middleware вже існує (middleware.go:167) і вже підключений у
//    loginChain (router.go:54). Він читає X-Real-IP / X-Forwarded-For тільки від trusted proxy.
//    NewRouteRateLimiter вже використовує appcontext.IPAddressFromContext(r.Context()).
//    Тут (рядок з req.IPAddress) теж читаємо з контексту через appcontext.IPAddressFromContext.
//
// 4. Відсутній request_id / trace_id у відповіді та логах. - це тоже в контекст записувати?
//    ВІДПОВІДЬ: ТАК — і SetRequestID middleware вже існує і підключений (middleware.go:178,
//    router.go:54). Він генерує UUID, пише у X-Request-ID header і в context.
//    Щоб додати до 5xx лога: h.jsonLogger.Error(err, map[string]any{
//      "request_id": appcontext.RequestIDFromContext(ctx),
//    })

package authhttp

import (
	"errors"
	"fmt"
	authdomain "learnflow_backend/internal/auth/domain"
	"learnflow_backend/internal/infrastructure/helpers"
	appcontext "learnflow_backend/internal/shared/context"
	"net/http"
)

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req authdomain.LoginRequest
	if err := helpers.ReadJSON(w, r, &req); err != nil {
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
		return
	}
	ctx := r.Context()

	ua := r.UserAgent()

	req.IPAddress = fmt.Sprintf("%s", r.Context().Value(appcontext.IPAddressContextKey))
	req.UserAgent = ua

	err := req.Validate()
	if err != nil {
		if respErr := helpers.BadRequestResponse(w, err); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
		return
	}

	tokens, err := h.svc.Login(ctx, req)
	if err != nil {
		h.handleErrorLoginResponse(w, err)
		return
	}

	err = helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"tokens": tokens}, nil)
	if err != nil {
		h.jsonLogger.Error(err, nil)
	}
}

func (h *Handler) handleErrorLoginResponse(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authdomain.ErrInvalidCredentials):
		if respErr := helpers.InvalidCredentialsResponse(w); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	case errors.Is(err, authdomain.ErrAccountLocked):
		w.Header().Set("Retry-After", "900")
		if respErr := helpers.ErrorResponse(w, http.StatusTooManyRequests, "account temporarily locked"); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	case errors.Is(err, authdomain.ErrAccountBlocked):
		if respErr := helpers.ForbiddenResponse(w, helpers.Envelope{"error": "account is blocked", "code": "account_blocked"}); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	case errors.Is(err, authdomain.ErrEmailNotVerified):
		if respErr := helpers.ForbiddenResponse(w, helpers.Envelope{"error": "email not verified", "code": "email_not_verified"}); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	default:
		h.jsonLogger.Error(err, nil)
		if respErr := helpers.ServerErrorResponse(w, err); respErr != nil {
			h.jsonLogger.Error(err, nil)
		}
	}
}
