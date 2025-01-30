package limits

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter - структура для управления ограничением запросов для разных IP адресов.
type Limiter struct {
	// limiters - карта, где ключ - IP-адрес, а значение - состояние лимитера для этого IP.
	limiters map[string]*LimiterState
	// mu - мьютекс для синхронизации доступа к карте limiters.
	mu sync.Mutex
}

// LimiterState - хранит текущее состояние ограничения для конкретного IP.
type LimiterState struct {
	// lastBurstTime - время последнего всплеска запросов.
	lastBurstTime time.Time
	// requestsCount - количество обработанных запросов с момента последнего всплеска.
	requestsCount int
	// limiter - указатель на rate.Limiter, который управляет лимитом скорости.
	limiter *rate.Limiter
}

// NewLimiter создает новый Limiter с заданным максимальным количеством запросов в минуту.
func NewLimiter(maxRequestsPerMinute int) *Limiter {
	return &Limiter{
		// Инициализируем карту для хранения состояний лимитеров.
		limiters: make(map[string]*LimiterState),
	}
}

// GetLimiter возвращает или создает лимитер для заданного IP и проверяет, можно ли сделать запрос.
func (l *Limiter) GetLimiter(ip string, maxRequestsPerMinute int) (*rate.Limiter, bool) {
	// Блокировка мьютекса для безопасного доступа к мапе.
	l.mu.Lock()
	defer l.mu.Unlock()

	// Текущее время для проверки временных интервалов.
	now := time.Now()
	state, exists := l.limiters[ip]
	if !exists {
		// Если для IP еще нет состояния лимитера, создаем новое.
		state = &LimiterState{
			lastBurstTime: now,
			requestsCount: 0,
			limiter:       rate.NewLimiter(rate.Limit(float64(maxRequestsPerMinute)/60), maxRequestsPerMinute),
		}
		l.limiters[ip] = state
	} else {
		// Проверка, прошла ли минута с последнего всплеска запросов.
		if now.Sub(state.lastBurstTime) >= time.Minute {
			state.lastBurstTime = now
			state.requestsCount = 0
		}

		// Проверка, не превышено ли максимальное количество запросов.
		if state.requestsCount >= maxRequestsPerMinute {
			// Возвращаем nil или что-то, что сигнализирует о блокировке
			return nil, false
		}
	}

	// Увеличиваем счетчик запросов.
	state.requestsCount++
	return state.limiter, true
}

// RateLimitMiddleware оборачивает HTTP обработчик, добавляя ограничение скорости.
func RateLimitMiddleware(next http.Handler, limiter *Limiter, maxRequestsPerMinute int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем IP-адрес из RemoteAddr, игнорируя порт.
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		// Получаем лимитер для текущего IP и проверяем разрешен ли запрос.
		limiter, allowed := limiter.GetLimiter(ip, maxRequestsPerMinute)

		// Если запрос не разрешен или лимитер не позволяет сделать запрос, возвращаем ошибку.
		if !allowed || (limiter != nil && !limiter.Allow()) {
			http.Error(w, "Too many requests, limits 300 per minute", http.StatusTooManyRequests)
			return
		}

		// Если запрос разрешен, передаем его дальше.
		next.ServeHTTP(w, r)
	})
}
